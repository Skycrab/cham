package websocket

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
)

var (
	ErrUpgrade     = errors.New("Can \"Upgrade\" only to \"WebSocket\"")
	ErrConnection  = errors.New("\"Connection\" must be \"Upgrade\"")
	ErrCrossOrigin = errors.New("Cross origin websockets not allowed")
	ErrSecVersion  = errors.New("HTTP/1.1 Upgrade Required\r\nSec-WebSocket-Version: 13\r\n\r\n")
	ErrSecKey      = errors.New("\"Sec-WebSocket-Key\" must not be  nil")
	ErrHijacker    = errors.New("Not implement http.Hijacker")
)

var (
	ErrReservedBits    = errors.New("Reserved_bits show using undefined extensions")
	ErrFrameOverload   = errors.New("Control frame payload overload")
	ErrFrameFragmented = errors.New("Control frame must not be fragmented")
	ErrInvalidOpcode   = errors.New("Invalid frame opcode")
)

var (
	crlf         = []byte("\r\n")
	challengeKey = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")
)

//referer https://github.com/Skycrab/skynet_websocket/blob/master/websocket.lua

type WsHandler interface {
	CheckOrigin(origin, host string) bool
	OnOpen(ws *Websocket)
	OnMessage(ws *Websocket, message []byte)
	OnClose(ws *Websocket, code uint16, reason []byte)
	OnPong(ws *Websocket, data []byte)
}

type WsDefaultHandler struct {
	checkOriginOr bool // 是否校验origin, default true
}

func (wd WsDefaultHandler) CheckOrigin(origin, host string) bool {
	return true
}

func (wd WsDefaultHandler) OnOpen(ws *Websocket) {
}

func (wd WsDefaultHandler) OnMessage(ws *Websocket, message []byte) {
}

func (wd WsDefaultHandler) OnClose(ws *Websocket, code uint16, reason []byte) {
}

func (wd WsDefaultHandler) OnPong(ws *Websocket, data []byte) {

}

type Websocket struct {
	conn             net.Conn
	rw               *bufio.ReadWriter
	handler          WsHandler
	clientTerminated bool
	serverTerminated bool
	maskOutgoing     bool
}

type Option struct {
	Handler      WsHandler // 处理器, default WsDefaultHandler
	MaskOutgoing bool      //发送frame是否mask, default false
}

func challengeResponse(key, protocol string) []byte {
	sha := sha1.New()
	sha.Write([]byte(key))
	sha.Write(challengeKey)
	accept := base64.StdEncoding.EncodeToString(sha.Sum(nil))
	buf := bytes.NewBufferString("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: ")
	buf.WriteString(accept)
	buf.Write(crlf)
	if protocol != "" {
		buf.WriteString("Sec-WebSocket-Protocol: ")
		buf.WriteString(protocol)
		buf.Write(crlf)
	}
	buf.Write(crlf)

	return buf.Bytes()
}

func acceptConnection(r *http.Request, h WsHandler) (challenge []byte, err error) {
	//Upgrade header should be present and should be equal to WebSocket
	if strings.ToLower(r.Header.Get("Upgrade")) != "websocket" {
		return nil, ErrUpgrade
	}

	//Connection header should be upgrade. Some proxy servers/load balancers
	// might mess with it.
	if !strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade") {
		return nil, ErrConnection
	}

	// Handle WebSocket Origin naming convention differences
	// The difference between version 8 and 13 is that in 8 the
	// client sends a "Sec-Websocket-Origin" header and in 13 it's
	// simply "Origin".
	if r.Header.Get("Sec-Websocket-Version") != "13" {
		return nil, ErrSecVersion
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = r.Header.Get("Sec-Websocket-Origin")
	}

	if origin != "" && !h.CheckOrigin(origin, r.Header.Get("Host")) {
		return nil, ErrCrossOrigin
	}

	key := r.Header.Get("Sec-Websocket-Key")
	if key == "" {
		return nil, ErrSecKey
	}

	protocol := r.Header.Get("Sec-Websocket-Protocol")
	if protocol != "" {
		idx := strings.IndexByte(protocol, ',')
		if idx != -1 {
			protocol = protocol[:idx]
		}
	}

	return challengeResponse(key, protocol), nil

}

func websocketMask(mask []byte, data []byte) {
	for i := range data {
		data[i] ^= mask[i%4]
	}
}

func New(w http.ResponseWriter, r *http.Request, opt *Option) (*Websocket, error) {

	var h WsHandler
	var maskOutgoing bool
	if opt == nil {
		h = WsDefaultHandler{true}
		maskOutgoing = false
	} else {
		h = opt.Handler
		maskOutgoing = opt.MaskOutgoing
	}

	challenge, err := acceptConnection(r, h)
	if err != nil {
		var code int
		if err == ErrCrossOrigin {
			code = 403
		} else {
			code = 400
		}
		w.WriteHeader(code)
		w.Write([]byte(err.Error()))
		return nil, err
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, ErrHijacker
	}

	conn, rw, err := hj.Hijack()

	ws := new(Websocket)
	ws.conn = conn
	ws.rw = rw
	ws.handler = h
	ws.maskOutgoing = maskOutgoing

	if _, err := ws.conn.Write(challenge); err != nil {
		ws.conn.Close()
		return nil, err
	}
	ws.handler.OnOpen(ws)
	return ws, nil
}

func (ws *Websocket) read(buf []byte) error {
	_, err := io.ReadFull(ws.rw, buf)
	return err
}

func (ws *Websocket) SendFrame(fin bool, opcode byte, data []byte) error {
	//max frame header may 14 length
	buf := make([]byte, 0, len(data)+14)
	var finBit, maskBit byte
	if fin {
		finBit = 0x80
	} else {
		finBit = 0
	}

	buf = append(buf, finBit|opcode)
	length := len(data)
	if ws.maskOutgoing {
		maskBit = 0x80
	} else {
		maskBit = 0
	}
	if length < 126 {
		buf = append(buf, byte(length)|maskBit)
	} else if length < 0xFFFF {
		buf = append(buf, 126|maskBit, 0, 0)
		binary.BigEndian.PutUint16(buf[len(buf)-2:], uint16(length))
	} else {
		buf = append(buf, 127|maskBit, 0, 0, 0, 0, 0, 0, 0, 0)
		binary.BigEndian.PutUint64(buf[len(buf)-8:], uint64(length))
	}

	if ws.maskOutgoing {

	}

	buf = append(buf, data...)
	ws.rw.Write(buf)
	return ws.rw.Flush()
}

func (ws *Websocket) SendText(data []byte) error {
	return ws.SendFrame(true, 0x1, data)
}

func (ws *Websocket) SendBinary(data []byte) error {
	return ws.SendFrame(true, 0x2, data)
}

func (ws *Websocket) SendPing(data []byte) error {
	return ws.SendFrame(true, 0x9, data)
}

func (ws *Websocket) SendPong(data []byte) error {
	return ws.SendFrame(true, 0xA, data)
}

func (ws *Websocket) Close(code uint16, reason []byte) {
	if !ws.serverTerminated {
		data := make([]byte, 0, len(reason)+2)
		if code == 0 && reason != nil {
			code = 1000
		}
		if code != 0 {
			data = append(data, 0, 0)
			binary.BigEndian.PutUint16(data, code)
		}
		if reason != nil {
			data = append(data, reason...)
		}
		ws.SendFrame(true, 0x8, data)
		ws.serverTerminated = true
	}
	if ws.clientTerminated {
		ws.conn.Close()
	}

}

func (ws *Websocket) RecvFrame() (final bool, message []byte, err error) { //text 数据报文
	buf := make([]byte, 8, 8)
	err = ws.read(buf[:2])
	if err != nil {
		return
	}
	header, payload := buf[0], buf[1]
	final = header&0x80 != 0
	reservedBits := header&0x70 != 0
	frameOpcode := header & 0xf
	frameOpcodeIsControl := frameOpcode&0x8 != 0

	if reservedBits {
		// client is using as-yet-undefined extensions
		err = ErrReservedBits
		return
	}

	maskFrame := payload&0x80 != 0
	payloadlen := uint64(payload & 0x7f)

	if frameOpcodeIsControl && payloadlen >= 126 {
		err = ErrFrameOverload
		return
	}

	if frameOpcodeIsControl && !final {
		err = ErrFrameFragmented
		return
	}

	//解析frame长度
	var frameLength uint64
	if payloadlen < 126 {
		frameLength = payloadlen
	} else if payloadlen == 126 {
		err = ws.read(buf[:2])
		if err != nil {
			return
		}
		frameLength = uint64(binary.BigEndian.Uint16(buf[:2]))

	} else { //payloadlen == 127
		err = ws.read(buf[:8])
		if err != nil {
			return
		}
		frameLength = binary.BigEndian.Uint64(buf[:8])
	}

	frameMask := make([]byte, 4, 4)
	if maskFrame {
		err = ws.read(frameMask)
		if err != nil {
			return
		}
	}

	// fmt.Println("final_frame:", final, "frame_opcode:", frameOpcode, "mask_frame:", maskFrame, "frame_length:", frameLength)

	message = make([]byte, frameLength, frameLength)
	if frameLength > 0 {
		err = ws.read(message)
		if err != nil {
			return
		}
	}

	if maskFrame && frameLength > 0 {
		websocketMask(frameMask, message)
	}

	if !final {
		return
	} else {
		switch frameOpcode {
		case 0x1: //text
		case 0x2: //binary
		case 0x8: // close
			var code uint16
			var reason []byte
			if frameLength >= 2 {
				code = binary.BigEndian.Uint16(message[:2])
			}
			if frameLength > 2 {
				reason = message[2:]
			}
			message = nil
			ws.clientTerminated = true
			ws.Close(0, nil)
			ws.handler.OnClose(ws, code, reason)
		case 0x9: //ping
			message = nil
			ws.SendPong(nil)
		case 0xA:
			ws.handler.OnPong(ws, message)
			message = nil
		default:
			err = ErrInvalidOpcode
		}
		return
	}

}

func (ws *Websocket) Recv() ([]byte, error) {
	data := make([]byte, 0, 8)
	for {
		final, message, err := ws.RecvFrame()
		if final {
			data = append(data, message...)
			break
		} else {
			data = append(data, message...)
		}
		if err != nil {
			return data, err
		}
	}
	if len(data) > 0 {
		ws.handler.OnMessage(ws, data)
	}
	return data, nil
}

func (ws *Websocket) Start() {
	for {
		_, err := ws.Recv()
		if err != nil {
			ws.conn.Close()
		}
	}

}
