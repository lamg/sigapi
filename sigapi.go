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
	Tel     string `json:"tel"`
	Career  string `json:"career"`
	Faculty string `json:"faculty"`
	Status  string `json:"status"`
}

func NewPostgreSDB(addr, user, pass string) (d *SDB, e error) {
	var db *sql.DB
	db, e = sql.Open("postgres",
		fmt.Sprintf("postgres://%s:%s@%s", user, pass, addr))
	if e == nil {
		d = &SDB{Db: db}
	}
	return
}

type SDB struct {
	Db *sql.DB
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
	// facultad, graduado
	query := fmt.Sprintf(
		"SELECT id_student,identification,name,"+
			"middle_name,last_name,address,phone,student_status_fk,"+
			"faculty_fk,career_fk FROM student LIMIT %s OFFSET %s",
		size,
		offset)
	var r *sql.Rows
	r, e = d.Db.Query(query)
	s = make([]DBRecord, 0)
	for e == nil && r.Next() {
		st := DBRecord{}
		var name, middle_name, last_name, car_fk string
		var stat_fk, fac_fk sql.NullString
		e = r.Scan(&st.Id, &st.IN, &name, &middle_name, &last_name,
			&st.Addr, &st.Tel, &stat_fk, &fac_fk, &car_fk)
		var ax *auxInf
		if e == nil {
			ax, _ = d.queryAux(stat_fk.String, fac_fk.String, car_fk)
		}
		if e == nil {
			st.Career, st.Faculty, st.Status = ax.career, ax.faculty,
				ax.status
			st.Name = name + " " + middle_name + " " + last_name
			s = append(s, st)
		}
	}
	if r != nil && e == nil {
		e = r.Err()
		r.Close()
	}
	return
}

type auxInf struct {
	status  string
	faculty string
	career  string
}

func (d *SDB) queryAux(stat_fk, fac_fk,
	car_fk string) (f *auxInf, e error) {
	f = new(auxInf)
	q, a, r := []string{
		"SELECT kind FROM student_status WHERE id_student_status = ",
		"SELECT name FROM faculty WHERE id_faculty = ",
		"SELECT national_career_fk FROM career WHERE id_career = ",
	},
		[]string{
			stat_fk,
			fac_fk,
			car_fk,
		},
		make([]sql.NullString, 3)

	for i := 0; e == nil && i != len(q); i++ {
		qr := q[i] + "'" + a[i] + "'"
		s, e := d.Db.Query(qr)
		if e == nil {
			if s.Next() {
				e = s.Scan(&r[i])
			} else {
				e = fmt.Errorf("Información no encontrada")
			}
		}
		if e == nil {
			e = s.Err()
			s.Close()
		}
	}
	var s *sql.Rows
	if e == nil {
		f.status = r[0].String
		f.faculty = r[1].String
		if r[2].Valid {
			ncq := "SELECT name FROM national_career" +
				" WHERE id_national_career = '" + r[2].String + "'"
			s, e = d.Db.Query(ncq)
		}
	}
	if e == nil {
		if r[2].Valid {
			if s.Next() {
				e = s.Scan(&f.career)
			} else {
				e = fmt.Errorf("Información no encontrada")
			}
		}
	}
	if e == nil && s != nil {
		e = s.Err()
		s.Close()
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
