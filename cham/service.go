package cham

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
)

const (
	DEFAULT_SERVICE_WORKER int = 1
)

var (
	serviceMutex *sync.Mutex
)

type Dispatch func(session int32, source Address, ptype uint8, args ...interface{}) []interface{}
type Start func(service *Service, args ...interface{}) Dispatch

type Msg struct {
	source  Address
	session int32
	ptype   uint8
	args    interface{}
}

type Service struct {
	session   int32
	Name      string
	Addr      Address
	queue     *Queue
	closed    bool
	quit      chan struct{}
	rlock     *sync.Mutex
	rcond     *sync.Cond
	pending   map[int32]chan *Msg
	dispatchs map[uint8]Dispatch
}

func Ret(args ...interface{}) []interface{} {
	return args
}

func NewMsg(source Address, session int32, ptype uint8, args interface{}) *Msg {
	return &Msg{source, session, ptype, args}
}

//args[0] is worker number, args[1:] will pass to start
func NewService(name string, start Start, args ...interface{}) *Service {
	service := new(Service)
	service.session = 0
	service.Name = name
	service.Addr = GenAddr()
	service.queue = NewQueue()
	service.closed = false
	service.quit = make(chan struct{})
	service.rlock = new(sync.Mutex)
	service.rcond = sync.NewCond(service.rlock)
	service.pending = make(map[int32]chan *Msg)

	var n int = 1
	if len(args) > 0 {
		n = args[0].(int)
		if n <= 0 {
			n = DEFAULT_SERVICE_WORKER
		}
		args = args[1:]
	}

	service.dispatchs = map[uint8]Dispatch{PTYPE_GO: start(service, args...)}

	// start may failed, user can invoke service.Stop(), so check service.closed flag
	if service.closed {
		return nil
	}
	master.Register(service)

	for i := 0; i < n; i++ {
		go service.Start(i)
	}

	return service
}

//create or return already name
func UniqueService(name string, start Start, args ...interface{}) *Service {
	serviceMutex.Lock()
	defer serviceMutex.Unlock()

	s := master.UniqueService(name)
	if s == nil {
		s = NewService(name, start, args...)
	}
	return s
}

func (s *Service) Start(i int) {
	_ = string(strconv.Itoa(i))
	for {
		select {
		case <-s.quit:
			return
		default:
			msg := s.queue.Pop()
			if msg == nil {
				s.rlock.Lock()
				s.rcond.Wait()
				s.rlock.Unlock()
			} else {
				s.dispatchMsg(msg)
			}
		}
	}
}

func (s *Service) dispatchMsg(msg *Msg) {
	if msg.session == 0 {
		s.dispatchs[msg.ptype](msg.session, msg.source, msg.ptype, msg.args.([]interface{})...)
	} else if msg.session > 0 {
		result := s.dispatchs[msg.ptype](msg.session, msg.source, msg.ptype, msg.args.([]interface{})...)
		resp := &Msg{s.Addr, -msg.session, msg.ptype, result}
		dest := msg.source.GetService()
		dest.Push(resp)
	} else {
		session := -msg.session
		done := s.pending[session]
		delete(s.pending, session)
		done <- msg
	}
}

func (s *Service) RegisterProtocol(ptype uint8, start Start, args ...interface{}) {
	if _, ok := s.dispatchs[ptype]; ok {
		panic(s.String() + "duplicate register protocol")
	}
	s.dispatchs[ptype] = start(s, args...)
}

func (s *Service) send(query interface{}, ptype uint8, session int32, args ...interface{}) chan *Msg {
	if session != 0 {
		session = atomic.AddInt32(&s.session, 1)
	}
	msg := &Msg{s.Addr, session, ptype, args}
	dest := master.GetService(query)
	dest.Push(msg)
	var done chan *Msg
	if session != 0 { // need reply
		done = make(chan *Msg, 1)
		s.pending[session] = done
	}

	return done
}

// wait response, query can service name/addr/service
func (s *Service) Call(query interface{}, ptype uint8, args ...interface{}) []interface{} {
	m := <-s.send(query, ptype, 1, args...)
	return m.args.([]interface{})
}

// no reply
func (s *Service) Notify(query interface{}, ptype uint8, args ...interface{}) {
	s.send(query, ptype, 0, args...)
}

// no wait response
func (s *Service) Send(query interface{}, ptype uint8, args ...interface{}) chan *Msg {
	return s.send(query, ptype, 1, args...)
}

func (s *Service) Push(msg *Msg) {
	s.queue.Push(msg)
	s.rcond.Signal()
}

func (s *Service) Stop() bool {
	if !s.closed {
		s.closed = true
		close(s.quit)
		s.rcond.Signal()
		master.Unregister(s)
		return true
	}
	return false
}

func (s *Service) String() string {
	return fmt.Sprintf("SERVICE [addr->%d, name->%s] ", s.Addr, s.Name)
}

func (s *Service) Status() int {
	return s.queue.Length()
}

func Redirect(source Address, msg *Msg) {
	dest := source.GetService()
	dest.Push(msg)
}

func init() {
	serviceMutex = new(sync.Mutex)
}
