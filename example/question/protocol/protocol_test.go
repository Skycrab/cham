package protocol

import (
	"fmt"
	"testing"
)

func TestProto(t *testing.T) {
	d := []byte(`{"service":"user","data":{"openid":"lwy"}}`)
	fmt.Println(Decode(d))
}
