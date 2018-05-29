package sigapi

import (
	"fmt"
	h "net/http"
	"strings"
)

const (
	EmployeeID = "employeeID"
)

func (d *SDB) evaluationsHn(w h.ResponseWriter, r *h.Request) {
	c, e := d.cr.decrypt(r)
	var mp map[string][]string
	if e == nil {
		mp, e = d.Ld.FullRecord(c.User, c.Pass, c.User)
	}
	var ci string
	if e == nil {
		cia, ok := mp[EmployeeID]
		if ok && len(cia) != 0 {
			ci = strings.TrimSpace(cia[0])
		}
		if ci == "" {
			e = NoEmployeeIDField(c.User)
		}
	}
	var gs []StudentEvl
	if e == nil {
		gs, e = d.queryEvl(ci)
	}
	subMp := make(map[string]map[string][]SubjEval)
	if e == nil {
		for _, j := range gs {
			_, ok := subMp[j.Year]
			if !ok {
				subMp[j.Year] = make(map[string][]SubjEval)
			}
			_, ok = subMp[j.Year][j.Period]
			if !ok {
				subMp[j.Year][j.Period] = make([]SubjEval, 0)
			}
			subMp[j.Year][j.Period] = append(subMp[j.Year][j.Period],
				SubjEval{
					Subject: j.SubjectName,
					Eval:    j.EvalValue,
				},
			)
		}
		e = Encode(w, subMp)
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
