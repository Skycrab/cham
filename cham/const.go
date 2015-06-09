package cham

type NULL struct{}

const (
	PTYPE_GO        uint8 = iota //service -> service
	PTYPE_MULTICAST              //multicast -> service
	PTYPE_CLIENT                 //client -> gate
	PTYPE_RESPONSE               //gate -> client
)

var (
	NULLVALUE = NULL{}
	NORET     []interface{}
)
