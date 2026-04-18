package memory

import (
	"context"
	"sync"

	"github.com/vizurth/url_shortener/internal/storage"
)

type MemoryStorage struct {
	byShort    map[string]string
	byOriginal map[string]string

	mu *sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		byShort:    make(map[string]string),
		byOriginal: make(map[string]string),
		mu:         &sync.RWMutex{},
	}
}

func (s *MemoryStorage) Save(ctx context.Context, originalURL, shortCode string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if code, ok := s.byOriginal[originalURL]; ok {
		return code, false, nil
	}

	s.byShort[shortCode] = originalURL
	s.byOriginal[originalURL] = shortCode
	return shortCode, true, nil
}

func (s *MemoryStorage) Resolve(ctx context.Context, shortCode string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	originalUrl, exists := s.byShort[shortCode]
	if !exists {
		return "", storage.ErrNotFound
	}
	return originalUrl, nil
}

func (s *MemoryStorage) Close() error {
	return nil
}
