package storage

import (
	"sync"
)

type CachingStore struct {
	cache map[string]string
	mu    sync .RWMutex
	next  Storage
}


func NewCachingStore(next Storage) (*CachingStore) {

	return &CachingStore{
		cache: make(map[string]string),
		mu: sync.RWMutex{},
		next: next,
	}
}

func (c *CachingStore) Get(key string) (string,error) {
	c.mu.RLock()
	url ,ok := c.cache[key]
	c.mu.RUnlock()

	if ok {
		return url, nil
	}

	url , err := c.next.Get(key)
	if err != nil {
		return "",err
	}

	c.mu.Lock()
	c.cache[key] = url
	c.mu.Unlock()

	return url,nil
}

func (c *CachingStore) Save(LongUrl string) (string,error) {
	key,err:=c.next.Save(LongUrl)
	if err != nil {
		return "",err
	}
	c.mu.Lock()
	c.cache[key] = LongUrl
	c.mu.Unlock()
	return key,nil
}