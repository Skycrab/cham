package log

import (
	"cham/cham"
	"io"
	"os"
)

const (
	OPEN uint8 = iota
	WRITE
)

type Loggers struct {
	outs []io.Writer
}

func Start(service *cham.Service, args ...interface{}) cham.Dispatch {
	loggers := new(Loggers)
	return func(session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {
		cmd := args[0].(uint8)
		switch cmd {
		case OPEN:
			name := args[1].(string)
			if name == "" {
				loggers.outs = append(loggers.outs, os.Stdout)
			} else {
				if f, err := os.Create(name); err != nil {
					panic("create log file:" + name + ":" + err.Error())
				} else {
					loggers.outs = append(loggers.outs, f)
				}
			}
		case WRITE:
			data := args[1].([]byte)
			for _, f := range loggers.outs {
				f.Write(data)
			}
		}
		return cham.NORET
	}
}
