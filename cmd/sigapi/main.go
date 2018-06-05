package main

import (
	"flag"
	"fmt"
	"github.com/lamg/sigapi"
	"net/http"
	"os"
	"time"
)

func main() {
	var addr, user, pass, srv, tmpl string
	flag.StringVar(&addr, "a", "", "Dirección del servidor")
	flag.StringVar(&user, "u", "", "Usuario")
	flag.StringVar(&pass, "p", "", "Contraseña")
	flag.StringVar(&srv, "s", ":8080",
		"Dirección para servir la API")
	flag.StringVar(&tmpl, "l", "",
		"Camino de la plantilla de la documentación")
	flag.Parse()
	dh, e := sigapi.NewPostgreSDB(addr, user, pass, tmpl)
	if e == nil {
		s := &http.Server{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  0,
			Addr:         srv,
			Handler:      dh.Handler,
		}
		e = s.ListenAndServe()
	}
	if e != nil {
		fmt.Fprintf(os.Stderr, "%s\n", e.Error())
		os.Exit(1)
	}
}
