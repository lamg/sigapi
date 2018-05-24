package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	h "net/http"
)

type Credentials struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

const (
	EmployeeID = "employeeID"
)

func (d *SDB) gradesHn(w h.ResponseWriter, r *h.Request) {
	bs, e := ioutil.ReadAll(r.Body)
	var c *Credentials
	if e == nil {
		r.Body.Close()
		c = new(Credentials)
		e = json.Unmarshal(bs, c)
	}
	var mp map[string][]string
	if e == nil {
		mp, e = d.Ld.FullRecord(c.User, c.Pass, c.User)
	}
	var ci string
	if e == nil {
		cia, ok := mp[EmployeeID]
		if ok && len(cia) != 0 {
			ci = cia[0]
		}
		if ci == "" {
			e = NoEmployeeIDField(c.User)
		}
	}
	var gs string
	if e == nil {
		gs, e = d.queryGrades(ci)
	}
	var gsbs []byte
	if e == nil {
		gsbs, e = json.Marshal(gs)
	}
	if e == nil {
		w.Write(gsbs)
	}
	writeErr(w, e)
}

func NoEmployeeIDField(user string) (e error) {
	e = fmt.Errorf("No %s field found for %s", EmployeeID, user)
	return
}
