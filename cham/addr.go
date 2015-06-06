package cham

import (
	"sync/atomic"
)

type Address uint32

var START_ADDR uint32 = 1

func GenAddr() Address {
	return Address(atomic.AddUint32(&START_ADDR, 1))
}

func (addr Address) GetService() *Service {
	return GetService(addr)
}
