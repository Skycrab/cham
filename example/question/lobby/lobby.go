package lobby

import (
	"cham/cham"
	"cham/service/log"
	"fmt"
	"question/room"
	"sync"
)

type Lobby struct {
	roomManager room.RoomManager
	rooms       map[room.Roomid]*cham.Service
}

func newLobby() *Lobby {
	lobby := &Lobby{
		roomManager: room.NewRoomManager(),
		rooms:       make(map[room.Roomid]*cham.Service),
	}
	return lobby
}

func (lobby *Lobby) Allocate() {
	roomId := lobby.roomManager.NextRoom()
	roomName := fmt.Sprintf("room-%d", int(roomId))
	lobby.rooms[roomId] = cham.NewService(roomName, room.Start, 1, roomId)
}

func Start(service *cham.Service, args ...interface{}) cham.Dispatch {
	log.Infoln("New Service ", lobby.String())
	lobby := newLobby()
	return func(session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {

	}
}
