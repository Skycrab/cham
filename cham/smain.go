package cham

import (
	"time"
)

func mainStart(service *Service, args ...interface{}) Dispatch {
	return func(session int32, source Address, ptype uint8, args ...interface{}) []interface{} {
		return NORET
	}
}

var Main *Service

var DTimer *Timer

func init() {
	Main = NewService("main", mainStart)
	DTimer = NewWheelTimer(time.Millisecond * 10)
	go func() { DTimer.Start() }()
}
