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

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tombell/memoir/internal/api"
	"github.com/tombell/memoir/internal/artworkstore"
	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/datastore"
	"github.com/tombell/memoir/internal/filestore"
	"github.com/tombell/memoir/internal/trackliststore"
	"github.com/tombell/memoir/internal/trackstore"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	cfgpath := flag.String("config", "config.dev.json", "")
	flag.Parse()

	cfg, err := config.Load(*cfgpath)
	if err != nil {
		logger.Error("config load failed", "err", err)
		os.Exit(1)
	}

	dbpool, err := pgxpool.New(context.Background(), cfg.DB)
	if err != nil {
		logger.Error("datastore initialise failed", "err", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	dataStore := datastore.NewStore(dbpool)
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
			logger.Error("could not shutdown api server", "err", err)
		}

		close(idleConnsClosed)
	}()

	logger.Info("starting api server", "address", fmt.Sprintf("http://%s", cfg.Address))

	if err := server.Run(); err != nil {
		logger.Error("could not start api server", "err", err)
		os.Exit(1)
	}

	<-idleConnsClosed
}
