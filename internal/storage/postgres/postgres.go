package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vizurth/url_shortener/pkg/logger"
	psg "github.com/vizurth/url_shortener/pkg/postgres"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, cfg psg.Config) (*PostgresStorage, error) {
	log := logger.From(ctx)
	pool, err := psg.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not connect to postgres: %w", err)
	}
	err = psg.Migrate(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not migrate postgres: %w", err)
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not ping postgres: %w", err)
	}
	log.Info(ctx, "successfully connected to postgres storage")
	return &PostgresStorage{pool: pool}, nil
}

func (s *PostgresStorage) Save(ctx context.Context, originalURL, shortCode string) (string, bool, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", false, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO urls (short_code, original_url)
		VALUES ($1, $2)
		ON CONFLICT (original_url)
		DO UPDATE SET original_url = EXCLUDED.original_url
		RETURNING short_code
	`

	var returnedCode string
	err = tx.QueryRow(ctx, query, shortCode, originalURL).Scan(&returnedCode)
	if err != nil {
		return "", false, fmt.Errorf("could not execute query: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return "", false, fmt.Errorf("could not commit transaction: %w", err)
	}

	isNew := returnedCode == shortCode

	return returnedCode, isNew, nil
}

func (s *PostgresStorage) Resolve(ctx context.Context, shortCode string) (string, error) {
	query := `SELECT original_url FROM urls WHERE short_code = $1`
	var originalURL string
	err := s.pool.QueryRow(ctx, query, shortCode).Scan(&originalURL)
	if err != nil {
		return "", fmt.Errorf("could not query postgres: %w", err)
	}
	return originalURL, nil
}

func (s *PostgresStorage) Close() error {
	s.pool.Close()
	return nil
}
