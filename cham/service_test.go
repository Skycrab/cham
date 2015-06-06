package cham

import (
	"fmt"
	"testing"
	"time"
)

func helloDispatch(session int32, source Address, args ...interface{}) []interface{} {
	fmt.Println(session, source, args)
	cmd := args[0].(string)
	time.Sleep(time.Second)
	if cmd == "Hello" {
		return Ret("World")
	} else {
		return Ret("Error")
	}

}

func WorldDispatch(session int32, source Address, args ...interface{}) []interface{} {
	return Ret("999")
}

func TestService(t *testing.T) {
	hello := NewService("Hello", 100, helloDispatch)
	world := NewService("World", 100, WorldDispatch)
	for i := 0; i < 5; i++ {
		go fmt.Println(world.Call("Hello", "Hello"))
		fmt.Println(world.Call(hello, ""))
	}
	time.Sleep(time.Second * 100)
}
