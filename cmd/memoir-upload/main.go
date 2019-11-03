package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tombell/memoir"
	"github.com/tombell/memoir/pkg/config"
	"github.com/tombell/memoir/pkg/datastore"
	"github.com/tombell/memoir/pkg/filestore/s3"
	"github.com/tombell/memoir/pkg/services"
)

const helpText = `usage: memoir-upload [<args>] <path to mix mp3>

  --config     Path to .env.toml configuration file
  --tracklist  Name of the tracklist to associate the uploaded mix with

Special options:
  --help       Show this message, then exit
  --version    Show the version number, then exit
`

var (
	cfgpath   = flag.String("config", ".env.dev.toml", "")
	tracklist = flag.String("tracklist", "", "")
	version   = flag.Bool("version", false, "")
)

func usage() {
	fmt.Fprintf(os.Stderr, helpText)
	os.Exit(2)
}

func validateFlags() {
	if *tracklist == "" {
		flag.Usage()
	}
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if *version {
		fmt.Fprintf(os.Stdout, "memoir-upload %s (%s)\n", memoir.Version, memoir.Commit)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
	}

	validateFlags()

	logger := log.New(os.Stderr, "", 0)

	cfg, err := config.Load(*cfgpath)
	if err != nil {
		logger.Fatalf("error loading config: %v", err)
	}

	ds, err := datastore.New(cfg.DB)
	if err != nil {
		logger.Fatalf("error connecting to database: %v\n", err)
	}
	defer ds.Close()

	fs := s3.New(cfg.AWS.Key, cfg.AWS.Secret)

	svc := &services.Services{
		Logger:    logger,
		DataStore: ds,
		FileStore: fs,
	}

	logger.Printf("uploading mix %s...\n", *tracklist)

	key, err := svc.UploadMix(args[0], *tracklist)
	if err != nil {
		logger.Fatalf("error uploading mix: %v\n", err)
	}

	logger.Printf("uploaded mix: %s\n", key)
}
