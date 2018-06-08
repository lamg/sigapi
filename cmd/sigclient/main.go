package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	h "net/http"
)

func main() {
	var addr, id, name, in string
	var offset, size int
	var pag bool
	flag.StringVar(&addr, "a", "", "sigapi server address")
	flag.BoolVar(&pag, "pg", false, "sigapi paginator")
	flag.IntVar(&offset, "o", 0, "page offset")
	flag.IntVar(&size, "sz", 10, "page size")
	flag.StringVar(&id, "id", "", "sigenu id")
	flag.StringVar(&name, "n", "", "sigenu name for filtering")
	flag.StringVar(&in, "in", "",
		"User identity number for getting evaluations")
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
	} else if id != "" {
		r, e = h.Get(fmt.Sprintf("%s/sigenu-id/%s", addr, id))
		printBody(r, e)
	} else if name != "" {
		r, e = h.Get(fmt.Sprintf("%s/sigenu-nombre/%s", addr, name))
		printBody(r, e)
	} else if in != "" {
		r, e = h.Get(fmt.Sprintf("%s/eval/%s", addr, in))
		printBody(r, e)
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
		fmt.Println(string(body))
	}
}
