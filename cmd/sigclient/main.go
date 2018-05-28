package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/lamg/sigapi"
	"io/ioutil"
	"log"
	h "net/http"
)

func main() {
	var addr, id, name, evals, user, pass string
	var offset, size int
	var pag, auth bool
	flag.StringVar(&addr, "a", "", "sigapi server address")
	flag.BoolVar(&pag, "pg", false, "sigapi paginator")
	flag.IntVar(&offset, "o", 0, "page offset")
	flag.IntVar(&size, "sz", 10, "page size")
	flag.StringVar(&id, "id", "", "sigenu id")
	flag.StringVar(&name, "n", "", "sigenu name for filtering")
	flag.BoolVar(&auth, "au", false, "authenticate user")
	flag.StringVar(&user, "u", "", "user name")
	flag.StringVar(&pass, "p", "", "password")
	flag.StringVar(&evals, "e", "",
		"JWT sent by auth to get user evaluations")
	flag.Parse()
	tr := &h.Transport{
		Proxy: nil,
	}
	h.DefaultClient.Transport = tr
	var e error
	var r *h.Response
	if pag {
		r, e = h.Get(fmt.Sprintf("%s/paginador/%d/%d", addr, offset, size))
		printBody(r, e)
	} else if auth {
		c := &sigapi.Credentials{
			User: user,
			Pass: pass,
		}
		var bs []byte
		bs, e = json.Marshal(c)
		var rq *h.Request
		if e == nil {
			bf := bytes.NewReader(bs)
			rq, e = h.NewRequest(h.MethodPost, addr+"/auth", bf)
		}
		if e == nil {
			r, e = h.DefaultClient.Do(rq)
		}
		printBody(r, e)
	} else if id != "" {
		r, e = h.Get(fmt.Sprintf("%s/sigenu-id/%s", addr, id))
		printBody(r, e)
	} else if name != "" {
		r, e = h.Get(fmt.Sprintf("%s/sigenu-nombre/%s", addr, name))
		printBody(r, e)
	} else if evals != "" {
		var q *h.Request
		q, e = h.NewRequest(h.MethodGet, addr+"/eval", nil)
		if e == nil {
			q.Header.Set(sigapi.AuthHd, evals)
			r, e = h.DefaultClient.Do(q)
			printBody(r, e)
		}
	}
	if e != nil {
		if r != nil {
			log.Fatalf("error: %s code: %d", e.Error(), r.StatusCode)
		} else {
			log.Fatalf("error: %s", e.Error())
		}
	}
}

func printBody(r *h.Response, e error) {
	var body []byte
	if e == nil {
		body, e = ioutil.ReadAll(r.Body)
		r.Body.Close()
	}
	if e == nil {
		fmt.Print(string(body))
	}
}
