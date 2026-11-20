package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/zeroqual/wallet-billing/internal/api"
	"github.com/zeroqual/wallet-billing/internal/config"
	"github.com/zeroqual/wallet-billing/internal/logger"
	"github.com/zeroqual/wallet-billing/internal/storage"
)

const (
	serviceName = "billing-service"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", serviceName, err)
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
	}

	cfg, err := config.Load[config.Billing]()
	if err != nil {
		return err
	}

	log := logger.New(serviceName, cfg.LogLevel)
	slog.SetDefault(log)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	dsn := cfg.Postgres.DSN()

	pool, err := storage.InitPostgres(ctx, dsn)
	if err != nil {
		return err
	}
	defer pool.Close()

	repo := api.NewPostgresRepository(pool)
	service := api.NewService(repo)

	handler := api.NewHandler(service)
	router := api.NewRouter(handler)

	if err := api.Serve(ctx, router, cfg.Addr); err != nil {
		return err
	}

	return nil
}
