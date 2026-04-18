package memory_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vizurth/url_shortener/internal/storage"
	"github.com/vizurth/url_shortener/internal/storage/memory"
)

func TestMemoryStorageNotFound(t *testing.T) {
	s := memory.NewMemoryStorage()

	_, err := s.Resolve(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestMemoryStorageDuplicate(t *testing.T) {
	ctx := context.Background()
	s := memory.NewMemoryStorage()

	shortCode := "abc123"
	originalURL := "https://example.com"

	code, inserted, err := s.Save(ctx, originalURL, shortCode)
	assert.Equal(t, shortCode, code)
	assert.True(t, inserted)
	assert.NoError(t, err)

	code, inserted, err = s.Save(ctx, originalURL, shortCode)
	assert.Equal(t, shortCode, code)
	assert.False(t, inserted)
	assert.NoError(t, err)
}

func TestMemoryStorageSaveAndGet(t *testing.T) {
	ctx := context.Background()
	s := memory.NewMemoryStorage()

	shortCode := "abc123"
	originalURL := "https://example.com"

	code, inserted, err := s.Save(ctx, originalURL, shortCode)
	assert.Equal(t, shortCode, code)
	assert.True(t, inserted)
	assert.NoError(t, err)

	retrievedURL, err := s.Resolve(ctx, shortCode)
	assert.NoError(t, err)
	assert.Equal(t, originalURL, retrievedURL)
}
