package main

import (
	"cham/cham"
	"cham/service/gate"
	"cham/service/log"
	"fmt"
	"question/lobby"
	"question/usermanager"
)

func brokerStart(service *cham.Service, args ...interface{}) cham.Dispatch {
	log.Infoln("New Service ", service.String())
	um := usermanager.New()
	service.RegisterProtocol(cham.PTYPE_CLIENT, brokerDispatch, um)
	return func(session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {
		return cham.NORET
	}
}

func brokerDispatch(service *cham.Service, args ...interface{}) cham.Dispatch {
	um := args[0].(*usermanager.UserManager)
	return func(session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {
		sessionid := args[0].(uint32)
		gt := args[1].(uint8)
		switch gt {
		// case gate.OnOpen:
		// 	fmt.Println("OnOpen ", sessionid)
		// case gate.OnClose:
		// 	fmt.Println("OnClose ", sessionid, args[2:])
		// case gate.OnPong:
		// 	fmt.Println("OnPong ", sessionid, args[2])
		case gate.OnMessage:
			data := args[2].([]byte)
			fmt.Println("OnMessage", sessionid, string(data))
			um.Handle(sessionid, data)
		}
		return cham.NORET
	}
}

func main() {
	gs := cham.NewService("gate", gate.Start, 8)
	bs := cham.NewService("broker", brokerStart, 8)
	bs.Call(gs, cham.PTYPE_GO, gate.OPEN, gate.NewConf("127.0.0.1:9998", 0, "/ws"))
	lobby := cham.NewService("lobby", lobby.Start)
	cham.Run()
}
