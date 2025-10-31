package storage

import (
	"database/sql"
	"errors"
	"log"

	"github.com/kireetivar/go-tinyurl/pkg/utils"
	"github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	db *sql.DB
}

func (s *SQLiteStore) Save(longUrl string) (string, error) {
	saveKeySQL := `INSERT INTO urls (short_key, long_url) VALUES (?, ?)`

	for i := 0; i < 5; i++ {
		shortKey := utils.GenerateShortKey()

		_, err := s.db.Exec(saveKeySQL, shortKey, longUrl)

		if err == nil {
			return shortKey, nil
		}

		// Check if it was a "UNIQUE constraint" error
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			// This is the fix: check ExtendedCode
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				// It was a duplicate key. Loop will try again.
				continue
			}
		}

		// It was some other, unexpected database error
		log.Printf("ERROR saving url in DB: %v\n", err)
		return "", err
	}

	return "", errors.New("failed to generate unique short key after 5 attempts")
}

func (s *SQLiteStore) Get(shortKey string) (string,error) {
	dbGet := `SELECT long_url FROM urls WHERE short_key = ?`

	var longURL string
	err := s.db.QueryRow(dbGet,shortKey).Scan(&longURL)

	if err != nil {
        if err == sql.ErrNoRows {
            return "", errors.New("key not found")
        }
        log.Printf("ERROR while fetching url: %v\n", err)
        return "", err
    }

	return longURL,nil
}


func NewSQLiteStore(dataSourceName string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	statement := `CREATE TABLE IF NOT EXISTS urls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    short_key TEXT NOT NULL UNIQUE,
    long_url TEXT NOT NULL
	);`

	_, err = db.Exec(statement)
	if err != nil {
		return nil, err
	}

	return &SQLiteStore{
		db: db,
	}, nil
}