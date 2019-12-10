package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/tombell/memoir/pkg/config"
	"github.com/tombell/memoir/pkg/datastore"
	"github.com/tombell/memoir/pkg/services"
)

const importHelpText = `usage: memoir-tracklists import [<args>]

  --config     Path to .env.toml configuration file
  --tracklist  Name of the tracklist to import
  --date       Date for the tracklist being imported
  --serato     Tracklist is an exported file from Serato
  --rekordbox  Tracklist is en exported file from Rekordbox DJ

Special options:

  --help       Show this message, then exit
`

func importTracklist() error {
	cmd := flag.NewFlagSet("import", flag.ExitOnError)
	cmd.Usage = usage(importHelpText)

	cfgpath := cmd.String("config", ".env.dev.toml", "")
	tracklist := cmd.String("tracklist", "", "")
	date := cmd.String("date", "", "")
	isSerato := cmd.Bool("serato", false, "")
	isRekordbox := cmd.Bool("rekordbox", false, "")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		return err
	}

	if *tracklist == "" || *date == "" {
		cmd.Usage()
	}

	if (!*isSerato && !*isRekordbox) || (*isSerato && *isRekordbox) {
		cmd.Usage()
	}

	if !cmd.Parsed() {
		return nil
	}

	args := cmd.Args()
	if len(args) != 1 {
		cmd.Usage()
	}

	parsedDate, err := time.Parse(dateTimeFormat, *date)
	if err != nil {
		return err
	}

	var records [][]string

	if *isSerato {
		if records, err = parseSeratoExport(args[0]); err != nil {
			return err
		}
	}

	if *isRekordbox {
		if records, err = parseRekordboxExport(args[0]); err != nil {
			return err
		}
	}

	cfg, err := config.Load(*cfgpath)
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	store, err := datastore.New(cfg.DB)
	defer store.Close()
	if err != nil {
		return err
	}

	svc := services.Services{
		DataStore: store,
	}

	if _, err := svc.ImportTracklist(*tracklist, parsedDate, records); err != nil {
		return err
	}

	return nil

}
