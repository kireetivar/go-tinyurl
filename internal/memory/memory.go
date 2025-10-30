package memory

import (
	"errors"
	"sync"

	"github.com/kireetivar/go-tinyurl/pkg/utils"
)

type MemoryStore struct {
	InMemory map[string]string
	mut      sync.Mutex
}


func (m *MemoryStore) Save(longUrl string) (string,error) {

	shortKey := utils.GenerateShortKey()
	for {
		m.mut.Lock()
		_,ok:= m.InMemory[shortKey]
		if !ok {
			m.InMemory[shortKey] = longUrl
			m.mut.Unlock()
			break
		}
		m.mut.Unlock()
		shortKey = utils.GenerateShortKey()
	}
	return shortKey,nil
}


func (m *MemoryStore) Get(shortKey string) (string,error) {
	m.mut.Lock()
	url,ok := m.InMemory[shortKey]
	m.mut.Unlock()
	if !ok {
		return "",errors.New("key not found")
	}

	return url,nil
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		InMemory: make(map[string]string),
		mut: sync.Mutex{},
	}
}