package database

import (
	"bytes"
	"database/sql"
	// "fmt"
	"reflect"
	"strings"
)

const (
	tablenameFunc = "TableName"
	fieldTag      = "field"
	primaryKey    = "id"
	attrTag       = "attr"
	autoAttr      = "auto"
)

type Model interface{}

type Database struct {
	db *sql.DB
}

func (d *Database) New(db *sql.DB) *Database {
	return &Database{db}
}

func (d *Database) Close() {
	d.db.Close()
}

//need value to reduce once reflect expense
// d.Get(&User{}, "openid", "123456")
func (d *Database) Get(m Model, field string, value interface{}) error {
	q, n := query(m, field, 0)
	row := d.db.QueryRow(q, value)
	v := reflect.ValueOf(m).Elem()
	args := make([]interface{}, n)
	for i := 0; i < n; i++ {
		args[i] = v.Field(i).Addr().Interface()
	}
	return row.Scan(args...)
}

//primary key get
func (d *Database) GetPk(m Model, value interface{}) error {
	return d.Get(m, primaryKey, value)
}

//mysql in usage
// ms := d.GetMultiIn(&User{}, "openid", "12", "34")
// ms[0].(*User) [return interface{} need type change]
func (d *Database) GetMultiIn(m Model, field string, value ...interface{}) (ms []Model, err error) {
	q, n := query(m, field, len(value))
	rows, err := d.db.Query(q, value...)
	if err != nil {
		return
	}
	defer rows.Close()
	args := make([]interface{}, n)
	t := reflect.TypeOf(m).Elem()
	for rows.Next() {
		v := reflect.New(t).Elem()
		for i := 0; i < n; i++ {
			args[i] = v.Field(i).Addr().Interface()
		}
		if err = rows.Scan(args...); err != nil {
			return
		}
		ms = append(ms, v.Addr().Interface())
	}
	err = rows.Err()
	return
}

func (d *Database) GetMultiPkIn(m Model, value ...interface{}) ([]Model, error) {
	return d.GetMultiIn(m, primaryKey, value...)
}

func (d *Database) Del(m Model, field string, value interface{}) (affect int64, err error) {
	q := delQuery(m, field)
	stmt, err := d.db.Prepare(q)
	if err != nil {
		return
	}
	res, err := stmt.Exec(value)
	if err != nil {
		return
	}
	return res.RowsAffected()
}

func (d *Database) DelPk(m Model, value interface{}) (int64, error) {
	return d.Del(m, primaryKey, value)
}

func (d *Database) Update(m Model, field string, value interface{}) (affect int64, err error) {
	q, n, auto := updateQuery(m, field)
	stmt, err := d.db.Prepare(q)
	if err != nil {
		return
	}
	values := structValues(m, n, auto)
	vv := make([]interface{}, len(values)+1)
	copy(vv, values)
	vv[len(values)] = value
	res, err := stmt.Exec(vv...)
	if err != nil {
		return
	}

	return res.RowsAffected()
}

func (d *Database) Insert(m Model) (last int64, err error) {
	q, n, auto := insertquery(m)
	stmt, err := d.db.Prepare(q)
	if err != nil {
		return
	}
	res, err := stmt.Exec(structValues(m, n, auto)...)
	if err != nil {
		return
	}
	return res.LastInsertId()
}

// return {struct keys, NumField, auto field index}
func structKeys(m Model, autoNeed bool) ([]string, int, int) {
	t := reflect.TypeOf(m).Elem()
	n := t.NumField()
	keys := make([]string, 0, n)
	auto := -1
	for i := 0; i < n; i++ {
		f := t.Field(i)
		if !autoNeed && f.Tag.Get(attrTag) == autoAttr {
			if auto != -1 {
				panic("table duplicate auto pk:" + tableName(m))
			}
			auto = i
			continue
		}
		tag := f.Tag.Get(fieldTag)
		if tag == "" {
			tag = strings.ToLower(f.Name)
		}
		keys = append(keys, tag)
	}

	return keys, n, auto
}

// m must reflect.Ptr
//get all struct value except auto
func structValues(m Model, n int, auto int) []interface{} {
	N := n
	if auto != -1 {
		n--
	}
	vv := make([]interface{}, 0, n)
	v := reflect.ValueOf(m).Elem()
	for i := 0; i < N; i++ {
		if i == auto {
			continue
		}
		vv = append(vv, v.Field(i).Interface())
	}
	return vv
}

// m must reflect.Ptr
func tableName(m Model) string {
	v := reflect.ValueOf(m)
	if method := v.MethodByName(tablenameFunc); method.IsValid() {
		return method.Call([]reflect.Value{})[0].String()
	} else {
		return strings.ToLower(v.Elem().Type().Name())
	}
}

// m must reflect.Ptr
// inlen : multiIN number
func query(m Model, field string, inlen int) (string, int) {
	buf := bytes.NewBufferString("SELECT ")
	keys, n, _ := structKeys(m, true)
	buf.WriteString(strings.Join(keys, ", "))
	buf.WriteString(" FROM ")
	buf.WriteString(tableName(m))
	buf.WriteString(" Where ")
	buf.WriteString(field)
	if inlen <= 0 {
		buf.WriteString("=?")
	} else {
		buf.WriteString(" IN(")
		var tmp2 []string
		// try reuse keys
		if cap(keys) >= inlen {
			tmp2 = keys[0:0]
		} else {
			tmp2 = make([]string, 0, inlen)
		}
		for i := 0; i < inlen; i++ {
			tmp2 = append(tmp2, "?")
		}
		buf.WriteString(strings.Join(tmp2, ", "))
		buf.WriteString(")")
	}
	return buf.String(), n
}

func delQuery(m Model, field string) string {
	buf := bytes.NewBufferString("DELETE FROM ")
	buf.WriteString(tableName(m))
	buf.WriteString(" WHERE ")
	buf.WriteString(field)
	buf.WriteString("=?")
	return buf.String()
}

func updateQuery(m Model, field string) (string, int, int) {
	buf := bytes.NewBufferString("UPDATE ")
	buf.WriteString(tableName(m))
	buf.WriteString(" SET ")
	keys, n, auto := structKeys(m, false)
	for i, k := range keys {
		buf.WriteString(k)
		buf.WriteString("=?")
		if n != 2 && i != n-1 { // one key || last
			buf.WriteString(" ,")
		}
	}
	buf.WriteString(" WHERE ")
	buf.WriteString(field)
	buf.WriteString("=?")

	return buf.String(), n, auto
}

// return {insert string, struct NumField, auto field index}
func insertquery(m Model) (string, int, int) {
	buf := bytes.NewBufferString("INSERT INTO ")
	buf.WriteString(tableName(m))
	buf.WriteString("(")
	keys, n, auto := structKeys(m, false)
	buf.WriteString(strings.Join(keys, ", "))
	buf.WriteString(")")
	buf.WriteString(" VALUES(")
	for i := 0; i < len(keys); i++ {
		keys[i] = "?"
	}
	buf.WriteString(strings.Join(keys, ", "))
	buf.WriteString(")")
	return buf.String(), n, auto
}
