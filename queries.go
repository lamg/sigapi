package sigapi

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/stdlib"
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

func (d *SDB) openDB() (e error) {
	d.db, e = sql.Open("pgx",
		fmt.Sprintf("postgres://%s:%s@%s", d.user, d.pass, d.addr))
	// if e == nil {
	// 	d.db.SetMaxOpenConns(20)
	// 	d.db.SetMaxIdleConns(0)
	// 	d.db.SetConnMaxLifetime(time.Nanosecond)
	// }
	return
}

func (d *SDB) queryName(name string) (s []DBRecord, e error) {
	query := "SELECT id_student,identification,name," +
		"middle_name,last_name,address,phone,student_status_fk," +
		"faculty_fk,career_fk FROM student WHERE " +
		"name LIKE '%" + name + "%' " +
		"OR middle_name LIKE '%" + name + "%' " +
		"OR last_name LIKE '%" + name + "%'"
	var r *sql.Rows
	if e == nil {
		s = make([]DBRecord, NamePgLn)
		r, e = d.db.Query(query)
	}
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
	if e == nil {
		r, e = d.db.Query(query)
	}
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
	r, e = d.db.Query(query)
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
		s, e := d.db.Query(qr)
		if e == nil {
			if s.Next() {
				e = s.Scan(&r[i])
			} else {
				e = fmt.Errorf("Información no encontrada")
			}
		}
		if s != nil {
			s.Close()
		}
		if e == nil {
			e = s.Err()
		}
	}
	var s *sql.Rows
	if e == nil {
		f.status = r[0].String
		f.faculty = r[1].String
		if r[2].Valid {
			ncq := "SELECT name FROM national_career" +
				" WHERE id_national_career = '" + r[2].String + "'"
			s, e = d.db.Query(ncq)
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
	if s != nil {
		s.Close()
	}
	if e == nil {
		e = s.Err()
	}
	return
}
