package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vizurth/url_shortener/internal/config"
	"github.com/vizurth/url_shortener/internal/service"
	"github.com/vizurth/url_shortener/internal/storage"
	"github.com/vizurth/url_shortener/internal/storage/memory"
	pgstore "github.com/vizurth/url_shortener/internal/storage/postgres"
	"github.com/vizurth/url_shortener/internal/transport"
	"github.com/vizurth/url_shortener/pkg/logger"
	psg "github.com/vizurth/url_shortener/pkg/postgres"
	"go.uber.org/zap"
)

type App struct {
	cfg     config.Config
	log     *logger.Logger
	storage storage.Storage
	server  *http.Server
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	log := logger.From(ctx)

	var (
		store storage.Storage
		err   error
	)

	switch cfg.Storage.Type {
	case "postgres":
		store, err = pgstore.NewPostgresStorage(ctx, psg.Config{
			Host:     cfg.Postgres.Host,
			Port:     cfg.Postgres.Port,
			Username: cfg.Postgres.Username,
			Password: cfg.Postgres.Password,
			Database: cfg.Postgres.Database,
			MaxConns: cfg.Postgres.MaxConns,
			MinConns: cfg.Postgres.MinConns,
		})
		if err != nil {
			return nil, fmt.Errorf("init postgres storage: %w", err)
		}
	case "memory":
		store = memory.NewMemoryStorage()
	default:
		return nil, fmt.Errorf("unknown storage type: %q", cfg.Storage.Type)
	}

	svc := service.New(store)
	handler := transport.NewHandler(svc, cfg.Query.ShortURLBase)
	router := transport.NewRouter(handler)

	srv := &http.Server{
		Addr:         cfg.Query.HTTPAddr,
		Handler:      router,
		ReadTimeout:  cfg.Query.ReadTimeout.Duration,
		WriteTimeout: cfg.Query.WriteTimeout.Duration,
		IdleTimeout:  cfg.Query.IdleTimeout.Duration,
	}

	log.Info(ctx, "app initialized",
		zap.String("storage", cfg.Storage.Type),
		zap.String("addr", cfg.Query.HTTPAddr),
	)

	return &App{
		cfg:     cfg,
		log:     log,
		storage: store,
		server:  srv,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		a.log.Info(ctx, "server starting", zap.String("addr", a.cfg.Query.HTTPAddr))
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.log.Fatal(ctx, "server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	a.log.Info(ctx, "shutdown signal received")
	return a.Shutdown()
}

func (a *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.log.Error(ctx, "server shutdown error", zap.Error(err))
	}

	if err := a.storage.Close(); err != nil {
		a.log.Error(ctx, "storage close error", zap.Error(err))
	}

	a.log.Info(ctx, "shutdown complete")
	_ = a.log.Sync()
	return nil
}
