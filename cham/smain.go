package cham

func mainStart(service *Service, args ...interface{}) Dispatch {
	return func(session int32, source Address, ptypt uint8, args ...interface{}) []interface{} {
		return NORET
	}
}

var Main *Service

func init() {
	Main = NewService("main", mainStart)
}
