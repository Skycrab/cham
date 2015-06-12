package service

import (
	"bufio"
	"cham/cham"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
)

const (
	GATE_OPEN uint8 = iota
	GATE_KICK
)

var (
	bufioReaderPool sync.Pool
	bufioWriterPool sync.Pool
)

type Conf struct {
	address   string //127.0.0.1:8000
	maxclient uint32 // 0 -> no limit
	path      string // "/ws" websocket, null is tcp
}

func NewConf(address string, maxclient uint32, path string) *Conf {
	return &Conf{address, maxclient, path}
}

type Gate struct {
	rwmutex   *sync.RWMutex
	Source    cham.Address
	dest      *cham.Service
	session   uint32
	clinetnum uint32
	maxclient uint32
	sessions  map[uint32]Backend
}

type Backend interface {
	Write(data []byte)
	Close()
}

type TcpBackend struct {
	session uint32
	conn    net.Conn
	brw     *bufio.ReadWriter
}

// tcp readbuf start
func newBufioReader(r io.Reader) *bufio.Reader {
	if v := bufioReaderPool.Get(); v != nil {
		br := v.(*bufio.Reader)
		br.Reset(r)
		return br
	}
	return bufio.NewReader(r)
}

func putBufioReader(r *bufio.Reader) {
	r.Reset(nil)
	bufioReaderPool.Put(r)
}

func newBufioWriter(w io.Writer) *bufio.Writer {
	if v := bufioWriterPool.Get(); v != nil {
		bw := v.(*bufio.Writer)
		bw.Reset(w)
		return bw
	}
	return bufio.NewWriter(w)
}

func putBufioWriter(w *bufio.Writer) {
	w.Reset(nil)
	bufioWriterPool.Put(w)
}

func newTcpBackend(session uint32, conn net.Conn) *TcpBackend {
	br := newBufioReader(conn)
	bw := newBufioWriter(conn)
	return &TcpBackend{session, conn, bufio.NewReadWriter(br, bw)}
}

// tcp readbuf end

func (t *TcpBackend) Close() {
	putBufioReader(t.brw.Reader)
	putBufioWriter(t.brw.Writer)
	t.conn.Close()
}

func (t *TcpBackend) Write(data []byte) {
	head := make([]byte, 2)
	binary.BigEndian.PutUint16(head, uint16(len(data)))
	t.brw.Write(head)
	t.brw.Write(data)
	t.brw.Flush()
}

func (t *TcpBackend) readFull(buf []byte) error {
	if _, err := io.ReadFull(t.brw, buf); err != nil {
		if e, ok := err.(net.Error); ok && !e.Temporary() {
			return err
		}
	}
	return nil
}

// bigendian 2byte length+data
func (t *TcpBackend) serve(g *Gate) {
	head := make([]byte, 2)
	for {
		if err := t.readFull(head); err != nil {
			g.kick(t.session)
			return
		}

		length := binary.BigEndian.Uint16(head)
		data := make([]byte, length, length)

		if err := t.readFull(data); err != nil {
			g.kick(t.session)
			return
		}
		msg := cham.NewMsg(0, 0, cham.PTYPE_CLIENT, cham.Ret(t.session, data))
		g.dest.Push(msg)
	}
}

type WebsocketBackend struct {
	*Websocket
}

func (w *WebsocketBackend) Close() {
	w.Websocket.Close(0, []byte("kick"))
}

func (w *WebsocketBackend) Write(data []byte) {
	w.SendText(data)
}

func newWebsocket(w http.ResponseWriter, r *http.Request, opt *Option, session uint32, gate *Gate) (*WebsocketBackend, error) {
	ws, err := NewWebsocket(w, r, opt, session, gate)
	if err != nil {
		return nil, err
	}
	return &WebsocketBackend{ws}, nil
}

//websocket start
type wsHandler struct {
}

func (wd wsHandler) CheckOrigin(origin, host string) bool {
	return true
}

func (wd wsHandler) OnOpen(ws *Websocket) {
	fmt.Println("OnOpen")
}

func (wd wsHandler) OnMessage(ws *Websocket, message []byte) {
	// fmt.Println("OnMessage:", string(message), len(message))
	msg := cham.NewMsg(0, 0, cham.PTYPE_CLIENT, cham.Ret(ws.session, message))
	ws.gate.dest.Push(msg)
}

func (wd wsHandler) OnClose(ws *Websocket, code uint16, reason []byte) {
	fmt.Println("OnClose", code, string(reason))
}

func (wd wsHandler) OnPong(ws *Websocket, data []byte) {
	fmt.Println("OnPong:", string(data))

}

//websocket end

func newGate(source cham.Address) *Gate {
	gate := new(Gate)
	gate.rwmutex = new(sync.RWMutex)
	gate.Source = source
	gate.clinetnum = 0
	gate.session = 0
	gate.sessions = make(map[uint32]Backend)
	return gate
}

func (g *Gate) nextSession() uint32 {
	return atomic.AddUint32(&g.session, 1)
}

func (g *Gate) addBackend(session uint32, b Backend) {
	g.rwmutex.Lock()
	g.sessions[session] = b
	g.rwmutex.Unlock()
}

//gate listen
func (g *Gate) open(conf *Conf) {
	maxclient := conf.maxclient
	g.maxclient = maxclient
	if conf.path == "" {
		listen, err := net.Listen("tcp", conf.address)
		if err != nil {
			panic("gate http open error:" + err.Error())
		}
		go func() {
			defer listen.Close()
			for {
				conn, err := listen.Accept()
				if err != nil {
					continue
				}
				if maxclient != 0 && g.clinetnum >= maxclient {
					conn.Close() //server close socket(!net.Error)
					break
				}
				g.clinetnum++
				session := g.nextSession()
				backend := newTcpBackend(session, conn)
				g.sessions[session] = backend // not need mutex, so not addBackend
				go backend.serve(g)
			}
		}()

	} else {
		var opt = Option{wsHandler{}, false}
		http.HandleFunc(conf.path, func(w http.ResponseWriter, r *http.Request) {

			if maxclient != 0 && g.clinetnum >= maxclient {
				return
			}
			session := g.nextSession()
			ws, err := newWebsocket(w, r, &opt, session, g)
			if err != nil {
				return
			}
			g.addBackend(session, ws)
			g.clinetnum++
			ws.Start()
		})
		go func() { http.ListenAndServe(conf.address, nil) }()
	}
}

func (g *Gate) kick(session uint32) {
	var b Backend
	var ok bool
	g.rwmutex.Lock()
	if b, ok = g.sessions[session]; ok {
		delete(g.sessions, session)
		g.clinetnum--
	}
	g.rwmutex.Unlock()
	if ok {
		b.Close()
	}
}

func (g *Gate) Write(session uint32, data []byte) {
	g.rwmutex.RLock()
	b, ok := g.sessions[session]
	g.rwmutex.RUnlock()
	if ok {
		b.Write(data)
	}
}

func GateResponseStart(service *cham.Service, args ...interface{}) cham.Dispatch {
	gate := args[0].(*Gate)
	return func(session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {
		sessionid := args[0].(uint32)
		data := args[1].([]byte)
		gate.Write(sessionid, data)
		return cham.NORET
	}
}

func GateStart(service *cham.Service, args ...interface{}) cham.Dispatch {
	gate := newGate(0)
	return func(session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {
		cmd := args[0].(uint8)
		result := cham.NORET
		switch cmd {
		case GATE_OPEN:
			gate.Source = source
			gate.dest = source.GetService()
			service.RegisterProtocol(cham.PTYPE_RESPONSE, GateResponseStart, gate)
			gate.open(args[1].(*Conf))
		case GATE_KICK:
			gate.kick(args[1].(uint32))
		}

		return result
	}
}
