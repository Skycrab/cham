package db

import (
	"cham/lib/database"
	"cham/lib/lru"
	"database/sql"
	// "fmt"
	_ "github.com/go-sql-driver/mysql"
)

var DbCache *QueryCache

type QueryCache struct {
	conn        *sql.DB
	Db          *database.Database
	queryset    map[string]*lru.Cache
	expires     chan lru.Value
	updateModel chan database.Model
}

func New(conn *sql.DB) *QueryCache {
	q := &QueryCache{
		conn:        conn,
		Db:          database.New(conn),
		queryset:    make(map[string]*lru.Cache),
		expires:     make(chan lru.Value, 1000),
		updateModel: make(chan database.Model, 1000),
	}
	go q.Run()
	return q
}

func (q *QueryCache) Run() {
	for {
		select {
		case m := <-q.updateModel:
			q.Db.UpdatePk(m)
		case v := <-q.expires:
			q.Db.UpdatePk(v.(database.Model))
		}
	}
}

func (q *QueryCache) UpdateModel(m database.Model) {
	q.updateModel <- m
}

func (q *QueryCache) Register(m database.Model, maxEntries int) {
	name := m.TableName()
	if _, ok := q.queryset[name]; ok {
		panic("duplicate queryset table:" + name)
	}
	q.queryset[name] = lru.New(maxEntries, func(key lru.Key, value lru.Value) {
		q.expires <- value
	})
}

func (q *QueryCache) GetPk(m database.Model) (database.Model, interface{}, error) {
	name := m.TableName()
	if _, ok := q.queryset[name]; !ok {
		panic("please Register first table:" + name)
	}
	_, value := database.GetPkValue(m)
	cache := q.queryset[name]
	v, ok := cache.Get(value)
	if !ok {
		err := q.Db.GetPk(m)
		if err != nil {
			return nil, nil, err
		}
		cache.Add(value, m)
		return m, value, nil
	}
	vv := v.(database.Model)
	return vv, value, nil
}

func init() {
	conn, err := sql.Open("mysql", "root:dajidan2bu2@tcp(10.9.28.162:3306)/brain")
	if err != nil {
		panic("mysql connect error," + err.Error())
	}
	DbCache = New(conn)
}
