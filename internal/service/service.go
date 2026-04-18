package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/vizurth/url_shortener/internal/storage"
)

var (
	ErrInvalidURL = errors.New("invalid URL")
	ErrNotFound   = errors.New("short code not found")
)

const (
	charset         = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	shortCodeLength = 10
)

type URLService interface {
	ShortenURL(ctx context.Context, originalURL string) (shortCode string, isNew bool, err error)
	Resolve(ctx context.Context, shortCode string) (originalURL string, err error)
}

type Service struct {
	storage storage.Storage
}

func New(store storage.Storage) *Service {
	return &Service{
		storage: store,
	}
}

func (s *Service) ShortenURL(ctx context.Context, originalURL string) (shortCode string, isNew bool, err error) {
	normalized, err := normalizeURL(originalURL)
	if err != nil {
		return "", false, err
	}

	for {
		code, err := generateShortCode()
		if err != nil {
			return "", false, fmt.Errorf("generate code: %w", err)
		}

		code, isNew, err = s.storage.Save(ctx, normalized, code)
		if err != nil {
			if errors.Is(err, storage.ErrShortCodeConflict) {
				continue
			}
			return "", false, fmt.Errorf("save url: %w", err)
		}
		return code, isNew, nil
	}
}

func normalizeURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidURL, err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("%w: %q is not a valid URL, example: https://example.com/path", ErrInvalidURL, raw)
	}

	if u.Host == "" {
		return "", fmt.Errorf("%w: %q is not a valid URL, example: https://example.com/path", ErrInvalidURL, raw)
	}

	// Hostname() обрезает дефолтный порт — https://example.com:443 → https://example.com
	u.Host = u.Hostname()
	u.Fragment = ""
	if u.Path == "/" {
		u.Path = ""
	}
	return u.String(), nil
}

func (s *Service) Resolve(ctx context.Context, shortCode string) (originalURL string, err error) {
	if len(shortCode) != shortCodeLength {
		return "", fmt.Errorf("%w: short code must be %d characters long", ErrNotFound, shortCodeLength)
	}
	originalURL, err = s.storage.Resolve(ctx, shortCode)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("resolve: %w", err)
	}
	return originalURL, nil
}

func generateShortCode() (string, error) {
	b := make([]byte, shortCodeLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}
