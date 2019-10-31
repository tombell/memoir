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
	"github.com/tombell/memoir/api"
	"github.com/tombell/memoir/datastore"
	"github.com/tombell/memoir/services"
)

const helpText = `usage: memoir [<args>]

  --db  connection string for connecting to the database

Special options:
  --help      show this message, then exit
  --version   show the version number, then exit
`

var (
	dsn     = flag.String("db", "", "")
	version = flag.Bool("version", false, "")
)

func usage() {
	fmt.Fprintf(os.Stderr, helpText)
	os.Exit(2)
}

func validateFlags() {
	if *dsn == "" {
		flag.Usage()
	}
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if *version {
		fmt.Fprintf(os.Stdout, "memoir %s (%s)\n", memoir.Version, memoir.Commit)
		os.Exit(0)
	}

	logger := log.New(os.Stderr, "[memoir] ", log.LstdFlags)

	validateFlags()

	ds, err := datastore.New(*dsn)
	if err != nil {
		logger.Fatalf("error: %v\n", err)
	}
	defer ds.Close()

	cfg := &api.Config{
		Logger: logger,
		Services: &services.Services{
			Logger:    logger,
			DataStore: ds,
		},
	}

	s := api.NewServer(cfg)

	logger.Println("starting memoir api server...")

	go func() {
		if err := s.Start(":8080"); err != nil {
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
