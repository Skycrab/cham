package usermanager

import (
	"cham/cham"
	"cham/service/log"
	"fmt"
	"question/protocol"
	"question/usermanager/user"
	"sync"
)

//manager
type UserManager struct {
	sync.RWMutex
	users map[string]*cham.Service
}

func New() *UserManager {
	return &UserManager{users: make(map[string]*cham.Service)}
}

func (um *UserManager) Add(request *protocol.Login, session uint32) {
	openid := request.Openid
	um.RLock()
	if us, ok := um.users[openid]; ok {
		um.RUnlock()
		us.NotifySelf(cham.PTYPE_GO, user.CHANGE_SESSION, session)
		return
	}
	um.RUnlock()
	log.Infoln("new user,openid:", openid)
	us := cham.NewService(fmt.Sprintf("user-%s", openid), user.Start, 1, openid, session)
	if us == nil {
		log.Errorln("new user failed, openid:", openid)
		return
	}
	um.Lock()
	um.users[openid] = us
	um.Unlock()
}

func (um *UserManager) Delete(openid string) {
	um.Lock()
	delete(um.users, openid)
	um.Unlock()
}

func (um *UserManager) Handle(session uint32, protocolName string, request interface{}) {
	fmt.Printf("protocol: %#v\n", request)
	if protocolName == "login" {
		um.Add(request.(*protocol.Login), session)
	}

}
