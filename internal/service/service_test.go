package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vizurth/url_shortener/internal/storage"
	"github.com/vizurth/url_shortener/mocks"
)

func TestService_ShortenURL_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(t)
	svc := New(mockStorage)

	originalURL := "https://example.com/path"
	expectedCode := generateShortCode(originalURL)

	mockStorage.EXPECT().
		Save(ctx, expectedCode, originalURL).
		Return(expectedCode, true, nil)

	gotCode, isNew, err := svc.ShortenURL(ctx, originalURL)
	assert.NoError(t, err)
	assert.True(t, isNew)
	assert.Equal(t, expectedCode, gotCode)
}

func TestService_ShortenURL_SaveError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(t)
	svc := New(mockStorage)

	originalURL := "https://example.com/error"
	expectedCode := generateShortCode(originalURL)
	saveErr := errors.New("db down")

	mockStorage.EXPECT().
		Save(ctx, expectedCode, originalURL).
		Return("", false, saveErr)

	gotCode, isNew, err := svc.ShortenURL(ctx, originalURL)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "failed to save URL mapping")
	assert.ErrorIs(t, err, saveErr)
	assert.Empty(t, gotCode)
	assert.False(t, isNew)
}

func TestService_Resolve_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(t)
	svc := New(mockStorage)

	shortCode := "abc123XYZ9"
	originalURL := "https://example.com/ok"

	mockStorage.EXPECT().
		Resolve(ctx, shortCode).
		Return(originalURL, nil)

	gotURL, err := svc.Resolve(ctx, shortCode)
	assert.NoError(t, err)
	assert.Equal(t, originalURL, gotURL)
}

func TestService_Resolve_Error(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(t)
	svc := New(mockStorage)

	shortCode := "missing0001"

	mockStorage.EXPECT().
		Resolve(ctx, shortCode).
		Return("", storage.ErrNotFound)

	gotURL, err := svc.Resolve(ctx, shortCode)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "failed to resolve short code")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	assert.Empty(t, gotURL)
}

func TestGenerateShortCode_DeterministicAndLength(t *testing.T) {
	t.Parallel()

	url := "https://example.com/same"
	code1 := generateShortCode(url)
	code2 := generateShortCode(url)

	assert.Equal(t, code1, code2)
	assert.Len(t, code1, shortCodeLength)

	for i := 0; i < len(code1); i++ {
		assert.Contains(t, base62Alphabet, string(code1[i]))
	}
}
