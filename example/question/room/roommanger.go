package room

import (
	"sync/atomic"
)

type Roomid uint32

type RoomManager struct {
	current Roomid
}

func NewRoomManager() *RoomManager {
	return &RoomManager{0}
}

func (rm *RoomManager) NextRoom() Roomid {
	return Roomid(atomic.AddUint32(&rm.current, 1))
}
