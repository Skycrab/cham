package cham

import (
	"fmt"
	// "runtime"
	// "sync"
	"testing"
	"time"
)

func helloStart(service *Service, args ...interface{}) Dispatch {
	return func(session int32, source Address, ptype uint8, args ...interface{}) []interface{} {
		fmt.Println(session, source, args)
		time.Sleep(time.Second * 4)
		cmd := args[0].(string)
		if cmd == "Hello" {
			return Ret("World")
		} else if cmd == "Notify" {
			fmt.Println("no reply")
			return Ret(nil)
		} else {
			return Ret("Error")
		}
	}
}

func init() {
	// runtime.GOMAXPROCS(4)
}

func worldStart(service *Service, args ...interface{}) Dispatch {
	other := args[0].(string)
	fmt.Println("worldStart:", other)
	return func(session int32, source Address, ptypt uint8, args ...interface{}) []interface{} {
		return Ret("999")
	}
}

func TestService(t *testing.T) {
	hello := NewService("Hello", helloStart, 4)               // 4 worker of goroutine
	world := NewService("World", worldStart, 1, "other args") // 1 worker of goroutine, "other args" will pass to worldStart
	for i := 0; i < 5; i++ {
		// world.Call("Hello", "Hello")
		go func() {
			fmt.Println(world.Call(hello, PTYPE_GO, "Hello")) // worker < 5 :  there is a request will wait
		}()
		// world.Send(hello, PTYPE_GO, "send")
	}
	time.Sleep(time.Second * 100)
}
