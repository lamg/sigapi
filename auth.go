package main

import (
	"io/ioutil"
	h "net/http"
)

func (d *SDB) authHn(w h.ResponseWriter, r *h.Request) {
	c, e := credentials(r)
	if e == nil {
		e = d.Ld.Authenticate(c.User, c.Pass)
	}
	var s string
	if e == nil {
		s, e = encrypt(c)
	}
	if e == nil {
		w.Write([]byte(s))
	}
	writeErr(w, e)
}
