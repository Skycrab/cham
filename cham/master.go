package cham

import (
	// "runtime"
	"strings"
	"sync"
)

const (
	DEFAULT_SERVICE_SIZE = 64
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

func (m *Master) Register(s *Service) bool {
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

func (m *Master) Unregister(s *Service) bool {
	m.Lock()
	defer m.Unlock()
	delete(m.services, s.Addr)
	nss := m.nameservices[s.Name]
	var idx int = -1
	for i, ns := range nss {
		if ns.Name == s.Name {
			idx = i
			break
		}
	}
	if idx == -1 {
		return false
	} else {
		m.nameservices[s.Name] = append(nss[:idx], nss[idx+1:]...)
		return true
	}
}

func (m *Master) GetService(query interface{}) *Service {
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

func (m *Master) UniqueService(name string) *Service {
	ns := m.nameservices[name]
	if len(ns) > 1 {
		panic("unique service duplicate")
	} else if len(ns) == 1 {
		return ns[0]
	} else {
		return nil
	}
}

func (m *Master) AllService() map[Address]*Service {
	m.RLock()
	defer m.RUnlock()
	return master.services
}

func DumpService() string {
	services := master.AllService()
	info := make([]string, 0, len(services))
	for _, s := range services {
		info = append(info, s.String())
	}
	return strings.Join(info, "\r\n")
}

func init() {
	master = NewMaster()
}
