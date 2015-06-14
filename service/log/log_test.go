package log

import (
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
}

func TestAllFlagLog(t *testing.T) {
	ll := New("log.txt", LALL, LDEBUG)
	ll.Infoln("hello")
	ll.Debugln("world")
}
