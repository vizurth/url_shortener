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
	log, err := logger.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	defer func() {
		if err := log.Sync(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = logger.With(ctx, log)

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(ctx, "failed to load config", zap.Error(err))
	}
	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatal(ctx, "failed to create application", zap.Error(err))
		return
	}

	if err := application.Run(ctx); err != nil {
		log.Fatal(ctx, "application error", zap.Error(err))
		return
	}
}
