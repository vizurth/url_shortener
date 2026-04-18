package main

import (
	"context"
	"fmt"
	"os"

	"github.com/vizurth/url_shortener/internal/app"
	"github.com/vizurth/url_shortener/internal/config"
	"github.com/vizurth/url_shortener/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	log, err := logger.New()
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer func() { _ = log.Sync() }()

	ctx := logger.With(context.Background(), log)

	cfg, err := config.Load()
	if err != nil {
		log.Error(ctx, "failed to load config", zap.Error(err))
		return fmt.Errorf("load config: %w", err)
	}

	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Error(ctx, "failed to create application", zap.Error(err))
		return fmt.Errorf("init app: %w", err)
	}

	return application.Run(ctx)
}
