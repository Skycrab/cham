package cham

import (
	// "runtime"
	"sync"
)

const (
	DEFAULT_SERVICE_SIZE = 512
)

var master *Master

type Master struct {
	sync.RWMutex
	services     map[Address]*Service
	nameservices map[string][]*Service
}

func NewMaster() *Master {
	master := new(Master)
	master.services = make(map[Address]*Service, DEFAULT_SERVICE_SIZE)
	master.nameservices = make(map[string][]*Service, DEFAULT_SERVICE_SIZE)
	return master
}

func (m *Master) register(s *Service) bool {
	m.Lock()
	defer m.Unlock()
	m.services[s.Addr] = s
	if ns, ok := m.nameservices[s.Name]; ok {
		ns = append(ns, s)
		return false
	} else {
		ss := []*Service{s}
		m.nameservices[s.Name] = ss
		return true
	}
}

func (m *Master) getService(query interface{}) *Service {
	m.RLock()
	defer m.RUnlock()
	switch v := query.(type) {
	case Address:
		return m.services[v]
	case string:
		ss := m.nameservices[v]
		if len(ss) > 0 {
			return ss[0]
		}
		return nil
	case *Service:
		return v
	default:
		panic("not reachable")
	}

}

func Register(s *Service) bool {
	return master.register(s)
}

func GetService(query interface{}) *Service {
	return master.getService(query)
}

func init() {
	master = NewMaster()
}
