package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
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

func NewSDB(addr, user, pass string) (d *SDB, e error) {
	d = new(SDB)
	d.db, e = sql.Open("postgres",
		fmt.Sprintf("postgres://%s:%s@%s", user, pass, addr))
	return
}

type SDB struct {
	db *sql.DB
}

func (d *SDB) ServeHTTP(w h.ResponseWriter, r *h.Request) {
	var e error
	if r.Method == h.MethodGet {
		v := r.URL.Query()
		page, size := v.Get("page"), v.Get("size")
		d.query(page, size)
	} else {
		e = NotSuppMeth(r.Method)
	}
	writeErr(w, e)
}

func writeErr(w h.ResponseWriter, e error) {
	if e != nil {
		// The order of the following commands matter since
		// httptest.ResponseRecorder ignores parameter sent
		// to WriteHeader if Write was called first
		w.WriteHeader(h.StatusBadRequest)
		w.Write([]byte(e.Error()))
	}
}

// NotSuppMeth is the not supported method message
func NotSuppMeth(m string) (e error) {
	e = fmt.Errorf("Not supported method %s", m)
	return
}

func (d SDB) query(page, size string) (s []DBRecord, e error) {
	query := fmt.Sprintf(
		"SELECT id_student,identification,name,"+
			"middle_name,last_name,address,phone FROM student LIMIT %s"+
			"OFFSET %s",
		size,
		page)
	var r *sql.Rows
	r, e = d.db.Query(query)
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
		b = r.Next()
	}
	return
}
