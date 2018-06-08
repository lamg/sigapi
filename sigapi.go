package sigapi

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"html/template"
	h "net/http"
	"sync"
)

type SDB struct {
	db       *sql.DB
	Handler  h.Handler
	tp       *template.Template
	user     string
	pass     string
	addr     string
	pagPath  string
	idPath   string
	namePath string
	evalPath string
	once     sync.Once
}

const (
	Offset   = "offset"
	Size     = "size"
	Id       = "id"
	Name     = "nombre"
	IN       = "identity"
	NamePgLn = 100
)

func NewPostgreSDB(addr, user, pass, tpf string) (d *SDB, e error) {
	var tp *template.Template
	tp, e = template.New("doc").ParseFiles(tpf)
	if e == nil {
		d = &SDB{tp: tp, user: user, pass: pass, addr: addr}
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
		e = d.openDB()
	}
	return
}

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
