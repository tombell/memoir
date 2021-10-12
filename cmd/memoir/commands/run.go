package commands

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/tombell/memoir/pkg/api"
	"github.com/tombell/memoir/pkg/config"
	"github.com/tombell/memoir/pkg/datastore"
	"github.com/tombell/memoir/pkg/filestore"
	"github.com/tombell/memoir/pkg/services"
)

const runHelpText = `usage: memoir run [<args>]

  --config  Path to .env.toml configuration file

Special options:

  --help    Show this message, then exit
`

// RunCommand ...
func RunCommand(logger *log.Logger) {
	cmd := flag.NewFlagSet("migrate", flag.ExitOnError)
	cmd.Usage = usageText(runHelpText)

	cfgpath := cmd.String("config", ".env.dev.toml", "")

	cfg, err := config.Load(*cfgpath)
	if err != nil {
		logger.Fatalf("error: config load failed: %s", err)
	}

	ds, err := datastore.New(cfg.DB)
	defer ds.Close()
	if err != nil {
		logger.Fatalf("error: datastore initialise failed: %s", err)
	}

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
