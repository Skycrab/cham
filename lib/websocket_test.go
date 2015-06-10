package websocket

import (
	// "encoding/binary"
	"fmt"
	"net/http"
	"testing"
	"time"
)

type MyHandler struct {
}

func (wd MyHandler) CheckOrigin(origin, host string) bool {
	return true
}

func (wd MyHandler) OnOpen(ws *Websocket) {
	fmt.Println("OnOpen")
}

func (wd MyHandler) OnMessage(ws *Websocket, message []byte) {
	fmt.Println("OnMessage:", string(message), len(message))
	// 不知道为啥用的是小端
	// v := [10]uint32{}
	// j := 0
	// for i := 0; i < len(message); i = i + 4 {
	// 	v[j] = binary.LittleEndian.Uint32(message[i:])
	// 	j++
	// }
	// fmt.Println(v)
}

func (wd MyHandler) OnClose(ws *Websocket, code uint16, reason []byte) {
	fmt.Println("OnClose", code, string(reason))
}

func (wd MyHandler) OnPong(ws *Websocket, data []byte) {
	fmt.Println("OnPong:", string(data))

}

func TestWebsocket(t *testing.T) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("...")
		var opt = Option{MyHandler{}, false}
		ws, err := New(w, r, &opt)
		if err != nil {
			t.Fatal(err.Error())
		}
		time.AfterFunc(time.Second*5, func() {
			ws.Close(0, []byte("kick"))
		})
		ws.Start()
	})
	fmt.Println("server start")
	http.ListenAndServe(":8001", nil)
}
