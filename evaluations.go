package sigapi

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	h "net/http"
	"sort"
)

const (
	EmployeeID = "employeeID"
)

func (d *SDB) evaluationsHn(w h.ResponseWriter, r *h.Request) {
	rg := mux.Vars(r)
	ci := rg[IN]
	gs, e := d.queryEvl(ci)
	if e == nil {
		ev := sortAndRemDup(gs)
		e = Encode(w, ev)
	}
	writeErr(w, e)
}

type SubjEval struct {
	Subject string `json:"subject"`
	Eval    string `json:"eval"`
}

func NoEmployeeIDField(user string) (e error) {
	e = fmt.Errorf("No %s field found for %s", EmployeeID, user)
	return
}

type EvYear struct {
	Year    string     `json:"year"`
	Periods []EvPeriod `json:"periods"`
}

type EvPeriod struct {
	Period string     `json:"period"`
	Evs    []SubjEval `json:"evs"`
}

type ByYearPeriod []StudentEvl

func (b ByYearPeriod) Len() (n int) {
	n = len(b)
	return
}

func (b ByYearPeriod) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b ByYearPeriod) Less(i, j int) (r bool) {
	r = (b[i].Year == b[j].Year && b[i].Period < b[j].Period) ||
		b[i].Year < b[j].Year
	return
}

func sortAndRemDup(ev []StudentEvl) (ys []EvYear) {
	sort.Sort(ByYearPeriod(ev))
	cy, cp, cyi, cpi := "", "", -1, -1 //current year, current period,
	// current year index, current period index
	ys = make([]EvYear, 0)
	for _, j := range ev {
		if j.Year != cy {
			ny := EvYear{
				Year:    j.Year,
				Periods: make([]EvPeriod, 0),
			}
			ys = append(ys, ny)
			cy, cyi = j.Year, cyi+1
			cp, cpi = "", -1
		}
		if j.Period != cp {
			np := EvPeriod{
				Period: j.Period,
				Evs:    make([]SubjEval, 0),
			}
			ys[len(ys)-1].Periods = append(ys[len(ys)-1].Periods, np)
			cp, cpi = j.Period, cpi+1
		}
		nv := SubjEval{
			Eval:    j.EvalValue,
			Subject: j.SubjectName,
		}
		ok := canUpdate(ys[cyi].Periods[cpi].Evs, nv)
		if !ok {
			ys[cyi].Periods[cpi].Evs = append(ys[cyi].Periods[cpi].Evs, nv)
		}
	}

	return
}

func canUpdate(a []SubjEval, v SubjEval) (ok bool) {
	ok = false
	f, i := false, 0 //f: found
	for !f && i != len(a) {
		f = a[i].Subject == v.Subject
		if !f {
			i = i + 1
		}
	}
	if f && v.Eval > a[i].Eval {
		a[i] = v
	}
	ok = f
	return
}

type StudentEvl struct {
	SubjectName string `json:"subjectName"`
	EvalValue   string `json:"evalValue"`
	Period      string `json:"period"`
	Year        string `json:"year"`
}

func (d *SDB) queryEvl(idStudent string) (es []StudentEvl, e error) {
	query := fmt.Sprintf("SELECT id_student FROM student WHERE "+
		" identification = '%s'", idStudent)
	var r *sql.Rows
	if e == nil {
		r, e = d.db.Query(query)
	}
	var studDBId string
	ok := r.Next()
	if e == nil && ok {
		e = r.Scan(&studDBId)
	}
	if r != nil {
		r.Close()
	}
	if e == nil {
		query = fmt.Sprintf(
			"SELECT evaluation_value_fk,matriculated_subject_fk "+
				" FROM evaluation WHERE student_fk = '%s'", studDBId)
		r, e = d.db.Query(query)
	}
	evalValId, matSubjId := make([]string, 0), make([]string, 0)
	for i := 0; e == nil && r.Next(); i++ {
		var ev, ms sql.NullString
		e = r.Scan(&ev, &ms)
		if e == nil && ev.Valid && ms.Valid {
			evalValId, matSubjId = append(evalValId, ev.String),
				append(matSubjId, ms.String)
		}
	}
	if r != nil {
		r.Close()
	}
	evalVal := make([]string, 0)
	for i := 0; e == nil && i != len(evalValId); i++ {
		query = fmt.Sprintf("SELECT value FROM evaluation_value WHERE "+
			"id_evaluation_value = '%s'", evalValId[i])
		// print("query evalValId: ")
		// println(query)
		r, e = d.db.Query(query)
		var ev string
		if e == nil && r.Next() {
			e = r.Scan(&ev)
		}
		if e == nil {
			evalVal = append(evalVal, ev)
		}
		if r != nil {
			r.Close()
		}
	}
	subjId := make([]string, 0)
	for i := 0; e == nil && i != len(matSubjId); i++ {
		query = fmt.Sprintf(
			"SELECT subject_fk FROM matriculated_subject WHERE "+
				"matriculated_subject_id = '%s'", matSubjId[i])
		r, e = d.db.Query(query)
		var si string
		if e == nil && r.Next() {
			e = r.Scan(&si)
		}
		if e == nil {
			subjId = append(subjId, si)
		}
		if r != nil {
			r.Close()
		}
	}
	subjNameId, subjPeriod, subjYear := make([]string, 0),
		make([]string, 0), make([]string, 0)
	for i := 0; e == nil && i != len(subjId); i++ {
		query = fmt.Sprintf(
			"SELECT subject_name_fk, period, year FROM subject WHERE "+
				"subject_id = '%s'", subjId[i])
		r, e = d.db.Query(query)
		var sni, period, year string
		if e == nil && r.Next() {
			e = r.Scan(&sni, &period, &year)
		}
		if e == nil {
			subjNameId, subjPeriod, subjYear =
				append(subjNameId, sni),
				append(subjPeriod, period),
				append(subjYear, year)
		}
		if r != nil {
			r.Close()
		}
	}
	subjName := make([]string, 0)
	for i := 0; e == nil && i != len(subjNameId); i++ {
		query = fmt.Sprintf("SELECT name FROM subject_name WHERE "+
			"subject_name_id = '%s'", subjNameId[i])
		r, e = d.db.Query(query)
		var sn string
		if e == nil && r.Next() {
			e = r.Scan(&sn)
		}
		if e == nil {
			subjName = append(subjName, sn)
		}
		if r != nil {
			r.Close()
		}
	}
	if r != nil {
		r.Close()
	}
	es = make([]StudentEvl, len(subjName))
	for i := 0; e == nil && i != len(subjName); i++ {
		es[i] = StudentEvl{
			SubjectName: subjName[i],
			Period:      subjPeriod[i],
			Year:        subjYear[i],
			EvalValue:   evalVal[i],
		}
	}
	return
}
