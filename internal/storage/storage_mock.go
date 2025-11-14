package storage

import (
	"github.com/kireetivar/go-tinyurl/internal/models"
)

type MockStore struct {
	WantGetURL string
	WantGetErr error

	WantSaveURL string
	WantSaveErr error
}

func (m *MockStore) Save(longURL string) (string, error) {
	return m.WantSaveURL, m.WantSaveErr
}

func (m *MockStore) Get(shortKey string) (string, error) {
	return m.WantGetURL, m.WantGetErr
}

type MockUserStore struct {
	WantUser *models.User
	WantErr  error
	WantID   int64
}

func (m *MockUserStore) Create(username, email, hash string) (int64,error) {
	return m.WantID, m.WantErr
}

func  (m *MockUserStore) GetByUsername(username string) (*models.User, error) {
	return m.WantUser, m.WantErr
}