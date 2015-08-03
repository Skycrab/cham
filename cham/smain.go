package cham

import (
	"time"
)

const (
	CHAM_STOP uint8 = iota
)

var Main *Service

var DTimer *Timer

var stop chan NULL

func mainStart(service *Service, args ...interface{}) Dispatch {
	return func(session int32, source Address, ptype uint8, args ...interface{}) []interface{} {
		switch ptype {
		case CHAM_STOP:
			stop <- NULLVALUE
		}
		return NORET
	}
}

func Run() {
	<-stop
}

func init() {
	Main = NewService("main", mainStart)
	DTimer = NewWheelTimer(time.Millisecond * 10)
	go func() { DTimer.Start() }()
}
