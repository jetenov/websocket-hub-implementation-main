package main

import (
	"context"
	"log"
	"net/http"
)

const addr = ":5000"

func main() {
	hub := NewHub()
	serverPool := ServerPool{
		Sources: []string{
			"https://novasite.su/test1.php",
			"https://novasite.su/test2.php",
		},
		Hub: hub,
	}
	srv := http.Server{
		Addr:    addr,
		Handler: nil,
	}

	go hub.Run()
	go func() {
		err := serverPool.Run()
		if err != nil {
			srv.Shutdown(context.Background())
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		WSServe(hub, w, r)
	})

	log.Println("Server started")
	log.Fatal(srv.ListenAndServe())
}
