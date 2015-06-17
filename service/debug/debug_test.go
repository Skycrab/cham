package debug

import (
	"cham/cham"
	"cham/service/log"
	"testing"
	"time"
)

func TestDebug(t *testing.T) {
	ll := log.New("log.txt", log.LDEFAULT, log.LDEBUG)
	ll.Infoln("hello")
	cham.NewService("debug", Start, 1, "127.0.0.1:8888")
	time.Sleep(time.Minute * 30)
}
