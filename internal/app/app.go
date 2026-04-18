package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/vizurth/url_shortener/internal/config"
	"github.com/vizurth/url_shortener/pkg/logger"
	"go.uber.org/zap"
)

type App struct {
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	return &App{}, nil
}

func (a *App) Run(ctx context.Context) error {
	log := logger.From(ctx)
	log.Info(ctx, "starting application")
	defer func() {
		if p := recover(); p != nil {
			log.Error(ctx, "application panicked", zap.Any("panic", p))
			if err := a.Shutdown(ctx); err != nil {
				log.Error(ctx, "error during shutdown after panic", zap.Error(err))
			}
			os.Exit(1)
		}
	}()

	// запуск HTTP сервера и других необходимых компонентов приложения
	errCh := make(chan error)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case <-quit:
		log.Info(ctx, "shutdown signal received")
	case err := <-errCh:
		if err != nil {
			log.Error(ctx, "server stopped with error", zap.Error(err))
			if shutdownErr := a.Shutdown(ctx); shutdownErr != nil {
				log.Error(ctx, "error during shutdown after server error", zap.Error(shutdownErr))
			}

			return err
		}
		if shutdownErr := a.Shutdown(ctx); shutdownErr != nil {
			log.Error(ctx, "error during shutdown after server stopped", zap.Error(shutdownErr))
		}

		return nil
	}

	return a.Shutdown(ctx)
}

func (a *App) Shutdown(ctx context.Context) error {
	log := logger.From(ctx)
	log.Info(ctx, "starting graceful shutdown...")

	log.Info(ctx, "closing HTTP server")
	// TODO: закрыть HTTP сервер, например, вызвав a.api.Shutdown(ctx) или аналогичный метод
	log.Info(ctx, "HTTP server shutdown successfully")

	log.Info(ctx, "closing database connection pool")
	// TODO: закрыть пул соединений с базой данных, например, вызвав a.db.Close() или аналогичный метод
	log.Info(ctx, "database pool closed successfully")

	log.Info(ctx, "graceful shutdown completed")

	return nil
}
