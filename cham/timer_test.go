package timer

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

var sum int32 = 0
var N int32 = 300
var tt *Timer

func now() {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
	atomic.AddInt32(&sum, 1)
	v := atomic.LoadInt32(&sum)
	if v == 2*N {
		tt.Stop()
	}

}

func TestTimer(t *testing.T) {
	timer := New(time.Millisecond * 10)
	tt = timer
	fmt.Println(timer)
	var i int32
	for i = 0; i < N; i++ {
		timer.NewTimer(time.Millisecond*time.Duration(10*i), now)
		timer.NewTimer(time.Millisecond*time.Duration(10*i), now)
	}
	timer.Start()
	if sum != 2*N {
		t.Error("failed")
	}
}
