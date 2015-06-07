package service

import (
	"cham/cham"
	"sync"
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

var (
	mul *multicast
)

type multicast struct {
	channel uint32
	groups  map[uint32]map[cham.Address]cham.NULL // channel->set(address)
}

func (m *multicast) new(addr cham.Address) uint32 {
	ch := atomic.AddInt32(&m.channel, 1)
	m.groups[ch] = make(map[cham.Address]cham.NULL, DEFAULT_CHANNEL_SIZE)
	return ch
}

func (m *multicast) sub(addr cham.Address, ch uint32) {
	if _, ok := m.groups[ch]; !ok {
		panic("should new a channel before sub a channel")
	}
	m.groups[ch][addr] = cham.NULLVALUE
}

func (m *multicast) unsub(addr cham.Address, ch uint32) {
	if _, ok := m.groups[ch]; !ok {
		panic("should new a channel before unsub a channel")
	}
	delete(m.groups[ch], addr)
}

// multi invode harmless
func (m *multicast) del(ch uint32) {
	delete(m.groups, ch)
}

func (m *multicast) pub(addr cham.Address, ch uint32, args ...interface{}) {
	peers := m.groups[ch]
	if len(peers) > 0 {
		data := make([]interface{}, 0, len(args)+1)
		data = append(data, ch, args...)
		msg := &cham.Msg{addr, 0, PTYPE_MULTICAST, data}
		for peer := range peers {
			cham.Redirect(peer, msg)
		}
	}
}

func MulticastDispatch(session int32, source Address, ptype uint8, args ...interface{}) []interface{} {
	cmd := args[0].(uint8)
	channel := args[1].(uint32)
	addr := args[2].(cham.Address)

	result := cham.NORET

	switch cmd {
	case MULTICAST_NEW:
		result = mul.new(addr)
	case MULTICAST_SUB:
		mul.sub(addr, channel)
	case MULTICAST_PUB:
		mul.pub(addr, channel, args[1:])
	case MULTICAST_UNSUB:
		mul.unsub(addr, channel)
	case MULTICAST_DEL:
		mul.del(channel)
	}

	return cham.Ret(result)
}

func init() {
	mul = new(multicast)
	mul.channel = 0
	mul.groups = make(map[uint32]map[cham.Address]cham.NULL, DEFAULT_GROUP_SIZE)
}
