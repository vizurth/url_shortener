package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vizurth/url_shortener/pkg/logger"
	psg "github.com/vizurth/url_shortener/pkg/postgres"
)

func TestNewPostgresStorage(t *testing.T) {
	log, err := logger.New()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	ctx := logger.With(context.Background(), log)

	cfg := psg.Config{
		Host:     "localhost",
		Port:     "5432",
		Username: "short",
		Password: "short",
		Database: "short",
		MaxConns: 10,
		MinConns: 2,
	}

	storage, err := NewPostgresStorage(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, storage)

	err = storage.Close()
	require.NoError(t, err)
}

func TestSave(t *testing.T) {
	log, err := logger.New()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	ctx := logger.With(context.Background(), log)
	storage := setupTestStorage(t, ctx)
	defer storage.Close()

	tests := []struct {
		name        string
		originalURL string
		shortCode   string
		expectNew   bool
	}{
		{
			name:        "save new url",
			originalURL: "https://example.com",
			shortCode:   "abc1234567",
			expectNew:   true,
		},
		{
			name:        "save existing url",
			originalURL: "https://example.com",
			shortCode:   "xyz7891234",
			expectNew:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returnedCode, isNew, err := storage.Save(ctx, tt.originalURL, tt.shortCode)
			require.NoError(t, err)
			require.Equal(t, tt.expectNew, isNew)

			if tt.expectNew {
				require.Equal(t, tt.shortCode, returnedCode)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	log, err := logger.New()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	ctx := logger.With(context.Background(), log)
	storage := setupTestStorage(t, ctx)
	defer storage.Close()

	originalURL := "https://example.com/test"
	shortCode := "test123456"

	_, _, err = storage.Save(ctx, originalURL, shortCode)
	require.NoError(t, err)

	resolved, err := storage.Resolve(ctx, shortCode)
	require.NoError(t, err)
	require.Equal(t, originalURL, resolved)
}

func TestResolveNotFound(t *testing.T) {
	log, err := logger.New()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	ctx := logger.With(context.Background(), log)
	storage := setupTestStorage(t, ctx)
	defer storage.Close()

	_, err = storage.Resolve(ctx, "nonexistent")
	require.Error(t, err)
}

func setupTestStorage(t *testing.T, ctx context.Context) *PostgresStorage {
	cfg := psg.Config{
		Host:     "localhost",
		Port:     "5432",
		Username: "short",
		Password: "short",
		Database: "short",
		MaxConns: 10,
		MinConns: 2,
	}

	storage, err := NewPostgresStorage(ctx, cfg)
	require.NoError(t, err)

	return storage
}
