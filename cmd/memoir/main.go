package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tombell/memoir/internal/api"
	"github.com/tombell/memoir/internal/auth"
	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/mail"
	"github.com/tombell/memoir/internal/stores/artworkstore"
	"github.com/tombell/memoir/internal/stores/datastore"
	"github.com/tombell/memoir/internal/stores/filestore"
	"github.com/tombell/memoir/internal/stores/trackliststore"
	"github.com/tombell/memoir/internal/stores/trackstore"
	"github.com/tombell/memoir/internal/telemetry"
)

func main() {
	os.Exit(run())
}

func run() int {
	logger := slog.New(log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
		TimeFunction:    log.NowUTC,
		TimeFormat:      time.RFC3339,
		ReportCaller:    true,
	}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed loading config", "err", err)
		return 1
	}

	consoleLogger := logger
	logger, shutdownLogs, err := telemetry.NewLogger(context.Background(), cfg, logger)
	if err != nil {
		consoleLogger.Error("failed configuring log export", "err", err)
		return 1
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := shutdownLogs(ctx); err != nil {
			consoleLogger.Error("failed shutting down log export", "err", err)
		}
	}()

	dbpool, err := pgxpool.New(context.Background(), cfg.DB)
	if err != nil {
		logger.Error("failed creating database connection pool", "err", err)
		return 1
	}
	defer dbpool.Close()

	dataStore := datastore.New(dbpool)
	fileStore := filestore.New(cfg)
	accounts, err := auth.New(dataStore, cfg.Auth, mail.New(cfg.SMTP), logger)
	if err != nil {
		logger.Error("failed configuring accounts", "err", err)
		return 1
	}
	cleanupCtx, stopCleanup := context.WithCancel(context.Background())
	defer stopCleanup()
	go accounts.RunCleanup(cleanupCtx)
	go accounts.RunEmail(cleanupCtx)

	server := api.New(
		logger,
		cfg,
		trackliststore.New(dataStore),
		trackstore.New(dataStore),
		artworkstore.New(fileStore),
		accounts,
	)

	idleConnsClosed := make(chan struct{})
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(done)

	go func() {
		<-done

		logger.Info("shutting down api server")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("failed to shutdown the api server", "err", err)
		}

		close(idleConnsClosed)
	}()

	logger.Info("starting api server", "address", fmt.Sprintf("http://%s", cfg.Address))

	if err := server.Run(); err != nil {
		logger.Error("failed to start the api server", "err", err)
		return 1
	}

	<-idleConnsClosed
	return 0
}
