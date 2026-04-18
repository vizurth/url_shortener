package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vizurth/url_shortener/internal/storage"
	"github.com/vizurth/url_shortener/mocks"
)

func TestService_ShortenURL_RetriesOnShortCodeConflict(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(t)
	svc := New(mockStorage)

	originalURL := "https://example.com/retry"
	returnedCode := "abc1234567"

	mockStorage.EXPECT().
		Save(ctx, originalURL, anyCode()).
		Return("", false, storage.ErrShortCodeConflict).
		Once()

	mockStorage.EXPECT().
		Save(ctx, originalURL, anyCode()).
		Return(returnedCode, true, nil).
		Once()

	gotCode, isNew, err := svc.ShortenURL(ctx, originalURL)
	assert.NoError(t, err)
	assert.True(t, isNew)
	assert.Equal(t, returnedCode, gotCode)
}

func anyCode() any {
	return mock.MatchedBy(func(s string) bool {
		if len(s) != shortCodeLength {
			return false
		}
		for _, c := range s {
			if !strings.ContainsRune(charset, c) {
				return false
			}
		}
		return true
	})
}

func TestService_ShortenURL_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(t)
	svc := New(mockStorage)

	originalURL := "https://example.com/path"
	returnedCode := "abc1234567"

	mockStorage.EXPECT().
		Save(ctx, originalURL, anyCode()).
		Return(returnedCode, true, nil)

	gotCode, isNew, err := svc.ShortenURL(ctx, originalURL)
	assert.NoError(t, err)
	assert.True(t, isNew)
	assert.Equal(t, returnedCode, gotCode)
}

func TestService_ShortenURL_InvalidURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := New(mocks.NewMockStorage(t))

	_, _, err := svc.ShortenURL(ctx, "not a url")
	assert.ErrorIs(t, err, ErrInvalidURL)
}

func TestService_ShortenURL_SaveError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(t)
	svc := New(mockStorage)

	originalURL := "https://example.com/error"
	saveErr := errors.New("db down")

	mockStorage.EXPECT().
		Save(ctx, originalURL, anyCode()).
		Return("", false, saveErr)

	gotCode, isNew, err := svc.ShortenURL(ctx, originalURL)
	assert.Error(t, err)
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

func TestService_Resolve_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockStorage := mocks.NewMockStorage(t)
	svc := New(mockStorage)

	// storage возвращает ErrNotFound — сервис должен транслировать в service.ErrNotFound
	mockStorage.EXPECT().
		Resolve(ctx, "abc123XYZ0").
		Return("", storage.ErrNotFound)

	gotURL, err := svc.Resolve(ctx, "abc123XYZ0")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Empty(t, gotURL)
}

func TestService_Resolve_InvalidCodeLength(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := New(mocks.NewMockStorage(t))

	// 11 символов — валидация отклонит до вызова storage
	_, err := svc.Resolve(ctx, "toolongcode1")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestGenerateShortCode_LengthAndCharset(t *testing.T) {
	t.Parallel()

	for range 20 {
		code, err := generateShortCode()
		require.NoError(t, err)
		assert.Len(t, code, shortCodeLength)
		for _, c := range code {
			assert.True(t, strings.ContainsRune(charset, c), "unexpected char %q in code %q", c, code)
		}
	}
}

func TestGenerateShortCode_Randomness(t *testing.T) {
	t.Parallel()

	seen := make(map[string]struct{}, 100)
	for range 100 {
		code, err := generateShortCode()
		require.NoError(t, err)
		seen[code] = struct{}{}
	}
	assert.Len(t, seen, 100)
}
