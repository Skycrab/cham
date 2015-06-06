package cham

import (
	// "fmt"
	"sync/atomic"
)

const (
	DEFAULT_QUEUE_SIZE = 1024
)

type handler func(session int32, source Address, args ...interface{}) []interface{}

type Msg struct {
	source  Address
	session int32
	args    []interface{}
	done    chan *Msg
}

type Service struct {
	session  int32
	Name     string
	Addr     Address
	queue    chan *Msg
	closed   bool
	quit     chan struct{}
	dispatch handler
}

func Ret(args ...interface{}) []interface{} {
	return args
}

func NewService(name string, size uint32, dispatch handler) *Service {
	service := new(Service)
	service.session = 0
	service.Name = name
	service.Addr = GenAddr()
	if size == 0 {
		size = DEFAULT_QUEUE_SIZE
	}
	service.queue = make(chan *Msg, size)
	service.closed = false
	service.quit = make(chan struct{})
	service.dispatch = dispatch

	Register(service)
	go service.Start()
	return service
}

func (s *Service) Start() {
	for {
		select {
		case msg := <-s.queue:
			go s.dispatchMsg(msg)
		case <-s.quit:
			return
		}
	}
}

func (s *Service) dispatchMsg(msg *Msg) {
	if msg.session == 0 {
		s.dispatch(msg.session, msg.source, msg.args)
	} else if msg.session > 0 {
		result := s.dispatch(msg.session, msg.source, msg.args...)
		resp := &Msg{s.Addr, -msg.session, result, msg.done}
		dest := msg.source.GetService()
		dest.Push(resp)
	} else {
		msg.done <- msg
	}
}

func (s *Service) Call(query interface{}, args ...interface{}) []interface{} {
	session := atomic.AddInt32(&s.session, 1)
	msg := &Msg{s.Addr, session, args, make(chan *Msg, 1)}
	dest := GetService(query)
	dest.Push(msg)
	m := <-msg.done
	return m.args
}

// no reply
func (s *Service) Notify(addr Address, args ...interface{}) {
	// msg := &Msg{s.Addr, 0, args}
	// dest := addr.GetService()
	// dest.Push(msg)
}

func (s *Service) Push(msg *Msg) {
	s.queue <- msg
}

func (s *Service) Stop() bool {
	if !s.closed {
		s.closed = true
		close(s.quit)
		return true
	}
	return false
}

func (s *Service) Status() int {
	return len(s.queue)
}
