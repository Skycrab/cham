package debug

import (
	"bufio"
	"cham/cham"
	"cham/service/log"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
)

type buffer struct {
	conn net.Conn
}

func (s buffer) WriteString(str string) {
	s.conn.Write([]byte(str))
}

type debug struct {
	addr  string
	funcs map[string]reflect.Value
}

func (d *debug) serve(conn net.Conn) {
	defer conn.Close()
	buf := buffer{conn}
	br := bufio.NewReader(conn)
	buf.WriteString("Welcome to cham console\r\n")

	for {
		line, _, err := br.ReadLine()
		if err != nil {
			return
		}
		if len(line) < 1 {
			continue
		}
		cmds := strings.Fields(string(line))
		result := d.handle(cmds)
		buf.WriteString(result)
		buf.WriteString("\r\n")
	}
}

func (d *debug) handle(cmds []string) string {
	defer func() {
		if err := recover(); err != nil {
			log.Infoln("debug cmd", cmds, " error:", err)
		}
	}()

	funName := strings.Title(cmds[0])
	if f, ok := d.funcs[funName]; !ok {
		return "Invalid command, type help for command list"
	} else {
		var in []reflect.Value
		n := len(cmds) - 1
		if n > 0 {
			args := cmds[1:]
			in = make([]reflect.Value, n)
			for i := 0; i < n; i++ {
				in[i] = reflect.ValueOf(args[i])
			}
		} else {
			in = []reflect.Value{}
		}
		result := f.Call(in)
		return result[0].String()
	}
}

func (d *debug) getfuncs() {
	v := reflect.ValueOf(d)
	t := v.Type()
	for i := 0; i < t.NumMethod(); i++ {
		name := t.Method(i).Name
		if strings.HasPrefix(name, "Cmd") {
			d.funcs[name[3:]] = v.Method(i)
		}
	}
}

func (d *debug) start() {
	listener, err := net.Listen("tcp", d.addr)
	if err != nil {
		panic("debug listen addr:" + d.addr + " ,error:" + err.Error())
	}
	log.Infoln("debug start, listen ", d.addr)
	defer listener.Close()
	d.getfuncs()
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go d.serve(conn)
	}
}

func (d *debug) CmdList() string {
	return cham.DumpService()
}

func (d *debug) CmdHelp() string {
	return `
    <-------------------->
    help  help message
    list  list all service
    send  send message to service(send addr message)
    <-------------------->

    `
}

func (d *debug) CmdSend(args ...string) string {
	if len(args) < 2 {
		return "FAILED, args not enough"
	}
	vv := make([]interface{}, len(args)-1)
	addr, err := strconv.Atoi(args[0])
	if err != nil {
		return "FAILED, addr error"
	}
	pt, err := strconv.Atoi(args[1])
	if err != nil {
		return "FAILED, ptype error"
	}
	vv[0] = uint8(pt)
	for i := 1; i < len(vv); i++ {
		vv[i] = args[i+2]
	}

	msg := cham.Main.Send(cham.Address(addr), cham.PTYPE_GO, vv...)
	return "SUCCESS," + fmt.Sprint(msg)
}

func Start(service *cham.Service, args ...interface{}) cham.Dispatch {
	d := &debug{args[0].(string), make(map[string]reflect.Value)}
	go d.start()
	return func(session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {
		return cham.NORET
	}
}
