package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/kireetivar/go-tinyurl/internal/handlers"
	"github.com/kireetivar/go-tinyurl/internal/storage"
)


func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, loading from environment")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("FATAL: DATABASE_URL environment variable is not set")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("FATAL: JWT_SECRET environment variable is not set")
	}

	db, err := storage.NewDB(dbURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	urlStore,err := storage.NewSQLiteStore(db)
	if err != nil {
		log.Fatalf("Could not create URL store: %v", err)
	}
	userStore,err := storage.NewSQLiteUserStore(db)
	if err != nil {
		log.Fatalf("Could not create User store: %v", err)
	}
	urlCache := storage.NewCachingStore(urlStore)

	var s = handlers.NewServer(urlCache, userStore, jwtSecret)
	mux := s.Routes()

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}


