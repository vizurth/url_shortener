package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vizurth/url_shortener/internal/storage"
	"github.com/vizurth/url_shortener/pkg/logger"
	psg "github.com/vizurth/url_shortener/pkg/postgres"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, cfg *psg.Config) (*PostgresStorage, error) {
	log := logger.From(ctx)

	if err := psg.Migrate(ctx, cfg); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	pool, err := psg.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not connect to postgres: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("could not ping postgres: %w", err)
	}

	log.Info(ctx, "successfully connected to postgres storage")
	return &PostgresStorage{pool: pool}, nil
}

func (s *PostgresStorage) Save(ctx context.Context, originalURL, shortCode string) (code string, isNew bool, err error) {
	query := `
		INSERT INTO urls (short_code, original_url)
		VALUES ($1, $2)
		ON CONFLICT (original_url)
		DO UPDATE SET original_url = EXCLUDED.original_url
		RETURNING short_code
	`

	var returnedCode string
	if err = s.pool.QueryRow(ctx, query, shortCode, originalURL).Scan(&returnedCode); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "urls_pkey" {
			return "", false, storage.ErrShortCodeConflict
		}
		return "", false, fmt.Errorf("could not execute query: %w", err)
	}

	return returnedCode, returnedCode == shortCode, nil
}

func (s *PostgresStorage) Resolve(ctx context.Context, shortCode string) (string, error) {
	query := `SELECT original_url FROM urls WHERE short_code = $1`
	var originalURL string
	err := s.pool.QueryRow(ctx, query, shortCode).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", storage.ErrNotFound
		}
		return "", fmt.Errorf("could not query postgres: %w", err)
	}
	return originalURL, nil
}

func (s *PostgresStorage) Close() error {
	s.pool.Close()
	return nil
}
