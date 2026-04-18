package memory

import (
	"context"
	"sync"

	"github.com/vizurth/url_shortener/internal/storage"
)

//nolint:govet // fieldalignment: maps grouped with mutex for readability
type MemoryStorage struct {
	mu         sync.RWMutex
	byShort    map[string]string
	byOriginal map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		byShort:    make(map[string]string),
		byOriginal: make(map[string]string),
	}
}

func (s *MemoryStorage) Save(ctx context.Context, originalURL, shortCode string) (code string, isNew bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if code, ok := s.byOriginal[originalURL]; ok {
		return code, false, nil
	}

	if _, exists := s.byShort[shortCode]; exists {
		return "", false, storage.ErrShortCodeConflict
	}

	s.byShort[shortCode] = originalURL
	s.byOriginal[originalURL] = shortCode
	return shortCode, true, nil
}

func (s *MemoryStorage) Resolve(ctx context.Context, shortCode string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	originalURL, exists := s.byShort[shortCode]
	if !exists {
		return "", storage.ErrNotFound
	}
	return originalURL, nil
}

func (s *MemoryStorage) Close() error {
	return nil
}
