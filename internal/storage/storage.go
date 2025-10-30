package storage

type Storage interface {
	Save(longURL string) (string, error)
	Get(shortKey string) (string, error)
}