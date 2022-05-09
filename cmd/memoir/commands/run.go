package commands

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

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
	cmd := flag.NewFlagSet("migrate", flag.ExitOnError)
	cmd.Usage = usageText(runHelpText)

	cfgpath := cmd.String("config", ".env.dev.toml", "")

	cfg, err := config.Load(*cfgpath)
	if err != nil {
		logger.Fatalf("error: config load failed: %s", err)
	}

	ds, err := datastore.New(cfg.DB)
	if err != nil {
		logger.Fatalf("error: datastore initialise failed: %s", err)
	}
	defer ds.Close()

	fs := filestore.New(cfg.AWS.Key, cfg.AWS.Secret, cfg.AWS.Region)

	srv := api.New(&services.Services{
		Logger:    logger,
		Config:    cfg,
		DataStore: ds,
		FileStore: fs,
	})

	go func() {
		if err := srv.Start(); err != nil {
			if err == http.ErrServerClosed {
				logger.Println("api server shutdown finished")
				return
			}

			logger.Fatalf("error: starting api server failed: %s", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("error: server shutdown failed: %s", err)
	}
}
