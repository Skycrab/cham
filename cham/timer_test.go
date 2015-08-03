package cham

import (
	"fmt"
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	timer := NewWheelTimer(time.Millisecond * 10)
	t1 := timer.NewTimer(time.Second * 10)
	t2 := timer.NewTicker(time.Second * 2)
	go timer.Start()
	// var now time.Time
	fmt.Println(time.Now())
	for {
		select {
		case <-t1.C:
			fmt.Println("t1,", time.Now())
		case <-t2.C:
			fmt.Println("t2", time.Now())
		}
	}

}
