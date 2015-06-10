package service

import (
	"cham/cham"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

func WatchDogDispatch(service *cham.Service, session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {
	return cham.NORET
}

func ClientDispatch(service *cham.Service, session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {
	sessionid := args[0].(uint32)
	data := string(args[1].([]byte))
	if data == "hello" {
		service.Notify("gate", cham.PTYPE_RESPONSE, sessionid, []byte("world"))
	}
	go func() {
		time.Sleep(time.Second * 2)
		fmt.Println("kick")
		service.Notify("gate", cham.PTYPE_GO, GATE_KICK, sessionid)
	}()
	return cham.NORET
}

func runClient() {
	conn, err := net.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		fmt.Println("client error:" + err.Error())
		return
	}
	for {
		data := []byte("hello")
		head := make([]byte, 2)
		binary.BigEndian.PutUint16(head, uint16(len(data)))
		if _, err := conn.Write(head); err != nil {
			fmt.Println("client error:" + err.Error())
			break
		}
		conn.Write(data)
		time.Sleep(time.Second)
		io.ReadFull(conn, head)
		length := binary.BigEndian.Uint16(head)
		result := make([]byte, length)
		io.ReadFull(conn, result)
		fmt.Println("client get:", string(result))
		time.Sleep(time.Second * 3)
	}
	fmt.Println("client end")

}

func TestGateService(t *testing.T) {
	ws := cham.NewService("watchdog", WatchDogDispatch)
	ws.RegisterProtocol(cham.PTYPE_CLIENT, ClientDispatch)
	gs := cham.NewService("gate", GateDispatch)
	ws.Call(gs, cham.PTYPE_GO, GATE_OPEN, NewConf("127.0.0.1:9999", 2))
	go runClient()
	time.Sleep(time.Minute * 2)

}
