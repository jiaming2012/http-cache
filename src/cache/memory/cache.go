package memory

import (
	log "github.com/jiaming2012/http-cache/src/logger"
	"sync"
	"time"
)

// CachedEntity is a cached reference
type CachedEntity struct {
	Content    []byte
	Expiration int64
}

// Expired returns true if the item has expired.
func (entity CachedEntity) Expired() bool {
	if entity.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > entity.Expiration
}

//Storage mechanism for caching strings in memory
type Storage struct {
	items map[string]CachedEntity
	mu    *sync.RWMutex
}

//NewStorage creates a new in memory storage
func NewStorage() *Storage {
	return &Storage{
		items: make(map[string]CachedEntity),
		mu:    &sync.RWMutex{},
	}
}

//Get a cached content by key
func (s Storage) Get(key string) []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item := s.items[key]
	if item.Expired() {
		log.Logger.Debugf("%v removed removed from cache for: expired", key)
		delete(s.items, key)
		return nil
	}
	return item.Content
}

//Set a cached content by key
func (s Storage) Set(key string, content []byte, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[key] = CachedEntity{
		Content:    content,
		Expiration: time.Now().Add(duration).UnixNano(),
	}
}
