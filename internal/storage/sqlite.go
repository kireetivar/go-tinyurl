package storage

import (
	"database/sql"
	"errors"
	"log"
	"os"

	"github.com/kireetivar/go-tinyurl/internal/models"
	"github.com/kireetivar/go-tinyurl/pkg/utils"
	"github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	db *sql.DB
}

type sqliteUserStore struct {
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

func (s *SQLiteStore) Get(shortKey string) (string, error) {
	dbGet := `SELECT long_url FROM urls WHERE short_key = ?`

	var longURL string
	err := s.db.QueryRow(dbGet, shortKey).Scan(&longURL)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("key not found")
		}
		log.Printf("ERROR while fetching url: %v\n", err)
		return "", err
	}

	return longURL, nil
}

func (s *sqliteUserStore) Create(username, email, hashedPassword string) (int64, error) {
	userCreate := `INSERT into users (username,email, hashed_password) VALUES (?,?,?)`

	res, err := s.db.Exec(userCreate, username, email, hashedPassword)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return 0, err // Return 0 and the error
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Panicf("Failed to get last insert ID: %v", err)
		return 0, err
	}

	return id, nil
}

func (s *sqliteUserStore) GetByUsername(username string) (*models.User, error) {
	getUser := `SELECT userId, username, email, hashed_password FROM users WHERE username = ?`

	var user models.User
	err := s.db.QueryRow(getUser, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.HashedPassword,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		log.Printf("ERROR while fetching user %s: %v", username, err)
		return nil, err
	}
	return &user, nil
}

func NewSQLiteStore(db *sql.DB) (Storage, error) {
	urlsTableStmt := `CREATE TABLE IF NOT EXISTS urls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    short_key TEXT NOT NULL UNIQUE,
    long_url TEXT NOT NULL
	);`

	_, err := db.Exec(urlsTableStmt)
	if err != nil {
		return nil, err
	}

	return &SQLiteStore{
		db: db,
	}, nil
}

func NewSQLiteUserStore(db *sql.DB) (UserStorage, error) {
	usersTableStmt := `CREATE TABLE IF NOT EXISTS users (
    userId INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL
    );`
	_, err := db.Exec(usersTableStmt)
	if err != nil {
		return nil, err
	}

	return &sqliteUserStore{
		db: db,
	}, nil
}

func NewDB(dataSourceName string) (*sql.DB, error) {
	// Create parent directory if it doesn't exist (for file-based SQLite)
	if dataSourceName != ":memory:" && dataSourceName != "" {
		dir := dataSourceName[:len(dataSourceName)-len(dataSourceName[len(dataSourceName)-1:])]
		// Find the last slash to get directory
		for i := len(dataSourceName) - 1; i >= 0; i-- {
			if dataSourceName[i] == '/' || dataSourceName[i] == '\\' {
				dir = dataSourceName[:i]
				break
			}
		}
		// Only try to create directory if there's a path component
		if dir != "" && dir != dataSourceName {
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Printf("Warning: could not create directory %s: %v", dir, err)
			}
		}
	}

	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
