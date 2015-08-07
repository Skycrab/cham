package user

import (
	"cham/cham"
	"cham/service/log"
	"question/db"
	"question/model"
	"time"
)

//user
type User struct {
	*model.UserModel
	service    *cham.Service
	session    uint32
	activeTime time.Time
}

func newUser(service *cham.Service, u *model.UserModel, session uint32) (*User, error) {
	m, _, err := db.DbCache.GetPk(u)
	if err != nil {
		return nil, err
	}
	user := &User{m.(*UserModel), service, session, time.Now()}
	go user.Run()
	return user, nil
}

func (user *User) Save() {
	db.DbCache.UpdateModel(m.UserModel)
}

func (user *User) Response(data []byte) error {
	result := user.service.Call("gate", cham.PTYPE_RESPONSE, user.session, data)
	err := result[0].(error)
	return err
}

func (user *User) Run() {
	t := cham.DTimer.NewTicker(time.Minute * 5)
	for {
		select {
		case t := <-t.C:
			log.Infoln("user every 5 Minute check start, openid:", user.Openid)
			user.Save()
			log.Infoln("user every 5 Minute check end, openid:", user.Openid)
		}
	}
}

func Start(service *cham.Service, args ...interface{}) cham.Dispatch {
	log.Infoln("New Service ", service.String())
	openid, session := args[0].(string), args[1].(uint32)
	user, err := newUser(service, &model.UserModel{Openid: openid}, session)
	if err != nil {
		log.Errorln("Service ", service.String(), "init error,", err.Error())
		service.Stop()
	}
	return func(session int32, source cham.Address, ptype uint8, args ...interface{}) []interface{} {
		return cham.NORET
	}
}
