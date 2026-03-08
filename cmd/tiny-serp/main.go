package main

import (
	"log"
	"net/http"
	"os"

	tinyserp "github.com/okayama-daiki/tiny-serp"
	"github.com/okayama-daiki/tiny-serp/httpapi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("tiny-serp listening on %s", addr)
	if err := http.ListenAndServe(addr, httpapi.NewHandler(tinyserp.NewService(nil))); err != nil {
		log.Fatal(err)
	}
}
