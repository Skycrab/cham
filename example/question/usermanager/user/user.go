package user

import (
	"cham/cham"
	"cham/service/log"
	// "fmt"
	"question/db"
	"question/model"
	"question/protocol"
	"time"
)

const (
	CHANGE_SESSION uint8 = iota
	DELETE_USER
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
	user := &User{m.(*model.UserModel), service, session, time.Now()}
	go user.Run()
	return user, nil
}

func (user *User) Save() {
	db.DbCache.UpdateModel(user.UserModel)
}

func (user *User) Response(data []byte) error {
	result := user.service.Call("gate", cham.PTYPE_RESPONSE, user.session, data)
	err := result[0]
	if err != nil {
		return err.(error)
	}
	return nil
}

func (user *User) Kill() {
	user.service.Call("broker", cham.PTYPE_GO, DELETE_USER, user.Openid)
	user.service.Stop()
	user.Save()
}

func (user *User) Run() {
	t := cham.DTimer.NewTicker(time.Minute * 5)
	t2 := cham.DTimer.NewTicker(time.Second * 5)
	heart := protocol.HeartBeat{}
	lost := 0
	for {
		select {
		case t := <-t.C:
			log.Infoln("every 5 Minute check start, openid:", user.Openid, " time:", t)
			user.Save()
		case t2 := <-t2.C:
			log.Infoln("every 5 Second heart beat, time:", t2)
			err := user.Response(protocol.Encode(0, heart))
			if err != nil {
				lost++
				log.Infoln("user heart beat lost,openid:", user.Openid)
			} else {
				lost = 0
			}
			// fmt.Println("---------", lost)
			if lost >= 5 {
				log.Infoln("heart beat lost more 5 time, close user service, openid:", user.Openid)
				user.Kill()
				return
			}
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
		cmd := args[0].(uint8)
		switch cmd {
		case CHANGE_SESSION:
			user.session = args[1].(uint32)
		}

		return cham.NORET
	}
}
