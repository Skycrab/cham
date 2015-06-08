package service

import (
	"cham/cham"
	// "fmt"
)

var multicast *cham.Service

type Channel struct {
	service *cham.Service
	Channel uint32
}

//Channel base on a service
func NewChannel(service *cham.Service, channel uint32, dispatch cham.Handler) *Channel {
	if channel == 0 {
		channel = service.Call(multicast, cham.PTYPE_GO, MULTICAST_NEW, uint32(0), service.Addr)[0].(uint32)
	}
	service.RegisterProtocol(cham.PTYPE_MULTICAST, dispatch)
	c := &Channel{service, channel}
	return c
}

func (c *Channel) Publish(args ...interface{}) {
	v := make([]interface{}, 0, len(args)+3)
	v = append(v, MULTICAST_PUB, c.Channel, c.service.Addr)
	v = append(v, args...)
	c.service.Call(multicast, cham.PTYPE_GO, v...)
}

func (c *Channel) Subscribe() {
	c.service.Call(multicast, cham.PTYPE_GO, MULTICAST_SUB, c.Channel, c.service.Addr)
}

func (c *Channel) Unsubscribe() {
	c.service.Notify(multicast, cham.PTYPE_GO, MULTICAST_UNSUB, c.Channel, c.service.Addr)
}

func (c *Channel) Delete() {
	c.service.Notify(multicast, cham.PTYPE_GO, MULTICAST_DEL, c.Channel, cham.Address(0))
}

func init() {
	multicast = cham.UniqueService("multicast", MulticastDispatch)
}
