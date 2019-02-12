package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tombell/memoir"
	"github.com/tombell/memoir/datastore"
	"github.com/tombell/memoir/filestore/s3"
	"github.com/tombell/memoir/services"
)

const helpText = `usage: memoir-upload [args] <path to mix mp3>

  --db         connection string for connecting to the database
  --tracklist  name of the tracklist to associate the uploaded mix with
  --aws-key    AWS access key ID
  --aws-secret AWS secret access key

Special options:
  --help     show this message, then exit
  --version  show the version number, then exit
`

var (
	dsn       = flag.String("db", "", "")
	tracklist = flag.String("tracklist", "", "")
	awsKey    = flag.String("aws-key", "", "")
	awsSecret = flag.String("aws-secret", "", "")
	version   = flag.Bool("version", false, "")
)

func usage() {
	fmt.Fprintf(os.Stderr, helpText)
	os.Exit(2)
}

func validateFlags() {
	if *dsn == "" {
		flag.Usage()
	}

	if *tracklist == "" {
		flag.Usage()
	}

	if *awsKey == "" || *awsSecret == "" {
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

	logger.Printf("uploading mix %s...\n", *tracklist)

	ds, err := datastore.New(*dsn)
	if err != nil {
		logger.Fatalf("error: %v\n", err)
	}
	defer ds.Close()

	fs := s3.New("memoir-uploads", *awsKey, *awsSecret)

	svc := &services.Services{
		Logger:    logger,
		DataStore: ds,
		FileStore: fs,
	}

	key, err := svc.UploadMix(args[0], *tracklist)
	if err != nil {
		logger.Fatalf("error uploading mix: %v\n", err)
	}

	logger.Printf("uploaded mix: %s\n", key)
}
