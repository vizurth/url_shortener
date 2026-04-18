package storage

import (
	"context"
	"errors"
)

var (
	ErrNotFound          = errors.New("short URL not found")
	ErrShortCodeConflict = errors.New("short code already taken")
)

type Storage interface {
	Save(ctx context.Context, originalURL, shortCode string) (string, bool, error)
	Resolve(ctx context.Context, shortCode string) (string, error)
	Close() error
}
