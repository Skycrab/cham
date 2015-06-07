package cham

type NULL struct{}

const (
	PTYPE_GO uint8 = iota
	PTYPE_MULTICAST
)

var (
	NULLVALUE = NULL{}
	NORET     []interface{}
)
