package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	slugify "github.com/metal3d/go-slugify"

	"github.com/tombell/memoir/database"
	"github.com/tombell/memoir/services"
)

const helpText = `usage: memoir-upload [args] <path to mix mp3>

  --db         connection string for connecting to the database
  --tracklist  name of the tracklist to associate the uploaded mix with
  --aws-key
  --aws-secret

Special options:
  --help      show this message, then exit
  --version   show the version number, then exit
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
		fmt.Fprintf(os.Stdout, "memoir-upload %s (%s)\n", Version, Commit)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
	}

	validateFlags()

	logger := log.New(os.Stderr, "", 0)

	db, err := database.Open(*dsn)
	if err != nil {
		logger.Fatalf("err: %v\n", err)
	}

	f, err := os.Open(args[0])
	if err != nil {
		logger.Fatalf("err: %v\n", err)
	}
	defer f.Close()

	var buf [512]byte

	if _, err := f.Read(buf[:]); err != nil {
		logger.Fatalf("err: %v\n", err)
	}

	contentType := http.DetectContentType(buf)

	cfg := &services.Config{
		AWS: &services.AWSConfig{
			Key:    *awsKey,
			Secret: *awsSecret,
		},
	}

	svc := &services.Services{
		Config: cfg,
		Logger: logger,
		DB:     db,
	}

	filename := filepath.Base(args[0])
	ext := filepath.Ext(filename)
	slug := slugify.Marshal(filename[:len(filename)-len(ext)])
	key := strings.ToLower(slug + ext)

	logger.Printf("uploading mix file %q as %q...", filename, key)

	location, err := svc.Upload(f, key, contentType)
	if err != nil {
		logger.Fatalf("err: %v\n", err)
	}

	if err := svc.AssociateUpload(filename, location, *tracklist); err != nil {
		logger.Fatalf("err: %v\n", err)
	}

	logger.Printf("uploaded mix file to %s\n", location)
}
