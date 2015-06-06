package cham

import (
	// "fmt"
	"sync"
	"sync/atomic"
)

type handler func(session int32, source Address, args ...interface{}) []interface{}

type Msg struct {
	source  Address
	session int32
	args    []interface{}
}

type Service struct {
	session  int32
	Name     string
	Addr     Address
	queue    *Queue
	closed   bool
	quit     chan struct{}
	rlock    *sync.Mutex
	rcond    *sync.Cond
	pending  map[int32]chan *Msg
	dispatch handler
}

func Ret(args ...interface{}) []interface{} {
	return args
}

func NewService(name string, dispatch handler) *Service {
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
	service.dispatch = dispatch

	Register(service)
	go service.Start()
	return service
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
		s.dispatch(msg.session, msg.source, msg.args...)
	} else if msg.session > 0 {
		result := s.dispatch(msg.session, msg.source, msg.args...)
		resp := &Msg{s.Addr, -msg.session, result}
		dest := msg.source.GetService()
		dest.Push(resp)
	} else {
		session := -msg.session
		done := s.pending[session]
		delete(s.pending, session)
		done <- msg
	}
}

func (s *Service) send(query interface{}, session int32, args ...interface{}) chan *Msg {
	if session != 0 {
		session = atomic.AddInt32(&s.session, 1)
	}
	msg := &Msg{s.Addr, session, args}
	dest := GetService(query)
	dest.Push(msg)
	var done chan *Msg
	if session != 0 { // need reply
		done = make(chan *Msg, 1)
		s.pending[session] = done
	}

	return done
}

// wait response
func (s *Service) Call(query interface{}, args ...interface{}) []interface{} {
	m := <-s.send(query, 1, args...)
	return m.args
}

// no reply
func (s *Service) Notify(query interface{}, args ...interface{}) {
	s.send(query, 0, args...)
}

// no wait response
func (s *Service) Send(query interface{}, args ...interface{}) chan *Msg {
	return s.send(query, 1, args...)
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
		return true
	}
	return false
}

func (s *Service) Status() int {
	return s.queue.Length()
}
