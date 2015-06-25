package helper

import (
	"os"
	"runtime"
	"syscall"
)

func Fork() (int, syscall.Errno) {
	r1, r2, err := syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)
	if err != 0 {
		return 0, err
	}
	if runtime.GOOS == "darwin" && r2 == 1 {
		r1 = 1
	}
	return int(r1), 0
}

// referer http://ikarishinjieva.github.io/blog/blog/2014/03/20/go-file-lock/
func LockFile(name string, truncate bool) (*os.File, error) {
	flag := os.O_RDWR | os.O_CREATE
	if truncate {
		flag |= os.O_TRUNC
	}
	f, err := os.OpenFile(name, flag, 0666)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}
