package sigapi

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"html/template"
	h "net/http"
	"time"
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

func NewPostgreSDB(addr, user, pass, tpf string) (d *SDB, e error) {
	var db *sql.DB
	db, e = sql.Open("postgres",
		fmt.Sprintf("postgres://%s:%s@%s", user, pass, addr))
	if e == nil {
		db.SetMaxOpenConns(200)
		db.SetMaxIdleConns(0)
		db.SetConnMaxLifetime(5 * time.Second)
	}
	var tp *template.Template
	if e == nil {
		tp, e = template.New("doc").ParseFiles(tpf)
	}
	if e == nil {
		d = &SDB{Db: db, tp: tp}
		rt := mux.NewRouter()
		rt.HandleFunc("/", d.docHn)
		d.pagPath = fmt.Sprintf("/paginador/{%s:[0-9]+}/{%s:[0-9]+}",
			Offset, Size)
		rt.HandleFunc(d.pagPath, d.pagHn).Methods(h.MethodGet)
		d.idPath = fmt.Sprintf("/sigenu-id/{%s:[a-z0-9:-]+}", Id)
		rt.HandleFunc(d.idPath, d.idHn).Methods(h.MethodGet)
		d.namePath = fmt.Sprintf("/sigenu-nombre/{%s}", Name)
		rt.HandleFunc(d.namePath, d.nameHn).Methods(h.MethodGet)
		d.evalPath = fmt.Sprintf("/eval/{%s:[a-z0-9:-]+}", IN)
		rt.HandleFunc(d.evalPath, d.evaluationsHn).Methods(h.MethodGet)
		d.Handler = cors.AllowAll().Handler(rt)
	}
	return
}

type SDB struct {
	Db       *sql.DB
	Handler  h.Handler
	tp       *template.Template
	pagPath  string
	idPath   string
	namePath string
	evalPath string
}

const (
	Offset   = "offset"
	Size     = "size"
	Id       = "id"
	Name     = "nombre"
	IN       = "identity"
	NamePgLn = 100
)

func (d *SDB) docHn(w h.ResponseWriter, r *h.Request) {
	e := d.tp.ExecuteTemplate(w, "doc", struct {
		PagPath  string
		IdPath   string
		NamePath string
	}{
		PagPath:  d.pagPath,
		IdPath:   d.idPath,
		NamePath: d.namePath,
	})
	writeErr(w, e)
}

func (d *SDB) pagHn(w h.ResponseWriter, r *h.Request) {
	rg := mux.Vars(r)
	offset := rg[Offset]
	size := rg[Size]
	s, e := d.queryRange(offset, size)
	if e == nil {
		e = Encode(w, s)
	}
	writeErr(w, e)
}

func (d *SDB) nameHn(w h.ResponseWriter, r *h.Request) {
	rg := mux.Vars(r)
	name := rg[Name]
	s, e := d.queryName(name)
	if e == nil {
		e = Encode(w, s)
	}
	writeErr(w, e)
}

func (d *SDB) idHn(w h.ResponseWriter, r *h.Request) {
	rg := mux.Vars(r)
	sigId := rg[Id]
	n, e := d.queryId(sigId)
	if e == nil {
		e = Encode(w, n)
	}
	writeErr(w, e)
}

func (d *SDB) queryName(name string) (s []DBRecord, e error) {
	query := "SELECT id_student,identification,name," +
		"middle_name,last_name,address,phone,student_status_fk," +
		"faculty_fk,career_fk FROM student WHERE " +
		"name LIKE '%" + name + "%' " +
		"OR middle_name LIKE '%" + name + "%' " +
		"OR last_name LIKE '%" + name + "%'"
	var r *sql.Rows
	r, e = d.Db.Query(query)
	s = make([]DBRecord, NamePgLn)
	end, i := e != nil, 0
	for !end {
		var n DBRecord
		next := r.Next()
		if next {
			n, e = d.scanStudent(r)
		}
		if e == nil && next {
			s[i], i = n, i+1
		}
		end = i == NamePgLn || e != nil || !next
	}
	if r != nil {
		r.Close()
	}
	s = s[:i]
	return
}

func (d *SDB) queryId(id string) (s *DBRecord, e error) {
	query := fmt.Sprintf("SELECT id_student,identification,name,"+
		"middle_name,last_name,address,phone,student_status_fk,"+
		"faculty_fk,career_fk FROM student WHERE id_student = '%s'",
		id)
	var r *sql.Rows
	r, e = d.Db.Query(query)
	var n DBRecord
	if e == nil && r.Next() {
		n, e = d.scanStudent(r)
	}
	if e == nil {
		s = &n
	}
	if r != nil {
		r.Close()
	}
	return
}

func (d *SDB) queryRange(offset, size string) (s []DBRecord, e error) {
	// facultad, graduado
	query := fmt.Sprintf("SELECT id_student,identification,name,"+
		"middle_name,last_name,address,phone,student_status_fk,"+
		"faculty_fk,career_fk FROM student LIMIT %s OFFSET %s",
		size, offset)
	var r *sql.Rows
	r, e = d.Db.Query(query)
	s = make([]DBRecord, 0)
	for e == nil && r.Next() {
		var n DBRecord
		n, e = d.scanStudent(r)
		if e == nil {
			s = append(s, n)
		}
	}
	if r != nil {
		r.Close()
	}
	if e == nil {
		e = r.Err()
	}
	return
}

func (d *SDB) scanStudent(r *sql.Rows) (s DBRecord, e error) {
	s = DBRecord{}
	var name, middle_name, last_name, car_fk string
	var stat_fk, fac_fk sql.NullString
	e = r.Scan(&s.Id, &s.IN, &name, &middle_name, &last_name,
		&s.Addr, &s.Tel, &stat_fk, &fac_fk, &car_fk)
	var ax *auxInf
	if e == nil {
		ax, e = d.queryAux(stat_fk.String, fac_fk.String, car_fk)
	}
	if e == nil {
		s.Career, s.Faculty, s.Status = ax.career, ax.faculty,
			ax.status
		s.Name = name + " " + middle_name + " " + last_name
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
