package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/tombell/memoir"
	"github.com/tombell/memoir/config"
	"github.com/tombell/memoir/datastore"
	"github.com/tombell/memoir/pkg/api"
	"github.com/tombell/memoir/services"
)

const helpText = `usage: memoir [<args>]

  --config   Path to .env.toml configuration file

Special options:
  --help     Show this message, then exit
  --version  Show the version number, then exit
`

var (
	cfgpath = flag.String("config", ".env.dev.toml", "")
	version = flag.Bool("version", false, "")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, helpText)
		os.Exit(2)
	}

	flag.Parse()

	if *version {
		fmt.Fprintf(os.Stdout, "memoir %s (%s)\n", memoir.Version, memoir.Commit)
		os.Exit(0)
	}

	logger := log.New(os.Stderr, "", log.LstdFlags)

	cfg, err := config.Load(*cfgpath)
	if err != nil {
		logger.Fatalf("error loading config: %v", err)
	}

	store, err := datastore.New(cfg.DB)
	if err != nil {
		logger.Fatalf("error connecting to database: %v", err)
	}
	defer store.Close()

	s := api.New(&services.Services{
		Logger:    logger,
		Config:    cfg,
		DataStore: store,
	})

	go func() {
		if err := s.Start(); err != nil {
			if err == http.ErrServerClosed {
				logger.Println("api server shutdown finished")
				return
			}

			logger.Fatalf("error starting api server: %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		logger.Fatalf("error shutting down api server: %v", err)
	}

}
