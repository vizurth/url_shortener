package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vizurth/url_shortener/pkg/logger"
)

type Config struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	MaxConns int32  `yaml:"max_conns" env:"MAX_CONNS" env-default:"10"`
	MinConns int32  `yaml:"min_conns" env:"MIN_CONNS" env-default:"5"`
}

func New(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.GetConnString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	return pool, nil
}

func Migrate(ctx context.Context, cfg Config) error {
	connString := cfg.GetConnString()
	log := logger.From(ctx)

	migrationPaths := []string{
		"migrations",
		"./migrations",
		"../migrations",
		"../../migrations",
	}

	var migrationPath string
	for _, path := range migrationPaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			migrationPath, _ = filepath.Abs(path)
			break
		}
	}

	if migrationPath == "" {
		return fmt.Errorf("migrations directory not found")
	}

	m, err := migrate.New("file://"+migrationPath, connString)

	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info(ctx, "migrated successfully")
	return nil
}

func (c *Config) GetConnString() string {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
	)
	return connString
}
