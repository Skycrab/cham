package service

import (
	"cham/cham"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"testing"
	"time"
)

func watchDogStart(service *cham.Service, args ...interface{}) cham.Dispatch {
	return func(session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {
		return cham.NORET
	}
}

func clientDispatch(service *cham.Service, args ...interface{}) cham.Dispatch {
	return func(session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {
		sessionid := args[0].(uint32)
		data := string(args[1].([]byte))
		time.Sleep(time.Second * 5)
		if data == "hello" {
			service.Notify("gate", cham.PTYPE_RESPONSE, sessionid, []byte("world"))
		}
		// go func() {
		// 	time.Sleep(time.Second * 2)
		// 	fmt.Println("kick")
		// 	service.Notify("gate", cham.PTYPE_GO, GATE_KICK, sessionid)
		// }()
		return cham.NORET
	}
}

func runClient(n int) {
	conn, err := net.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		fmt.Println("client error:" + err.Error())
		return
	}
	i := string(strconv.Itoa(n))
	fmt.Println("client start " + i)
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
		fmt.Println("client get:"+i, string(result))
		time.Sleep(time.Second * 15)
		// break
	}
	fmt.Println("client end" + i)

}

func TestGateService(t *testing.T) {
	ws := cham.NewService("watchdog", watchDogStart, 16) // 16 worker to process client data
	ws.RegisterProtocol(cham.PTYPE_CLIENT, clientDispatch)
	gs := cham.NewService("gate", GateStart, 16) // 16 worker to send data to client
	ws.Call(gs, cham.PTYPE_GO, GATE_OPEN, NewConf("127.0.0.1:9999", 100))
	for i := 0; i < 20; i++ {
		go runClient(i)
	}
	time.Sleep(time.Minute * 2)

}
