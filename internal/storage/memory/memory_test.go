package memory_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vizurth/url_shortener/internal/storage"
	"github.com/vizurth/url_shortener/internal/storage/memory"
)

func TestMemoryStorageNotFound(t *testing.T) {
	t.Parallel()
	s := memory.NewMemoryStorage()

	_, err := s.Resolve(context.Background(), "notexists1")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestMemoryStorageDuplicate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := memory.NewMemoryStorage()

	shortCode := "abc1234567"
	originalURL := "https://example.com"

	code, inserted, err := s.Save(ctx, originalURL, shortCode)
	require.NoError(t, err)
	assert.True(t, inserted)
	assert.Equal(t, shortCode, code)

	code, inserted, err = s.Save(ctx, originalURL, shortCode)
	require.NoError(t, err)
	assert.False(t, inserted)
	assert.Equal(t, shortCode, code)
}

func TestMemoryStorageSaveAndGet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := memory.NewMemoryStorage()

	shortCode := "abc1234567"
	originalURL := "https://example.com"

	_, _, err := s.Save(ctx, originalURL, shortCode)
	require.NoError(t, err)

	retrievedURL, err := s.Resolve(ctx, shortCode)
	require.NoError(t, err)
	assert.Equal(t, originalURL, retrievedURL)
}

func TestMemoryStorage_Concurrency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := memory.NewMemoryStorage()

	originalURL := "https://example.com"
	shortCode := "abc1234567"

	const numGoroutines = 50
	var insertedCount atomic.Int32
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for range numGoroutines {
		go func() {
			defer wg.Done()
			_, inserted, err := s.Save(ctx, originalURL, shortCode)
			assert.NoError(t, err)
			if inserted {
				insertedCount.Add(1)
			}

			got, err := s.Resolve(ctx, shortCode)
			assert.NoError(t, err)
			assert.Equal(t, originalURL, got)
		}()
	}

	wg.Wait()
	// только одна горут��на должна вставить запись
	assert.Equal(t, int32(1), insertedCount.Load())
}
