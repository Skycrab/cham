package multicast

import (
	"cham/cham"
	// "fmt"
	"sync/atomic"
)

const (
	DEFAULT_GROUP_SIZE   = 64
	DEFAULT_CHANNEL_SIZE = 16
)

const (
	MULTICAST_NEW uint8 = iota
	MULTICAST_SUB
	MULTICAST_PUB
	MULTICAST_UNSUB
	MULTICAST_DEL
)

type Multicast struct {
	channel uint32
	groups  map[uint32]map[cham.Address]cham.NULL // channel->set(address)
}

func (m *Multicast) new(addr cham.Address) uint32 {
	ch := atomic.AddUint32(&m.channel, 1)
	m.groups[ch] = make(map[cham.Address]cham.NULL, DEFAULT_CHANNEL_SIZE)
	return ch
}

func (m *Multicast) sub(addr cham.Address, ch uint32) {
	if _, ok := m.groups[ch]; !ok {
		panic("should new a channel before sub a channel")
	}
	m.groups[ch][addr] = cham.NULLVALUE
}

func (m *Multicast) unsub(addr cham.Address, ch uint32) {
	if _, ok := m.groups[ch]; !ok {
		panic("should new a channel before unsub a channel")
	}
	delete(m.groups[ch], addr)
}

// multi invode harmless
func (m *Multicast) del(ch uint32) {
	delete(m.groups, ch)
}

func (m *Multicast) pub(addr cham.Address, ch uint32, args ...interface{}) {
	peers := m.groups[ch]
	if len(peers) > 0 {
		data := make([]interface{}, 0, len(args)+1)
		data = append(data, ch)
		data = append(data, args...)
		msg := cham.NewMsg(addr, 0, cham.PTYPE_MULTICAST, data) // args[0] is channel id
		for peer := range peers {
			cham.Redirect(peer, msg)
		}
	}
}

//service self
func MulticastStart(service *cham.Service, args ...interface{}) cham.Dispatch {
	mul := new(Multicast)
	mul.channel = 0
	mul.groups = make(map[uint32]map[cham.Address]cham.NULL, DEFAULT_GROUP_SIZE)

	return func(session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {
		cmd := args[0].(uint8)
		channel := args[1].(uint32)
		addr := args[2].(cham.Address)

		result := cham.NORET

		switch cmd {
		case MULTICAST_NEW:
			result = cham.Ret(mul.new(addr))
		case MULTICAST_SUB:
			mul.sub(addr, channel)
		case MULTICAST_PUB:
			mul.pub(addr, channel, args[3:]...)
		case MULTICAST_UNSUB:
			mul.unsub(addr, channel)
		case MULTICAST_DEL:
			mul.del(channel)
		}
		return result
	}
}
