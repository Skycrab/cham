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
	Code   int         `json:"code"` // 0 correct
	Result interface{} `json:"result"`
}

//
//
//concrete protocol

type User struct {
	Openid string `json:"openid"`
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

func Encode(code int, result interface{}) []byte {
	response := Response{code, result}
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
	protos := []interface{}{&User{}}
	for _, p := range protos {
		register(p)
	}

}
