package commands

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/charmbracelet/log"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tombell/memoir/internal/api"
	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/datastore"
	"github.com/tombell/memoir/internal/filestore"
	"github.com/tombell/memoir/internal/services"
)

const runHelpText = `usage: memoir run [<args>]

  --config  Path to .env.toml configuration file

Special options:

  --help    Show this message, then exit
`

func RunCommand(logger *log.Logger) {
	cmd := flag.NewFlagSet("run", flag.ExitOnError)
	cmd.Usage = usageText(runHelpText)

	cfgpath := cmd.String("config", ".env.dev.toml", "")

	cfg, err := config.Load(*cfgpath)
	if err != nil {
		logger.Fatal("config load failed", "err", err)
	}

	dbpool, err := pgxpool.New(context.Background(), cfg.DB)
	if err != nil {
		logger.Fatal("datastore initialise failed", "err", err)
	}
	defer dbpool.Close()

	srv := api.New(&services.Services{
		Config:    cfg,
		DataStore: datastore.NewStore(dbpool),
		FileStore: filestore.New(cfg.AWS.Key, cfg.AWS.Secret, cfg.AWS.Region),
		Logger:    logger,
	})

	go func() {
		if err := srv.Start(); err != nil {
			if err == http.ErrServerClosed {
				logger.Info("api server shutdown finished")
				return
			}

			logger.Fatal("starting api server failed", "err", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server shutdown failed", "err", err)
	}
}
