package room

import (
	"cham/cham"
	"cham/service/log"
	"question/model"
)

type Room struct {
	rid     Roomid
	owner   *user.User
	players []*user.User
}

func newRoom(rid Roomid) *Room {
	return &Room{
		rid:     rid,
		owner:   nil,
		players: make([]*user.User, 0, 2),
	}
}

func Start(service *cham.Service, args ...interface{}) cham.Dispatch {
	log.Infoln("New Service ", room.String())
	room := newRoom(args[0].(Roomid))
	return func(session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {

	}
}
