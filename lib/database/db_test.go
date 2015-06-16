package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"testing"
)

var db *sql.DB

type User struct {
	Id   int `attr:"auto"`
	Name string
}

func (u *User) TableName() string {
	return "users"
}

type UserWithName struct {
	Id   int
	Name string `field:"username"`
}

func (u *UserWithName) TableName() string {
	return "usertable"
}

func TestTableName(t *testing.T) {
	u := &User{}
	if tableName(u) != "users" {
		t.Error("User tableName error")
	}
	uu := &UserWithName{}
	if tableName(uu) != "usertable" {
		t.Error("UserWithName tableName error")
	}
	fmt.Println("end")
}

func TestQuery(t *testing.T) {
	uu := &UserWithName{}
	fmt.Println(query(uu, "name", 0))
}

func TestGet(t *testing.T) {
	u := &User{}
	d := &Database{db}
	fmt.Println(d.Get(u, "id", 1))
	fmt.Println(u)

	fmt.Println("------------------")
	v, e := d.GetMultiIn(u, "id", 3, 4)
	if e != nil {
		t.Error("GetMultiIn error")
	} else {
		for _, m := range v {
			fmt.Println(m.(*User))
		}
	}

}

func TestDel(t *testing.T) {
	u := &User{}
	d := &Database{db}
	fmt.Println(d.Del(u, "id", 1))
	d.DelPk(u, 2)
}

func TestInsert(t *testing.T) {
	u := &User{Name: "kehan"}
	d := &Database{db}
	fmt.Println(d.Insert(u))
}

func TestUpdate(t *testing.T) {
	u := &User{}
	d := &Database{db}
	d.GetPk(u, 3)
	fmt.Println(u)
	u.Name = "3333"
	fmt.Println(d.Update(u, "id", u.Id))
	fmt.Println(u)
}
func init() {
	db = new(sql.DB)
	d, err := sql.Open("mysql", "root@tcp(localhost:3306)/test")
	if err != nil {
		fmt.Println("mysql open error, please check:" + err.Error())
		os.Exit(1)
	} else {
		db = d
	}
}
