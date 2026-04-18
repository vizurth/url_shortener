package service

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/vizurth/url_shortener/internal/storage"
)

const (
	base62Alphabet  = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	shortCodeLength = 10
)

type URLService interface {
	ShortenURL(ctx context.Context, originalURL string) (shortCode string, isNew bool, err error)
	Resolve(ctx context.Context, shortCode string) (originalURL string, err error)
}

type Service struct {
	storage storage.Storage
}

func New(storage storage.Storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) ShortenURL(ctx context.Context, originalURL string) (shortCode string, isNew bool, err error) {
	shortCode = generateShortCode(originalURL)

	shortCode, isNew, err = s.storage.Save(ctx, shortCode, originalURL)
	if err != nil {
		return "", false, fmt.Errorf("failed to save URL mapping: %w", err)
	}
	return shortCode, isNew, nil
}

func (s *Service) Resolve(ctx context.Context, shortCode string) (originalURL string, err error) {
	originalURL, err = s.storage.Resolve(ctx, shortCode)
	if err != nil {
		return "", fmt.Errorf("failed to resolve short code: %w", err)
	}
	return originalURL, nil
}

func generateShortCode(originalURL string) string {
	hash := sha256.Sum256([]byte(originalURL))

	num := binary.BigEndian.Uint64(hash[:8])

	codeLength := shortCodeLength
	shortCode := make([]byte, codeLength)

	for i := 0; i < codeLength; i++ {
		remainder := num % 62
		shortCode[i] = base62Alphabet[remainder]

		num = num / 62
	}

	return string(shortCode)
}
