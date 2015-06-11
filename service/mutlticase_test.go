package service

import (
	"cham/cham"
	"fmt"
	"testing"
)

func mainStart(service *cham.Service) cham.Dispatch {
	return func(session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {
		fmt.Println(service)
		fmt.Println(args)
		return cham.NORET
	}
}

// args[0] is channel id
func chatStart(service *cham.Service) cham.Dispatch {
	return func(session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {
		fmt.Println(service)
		fmt.Println(args)
		return cham.NORET
	}
}

func channelStart(service *cham.Service) cham.Dispatch {
	return func(session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {
		fmt.Println(args)
		return cham.NORET
	}
}

func TestMulticast(t *testing.T) {
	main := cham.NewService("Leader", mainStart)
	chat1 := cham.NewService("chat1", chatStart)
	chat2 := cham.NewService("chat2", chatStart)
	//create a channel
	channel := NewChannel(main, 0, channelStart)
	//bind a channel
	chat1Channel := NewChannel(chat1, channel.Channel, channelStart)
	chat2Channel := NewChannel(chat2, channel.Channel, channelStart)
	chat1Channel.Subscribe()
	chat2Channel.Subscribe()
	channel.Publish("hello world")
	chat1Channel.Publish("i am chat1")
	chat2Channel.Unsubscribe()
	chat2.Stop() // test stop services
	channel.Publish("last")
	channel.Delete()

}
