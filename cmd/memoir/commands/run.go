package commands

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
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

const runHelpText = `usage: memoir run [<args>]

  --config  Path to .env.toml configuration file

Special options:

  --help    Show this message, then exit
`

func RunCommand(logger *slog.Logger) {
	cmd := flag.NewFlagSet("run", flag.ExitOnError)
	cmd.Usage = usageText(runHelpText)

	cfgpath := cmd.String("config", "config.dev.json", "")

	cfg, err := config.Load(*cfgpath)
	if err != nil {
		logger.Error("config load failed", "err", err)
	}

	dbpool, err := pgxpool.New(context.Background(), cfg.DB)
	if err != nil {
		logger.Error("datastore initialise failed", "err", err)
	}
	defer dbpool.Close()

	dataStore := datastore.NewStore(dbpool)
	fileStore := filestore.New(cfg)

	srv := api.New(
		logger,
		cfg,
		trackliststore.New(dataStore),
		trackstore.New(dataStore),
		artworkstore.New(fileStore),
	)

	go func() {
		if err := srv.Start(logger); err != nil {
			if err == http.ErrServerClosed {
				logger.Info("api server shutdown finished")
				return
			}

			logger.Error("starting api server failed", "err", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx, logger); err != nil {
		logger.Error("server shutdown failed", "err", err)
	}
}
