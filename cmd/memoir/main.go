package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tombell/memoir/internal/api"
	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/stores/artworkstore"
	"github.com/tombell/memoir/internal/stores/datastore"
	"github.com/tombell/memoir/internal/stores/filestore"
	"github.com/tombell/memoir/internal/stores/trackliststore"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

func main() {
	logger := slog.New(log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
		TimeFunction:    log.NowUTC,
		TimeFormat:      time.RFC3339,
		ReportCaller:    true,
	}))

	cfgpath := flag.String("config", "config.dev.json", "")
	flag.Parse()

	cfg, err := config.Load(*cfgpath)
	if err != nil {
		logger.Error("failed loading config file", "err", err)
		os.Exit(1)
	}

	dbpool, err := pgxpool.New(context.Background(), cfg.DB)
	if err != nil {
		logger.Error("failed creating database connection pool", "err", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	dataStore := datastore.New(dbpool)
	fileStore := filestore.New(cfg)

	server := api.New(
		logger,
		cfg,
		trackliststore.New(dataStore),
		trackstore.New(dataStore),
		artworkstore.New(fileStore),
	)

	idleConnsClosed := make(chan struct{})

	go func() {
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
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
		os.Exit(1)
	}

	<-idleConnsClosed
}
