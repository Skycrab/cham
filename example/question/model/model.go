package model

import (
	"question/db"
	"time"
)

//UserModel
type UserModel struct {
	ID         int       `json:"id"`
	Openid     string    `json:"openid" pk:"true"`
	Name       string    `json:"name"`
	Headimgurl string    `json:"headimgurl"`
	Sex        int       `json:"sex"`
	LastLogin  time.Time `json:"lastlogin" field:"last_login"`
}

func (u *UserModel) TableName() string {
	return "question_user"
}

//register all model
func init() {
	dc := db.DbCache
	dc.Register(&UserModel{}, 1)
}
