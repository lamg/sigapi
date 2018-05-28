package sigapi

import (
	h "net/http"
)

func (d *SDB) authHn(w h.ResponseWriter, r *h.Request) {
	c, e := credentials(r)
	if e == nil {
		e = d.Ld.Authenticate(c.User, c.Pass)
	}
	var s string
	if e == nil {
		s, e = d.cr.encrypt(c)
	}
	if e == nil {
		w.Write([]byte(s))
	}
	writeErr(w, e)
}
