package service

import (
	"cham/cham"
)

var multicast *cham.Service

type Channel struct {
	service *cham.Service
	channel uint32
}

//Channel base on a service
func NewChannel(service *cham.Service, channel uint32, dispatch cham.Handler) *Channel {
	if channel == 0 {
		channel = multicast.Call(MUL_NEW, 0, service.Addr)[0].(uint32)
	}
	service.RegisterProtocol(cham.PTYPE_MULTICAST, dispatch)
	c := &Channel{service, channel}
	return c
}

func (c *Channel) Publish(args ...interface{}) {
	c.service.Call(multicast, MULTICAST_PUB, c.channel, c.service.Addr, args...)
}

func (c *Channel) Subscribe() {
	c.service.Call(multicast, MULTICAST_SUB, c.channel, c.service.Addr)
}

func (c *Channel) Unsubscribe() {
	c.service.Notify(multicast, MULTICAST_UNSUB, c.channel, c.service.Addr)
}

func (c *Channel) Delete() {
	c.service.Notify(multicast, MULTICAST_DEL, c.channel, 0)
}

func init() {
	multicast = cham.UniqueService("multicast", MulticastDispatch)
}
