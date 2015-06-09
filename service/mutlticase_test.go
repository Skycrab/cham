package service

import (
	"cham/cham"
	"fmt"
	"testing"
)

func MainDispatch(service *cham.Service, session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {
	fmt.Println(service)
	fmt.Println(args)
	return cham.NORET
}

func ChatDispatch(service *cham.Service, session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {
	fmt.Println(service)
	fmt.Println(args)
	return cham.NORET
}

// args[0] is channel id
func Chat2Dispatch(service *cham.Service, session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {
	fmt.Println(service)
	fmt.Println(args)
	return cham.NORET
}

func ChannelDispatch(service *cham.Service, session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {
	fmt.Println(service)
	fmt.Println(args)
	return cham.NORET
}

func TestMulticast(t *testing.T) {
	main := cham.NewService("Leader", MainDispatch)
	chat1 := cham.NewService("chat1", ChatDispatch)
	chat2 := cham.NewService("chat2", Chat2Dispatch)
	channel := NewChannel(main, 0, ChannelDispatch)
	chat1Channel := NewChannel(chat1, channel.Channel, ChannelDispatch)
	chat2Channel := NewChannel(chat2, channel.Channel, ChannelDispatch)
	chat1Channel.Subscribe()
	chat2Channel.Subscribe()
	channel.Publish("hello world")
	chat1Channel.Publish("i am chat1")
	chat2Channel.Unsubscribe()
	chat2.Stop() // test stop services
	channel.Publish("last")
	channel.Delete()

}
