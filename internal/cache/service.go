package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"

	"article-chat-system/internal/planner"
)

// Service provides a simple in-memory cache for request hashing.
type Service struct {
	store sync.Map
}

func NewService() *Service {
	return &Service{}
}

// GenerateCacheKey creates a stable, unique hash for a given query plan.
func (s *Service) GenerateCacheKey(plan *planner.QueryPlan) string {
	// Sort targets to ensure the key is stable regardless of order.
	sort.Strings(plan.Targets)
	keyData := fmt.Sprintf("%s:%v:%v", plan.Intent, plan.Targets, plan.Parameters)
	hash := sha256.Sum256([]byte(keyData))
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
