package helper

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestFork(t *testing.T) {
	fmt.Println("pid", os.Getpid())
	pid, err := fork()
	if err != 0 {
		t.Error("fork error,errno:", err)
	}
	if pid == 0 {
		fmt.Println("child:", os.Getpid(), "parent:", os.Getppid())
		time.Sleep(time.Second * 20)
		fmt.Println("child exit")
		os.Exit(0)
	} else {
		fmt.Println("parent:", os.Getpid(), "child:", pid)
	}
}
