package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	var addr, user, pass, srv string
	flag.StringVar(&addr, "a", "", "Dirección del servidor")
	flag.StringVar(&user, "u", "", "Usuario")
	flag.StringVar(&pass, "p", "", "Contraseña")
	flag.StringVar(&srv, "s", ":8080",
		"Dirección para servir la API")
	flag.Parse()
	dh, e := NewPostgreSDB(addr, user, pass)
	if e == nil {
		s := &http.Server{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  0,
			Addr:         srv,
			Handler:      dh,
		}
		e = s.ListenAndServe()
	}
	if e != nil {
		fmt.Fprintf(os.Stderr, e.Error())
		os.Exit(1)
	}
}
