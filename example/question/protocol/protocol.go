package protocol

import (
	"cham/service/log"
	"encoding/json"
	"reflect"
	"strings"
)

var protocols = make(map[string]reflect.Type)

//request  head
type Request struct {
	Service string          `json:"service"`
	Data    json.RawMessage `json:"data"`
}

//response head
type Response struct {
	Code    int         `json:"code"` // 0 correct
	Service string      `json:"service"`
	Result  interface{} `json:"result"`
}

//
//
//concrete protocol

type Login struct {
	Openid string `json:"openid"`
}

//to common, websocket and tcp use the same
type HeartBeat struct {
}

func Decode(data []byte) (string, interface{}) {
	request := &Request{}
	err := json.Unmarshal(data, request)
	if err != nil {
		log.Errorln("Request Decode Header error:", err.Error())
		return "", nil
	}
	if p, ok := protocols[request.Service]; ok {
		np := reflect.New(p).Interface()
		err := json.Unmarshal(request.Data, np)
		if err != nil {
			log.Errorln("Request Decode ", request.Service, "error:", err.Error())
		} else {
			return request.Service, np
		}
	}
	return "", nil

}

//result can ptr or value
func Encode(code int, result interface{}) []byte {
	t := reflect.TypeOf(result)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	response := Response{code, strings.ToLower(t.Name()), result}
	b, e := json.Marshal(response)
	if e != nil {
		log.Errorln("Response Encode error, ", e.Error())
		return nil
	}
	return b
}

func register(p interface{}) {
	v := reflect.ValueOf(p)
	name := strings.ToLower(v.Elem().Type().Name())
	protocols[name] = v.Type().Elem()
}

func init() {
	protos := []interface{}{&Login{}}
	for _, p := range protos {
		register(p)
	}

}
