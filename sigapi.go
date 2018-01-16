package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"io"
	h "net/http"
)

type DBRecord struct {
	//database key field
	Id string `json:"id"`
	//identity number
	IN string `json:"in"`
	//person name
	Name string `json:"name"`
	//address
	Addr string `json:"addr"`
	//telephone number
	Tel string `json:"tel"`
}

func NewPostgreSDB(addr, user, pass string) (d *SDB, e error) {
	var db *sql.DB
	db, e = sql.Open("postgres",
		fmt.Sprintf("postgres://%s:%s@%s", user, pass, addr))
	if e == nil {
		d = &SDB{Db: &sqQr{db: db}}
	}
	return
}

type sqQr struct {
	db *sql.DB
}

func (s *sqQr) Query(q string,
	args ...interface{}) (c Scanner, e error) {
	c, e = s.db.Query(q, args...)
	return
}

type SDB struct {
	Db Querier
}

type Querier interface {
	Query(string, ...interface{}) (Scanner, error)
}

type Scanner interface {
	Next() bool
	Scan(...interface{}) error
}

func (d *SDB) ServeHTTP(w h.ResponseWriter, r *h.Request) {
	var e error
	var s []DBRecord
	if r.Method == h.MethodGet {
		v := r.URL.Query()
		offset, size := v.Get("offset"), v.Get("size")
		s, e = d.query(offset, size)
	} else {
		e = NotSuppMeth(r.Method)
	}
	if e == nil {
		e = Encode(w, s)
	}
	writeErr(w, e)
}

func writeErr(w h.ResponseWriter, e error) {
	if e != nil {
		// The order of the following commands matters since
		// httptest.ResponseRecorder ignores parameter sent
		// to WriteHeader if Write was called first
		w.WriteHeader(h.StatusBadRequest)
		w.Write([]byte(e.Error()))
	}
}

func (d *SDB) query(offset, size string) (s []DBRecord, e error) {
	query := fmt.Sprintf(
		"SELECT id_student,identification,name,"+
			"middle_name,last_name,address,phone FROM student LIMIT %s"+
			" OFFSET %s",
		size,
		offset)
	var r Scanner
	r, e = d.Db.Query(query)
	b, s := e == nil, make([]DBRecord, 0)
	for b {
		st := DBRecord{}
		var name, middle_name, last_name string
		e = r.Scan(&st.Id, &st.IN, &name, &middle_name, &last_name,
			&st.Addr, &st.Tel)
		if e == nil {
			st.Name = name + " " + middle_name + " " + last_name
			s = append(s, st)
		}
		b = r.Next() && e == nil
	}
	return
}

// Encode encodes an object in JSON notation into w
func Encode(w io.Writer, v interface{}) (e error) {
	cd := json.NewEncoder(w)
	cd.SetIndent("	", "")
	e = cd.Encode(v)
	return
}

// NotSuppMeth is the not supported method message
func NotSuppMeth(m string) (e error) {
	e = fmt.Errorf("Not supported method %s", m)
	return
}
