package main

import (
	"log"
	"net/http"

	"github.com/kireetivar/go-tinyurl/internal/handlers"
)


func main() {
	var s = handlers.NewServer()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /shorten", s.ShortenHandler)
	mux.HandleFunc("GET /{shortKey}", s.RedirectHandler)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}


