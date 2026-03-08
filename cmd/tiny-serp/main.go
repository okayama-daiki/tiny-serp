package main

import (
	"log"
	"net/http"
	"os"

	"github.com/okayama-daiki/tiny-serp/httpapi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("tiny-serp listening on %s", addr)
	if err := http.ListenAndServe(addr, httpapi.NewHandler(nil, nil)); err != nil {
		log.Fatal(err)
	}
}
