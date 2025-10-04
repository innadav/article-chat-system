package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

// Service provides a simple in-memory cache for request hashing.
type Service struct {
	store sync.Map
}

func NewService() *Service {
	return &Service{}
}

// GenerateCacheKey now creates a stable hash from a simple string.
func (s *Service) GenerateCacheKey(query string) string {
	hash := sha256.Sum256([]byte(query))
	return hex.EncodeToString(hash[:])
}

func (s *Service) Get(key string) (string, bool) {
	val, ok := s.store.Load(key)
	if !ok {
		return "", false
	}
	return val.(string), true
}

func (s *Service) Set(key, value string) {
	s.store.Store(key, value)
}
