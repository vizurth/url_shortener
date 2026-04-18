package storage

import (
	"context"
	"errors"
)

var (
	ErrNotFound      = errors.New("short URL not found")
	ErrAlreadyExists = errors.New("short URL already exists")
)

type Storage interface {
	Save(ctx context.Context, originalURL, shortCode string) (string, bool, error)
	Resolve(ctx context.Context, shortCode string) (string, error)
	Close() error
}
