package usermanager

import (
	"cham/cham"
	"cham/service/log"
	"fmt"
	"question/usermanager/user"
	"sync"
	"time"
)

//manager
type UserManager struct {
	sync.RWMutex
	users map[string]*cham.Service
}

func New() *UserManager {
	return &UserManager{users: make(map[string]*cham.Service)}
}

func (um *UserManager) Add(openid string, session uint32) {
	um.RLock()
	if us, ok := um.users[openid]; ok {
		um.RUnlock()
		return
	}
	um.RUnlock()
	us := cham.NewService(fmt.Sprintf("user-%d", openid), user.Start, openid, session)
	um.Lock()
	um.users[openid] = us
	um.Unlock()
}

func (um *UserManager) Delete(openid string) {
	um.Lock()
	delete(um.users, openid)
	um.Unlock()
}

func (um *UserManager) Handle() {

}
