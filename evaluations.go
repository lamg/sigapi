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
	if e == nil {
		e = Encode(w, gs)
	}
	writeErr(w, e)
}

func NoEmployeeIDField(user string) (e error) {
	e = fmt.Errorf("No %s field found for %s", EmployeeID, user)
	return
}
