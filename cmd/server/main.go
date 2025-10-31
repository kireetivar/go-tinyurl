package main

import (
	"log"
	"net/http"

	"github.com/kireetivar/go-tinyurl/internal/handlers"
	"github.com/kireetivar/go-tinyurl/internal/storage"
)


func main() {
	store, err := storage.NewSQLiteStore("./shortener.db")
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	var s = handlers.NewServer(store)

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


