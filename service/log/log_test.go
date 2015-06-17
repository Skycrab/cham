package log

import (
	"cham/cham"
	"testing"
)

func TestLog(t *testing.T) {
	Info("hello")
	Debug("world")
}

func TestFileLog(t *testing.T) {
	ll := New("log.txt", LDEFAULT, LDEBUG)
	ll.Infoln("hello")
	ll.Debugln("world")
	cham.Main.Call("log", cham.PTYPE_GO, FLUSH)
}

func TestAllFlagLog(t *testing.T) {
	ll := New("log.txt", LALL, LDEBUG)
	ll.Infoln("hello")
	ll.Debugln("world")
}
