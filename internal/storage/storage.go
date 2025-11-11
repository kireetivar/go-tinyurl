package storage

import "github.com/kireetivar/go-tinyurl/internal/models"

type Storage interface {
	Save(longURL string) (string, error)
	Get(shortKey string) (string, error)
}

type UserStorage interface {
	Create(username, email, hashedPassword string) (int64, error)
	GetByUsername(username string) (*models.User, error)
}