package database

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

const (
	tablenameFunc = "TableName"
	fieldTag      = "field"
	pkTag         = "pk" // primary key
	attrTag       = "attr"
	autoAttr      = "auto" //auto increamemt,e.g. id
)

const (
	TIME_FORMATER = "2006-01-02 15:04:05"
)

var (
	TIME_PTR_TYPE = reflect.TypeOf(&time.Time{})
)

type Model interface {
	TableName() string
}

type Database struct {
	db    *sql.DB
	Debug bool
}

type Scanner interface {
	Scan(dest ...interface{}) error
}

func New(db *sql.DB) *Database {
	return &Database{db, false}
}

func (d *Database) Close() {
	d.db.Close()
}

//sql provide row and rows
func scan(row Scanner, v reflect.Value, args []interface{}) error {
	//need support time.Time,save time.Time index and value
	keys := make([]int, 0)
	values := make([]*string, 0)
	for i := 0; i < len(args); i++ {
		vv := v.Field(i).Addr()
		t := vv.Type()
		if t == TIME_PTR_TYPE {
			var tmp string
			args[i] = &tmp
			keys = append(keys, i)
			values = append(values, &tmp)
		} else {
			args[i] = vv.Interface()
		}

	}
	err := row.Scan(args...)
	if err == nil {
		for i := 0; i < len(keys); i++ {
			t, err := time.Parse(TIME_FORMATER, *values[i])
			if err == nil {
				v.Field(keys[i]).Set(reflect.ValueOf(t))
			}
		}
	}
	return err
}

//need value to reduce once reflect expense
// d.Get(&User{}, "openid", "123456")
func (d *Database) Get(m Model, field string, value interface{}) error {
	q, n := query(m, field, "", 0)
	if d.Debug {
		fmt.Println(q, " [", value, "]")
	}
	row := d.db.QueryRow(q, value)
	v := reflect.ValueOf(m).Elem()
	args := make([]interface{}, n)
	// for i := 0; i < n; i++ {
	// 	args[i] = v.Field(i).Addr().Interface()
	// }
	// return row.Scan(args...)
	return scan(row, v, args)
}

//primary key get
func (d *Database) GetPk(m Model) error {
	keys, _, _, pkIndex := structKeys(m, true)
	if pkIndex == -1 {
		panic("no pk tag, table: " + m.TableName())
	}
	value := reflect.ValueOf(m).Elem().Field(pkIndex).Interface()
	return d.Get(m, keys[pkIndex], value)
}

func (d *Database) Select(m Model, field string, condition interface{}, value ...interface{}) (ms []Model, err error) {
	q, n := query(m, field, condition, len(value))
	if d.Debug {
		fmt.Println(q, " ", value)
	}
	rows, err := d.db.Query(q, value...)
	if err != nil {
		return
	}
	defer rows.Close()
	args := make([]interface{}, n)
	t := reflect.TypeOf(m).Elem()
	for rows.Next() {
		v := reflect.New(t).Elem()
		// for i := 0; i < n; i++ {
		// 	args[i] = v.Field(i).Addr().Interface()
		// }
		// if err = rows.Scan(args...); err != nil {
		// 	return
		// }
		if err = scan(rows, v, args); err != nil {
			return
		}
		ms = append(ms, (v.Addr().Interface()).(Model))
	}
	err = rows.Err()
	return
}

// d.GetCondition(&User{}, "where level >10 and money <? limit ?, 10", 100, 5)
func (d *Database) GetCondition(m Model, condition interface{}, value ...interface{}) ([]Model, error) {
	return d.Select(m, "", condition, value...)
}

//mysql in usage
// ms := d.GetMultiIn(&User{}, "openid", "12", "34")
// ms[0].(*User) [return interface{} need type change]
func (d *Database) GetMultiIn(m Model, field string, value ...interface{}) ([]Model, error) {
	return d.Select(m, field, "", value...)
}

func (d *Database) GetMultiPkIn(m Model, value ...interface{}) ([]Model, error) {
	keys, _, _, pkIndex := structKeys(m, true)
	if pkIndex == -1 {
		panic("no pk tag, table: " + m.TableName())
	}
	return d.GetMultiIn(m, keys[pkIndex], value...)
}

func (d *Database) Del(m Model, field string, value interface{}) (affect int64, err error) {
	q := delQuery(m, field)
	if d.Debug {
		fmt.Println(q, " [", value, "]")
	}
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

func (d *Database) DelPk(m Model) (int64, error) {
	keys, _, _, pkIndex := structKeys(m, true)
	if pkIndex == -1 {
		panic("no pk tag, table: " + m.TableName())
	}
	value := reflect.ValueOf(m).Elem().Field(pkIndex).Interface()
	return d.Del(m, keys[pkIndex], value)
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
	if d.Debug {
		fmt.Println(q, " ", values)
	}
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
	values := structValues(m, n, auto)
	if d.Debug {
		fmt.Println(q, " ", values)
	}
	res, err := stmt.Exec(values...)
	if err != nil {
		return
	}
	lastId, err := res.LastInsertId()
	_, _, autoIndex, _ := structKeys(m, true)
	if autoIndex != -1 {
		reflect.ValueOf(m).Elem().Field(autoIndex).SetInt(lastId)
	}

	return lastId, err
}

//reflect is expensive,so cache model's field
var fieldCache = make(map[string][]interface{})

//delete autoIndex
func autoDelete(keys []string, autoIndex int) []string {
	fmt.Println(keys, autoIndex)
	keysNew := make([]string, len(keys)-1)
	copy(keysNew, keys[0:autoIndex])
	copy(keysNew, keys[autoIndex+1:])
	fmt.Println(keysNew)
	return keysNew
}

// return {struct keys, NumField, auto field index, pk field index}
func structKeys(m Model, autoNeed bool) ([]string, int, int, int) {
	//cache check start
	table := m.TableName()
	if cache, ok := fieldCache[table]; ok {
		keys, numField, autoIndex, pkIndex := cache[0].([]string), cache[1].(int), cache[2].(int), cache[3].(int)
		if !autoNeed && autoIndex != -1 {
			return autoDelete(keys, autoIndex), numField, autoIndex, pkIndex
		}
		return keys, numField, autoIndex, pkIndex
	}
	//cache check end

	t := reflect.TypeOf(m).Elem()
	n := t.NumField()
	keys := make([]string, 0, n)
	autoIndex, pkIndex := -1, -1
	for i := 0; i < n; i++ {
		f := t.Field(i)
		if f.Tag.Get(attrTag) == autoAttr {
			if autoIndex != -1 {
				panic("table duplicate auto :" + m.TableName())
			}
			autoIndex = i
		}
		tag := f.Tag.Get(fieldTag)
		if tag == "" {
			tag = strings.ToLower(f.Name)
		}
		pk := f.Tag.Get(pkTag)
		if pk == "true" {
			pkIndex = i
		}
		keys = append(keys, tag)
	}

	//cache
	fieldCache[table] = []interface{}{keys, n, autoIndex, pkIndex}
	if !autoNeed && autoIndex != -1 {
		return autoDelete(keys, autoIndex), n, autoIndex, pkIndex
	}

	return keys, n, autoIndex, pkIndex
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
		vt := v.Field(i).Addr().Type()
		var tmp interface{}
		if vt == TIME_PTR_TYPE {
			tmp = v.Field(i).Interface().(time.Time).Format(TIME_FORMATER)
		} else {
			tmp = v.Field(i).Interface()
		}
		vv = append(vv, tmp)
	}
	return vv
}

// if you don't want provide TableName, you can combine DeafultModel
type DeafultModel struct{}

func (m *DeafultModel) TableName() string {
	v := reflect.ValueOf(m)
	return strings.ToLower(v.Elem().Type().Name())
}

// m must reflect.Ptr
func tableName(m interface{}) string {
	v := reflect.ValueOf(m)
	if method := v.MethodByName(tablenameFunc); method.IsValid() {
		return method.Call([]reflect.Value{})[0].String()
	} else {
		return strings.ToLower(v.Elem().Type().Name())
	}
}

// m must reflect.Ptr
// inlen : multiIN number
// condition nil -> select all, string -> condition
func query(m Model, field string, condition interface{}, inlen int) (string, int) {
	buf := bytes.NewBufferString("SELECT ")
	keys, n, _, _ := structKeys(m, true)
	buf.WriteString(strings.Join(keys, ", "))
	buf.WriteString(" FROM ")
	buf.WriteString(m.TableName())

	if condition != nil {
		if c, ok := condition.(string); ok && c != "" {
			buf.WriteByte(' ')
			buf.WriteString(c)
		} else {
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
		}
	}
	return buf.String(), n
}

func delQuery(m Model, field string) string {
	buf := bytes.NewBufferString("DELETE FROM ")
	buf.WriteString(m.TableName())
	buf.WriteString(" WHERE ")
	buf.WriteString(field)
	buf.WriteString("=?")
	return buf.String()
}

func updateQuery(m Model, field string) (string, int, int) {
	buf := bytes.NewBufferString("UPDATE ")
	buf.WriteString(m.TableName())
	buf.WriteString(" SET ")
	keys, n, auto, _ := structKeys(m, false)
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
	buf.WriteString(m.TableName())
	buf.WriteString("(")
	keys, n, auto, _ := structKeys(m, false)
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
