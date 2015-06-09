package cham

import (
	"fmt"
	// "runtime"
	// "sync"
	"testing"
	"time"
)

func helloDispatch(service *Service, session int32, source Address, ptypt uint8, args ...interface{}) []interface{} {
	fmt.Println(session, source, args)
	cmd := args[0].(string)
	time.Sleep(time.Second * 4)
	if cmd == "Hello" {
		return Ret("World")
	} else if cmd == "Notify" {
		fmt.Println("no reply")
		return Ret(nil)
	} else {
		return Ret("Error")
	}

}

func init() {
	// runtime.GOMAXPROCS(4)
}

func WorldDispatch(service *Service, session int32, source Address, ptypt uint8, args ...interface{}) []interface{} {
	return Ret("999")
}

func TestService(t *testing.T) {
	hello := NewService("Hello", helloDispatch)
	world := NewService("World", WorldDispatch)
	for i := 0; i < 5; i++ {
		// world.Call("Hello", "Hello")
		go func() {
			fmt.Println(world.Call(hello, PTYPE_GO, "Hello"))
		}()
		// world.Send(hello, PTYPE_GO, "send")
	}
	time.Sleep(time.Second * 100)
}
