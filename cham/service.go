package cham

import (
	"fmt"
	"sync"
	"sync/atomic"
)

var (
	serviceMutex *sync.Mutex
)

type Handler func(service *Service, session int32, source Address, ptype uint8, args ...interface{}) []interface{}

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
	dispatchs map[uint8]Handler
}

func Ret(args ...interface{}) []interface{} {
	return args
}

func NewMsg(source Address, session int32, ptype uint8, args interface{}) *Msg {
	return &Msg{source, session, ptype, args}
}

func NewService(name string, dispatch Handler) *Service {
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
	service.dispatchs = map[uint8]Handler{PTYPE_GO: dispatch}

	master.Register(service)
	go service.Start()
	return service
}

//create or return already name
func UniqueService(name string, dispatch Handler) *Service {
	serviceMutex.Lock()
	defer serviceMutex.Unlock()

	s := master.UniqueService(name)
	if s == nil {
		s = NewService(name, dispatch)
	}
	return s
}

func (s *Service) Start() {
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
				go s.dispatchMsg(msg)
			}
		}
	}
}

func (s *Service) dispatchMsg(msg *Msg) {
	if msg.session == 0 {
		s.dispatchs[msg.ptype](s, msg.session, msg.source, msg.ptype, msg.args.([]interface{})...)
	} else if msg.session > 0 {
		result := s.dispatchs[msg.ptype](s, msg.session, msg.source, msg.ptype, msg.args.([]interface{})...)
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

func (s *Service) RegisterProtocol(ptype uint8, dispatch Handler) {
	if _, ok := s.dispatchs[ptype]; ok {
		panic(s.String() + "duplicate register protocol")
	}
	s.dispatchs[ptype] = dispatch
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
	return fmt.Sprintf("SERVICE[addr->%d, name->%s]:", s.Addr, s.Name)
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
