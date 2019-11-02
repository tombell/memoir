package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cheynewallace/tabby"

	"github.com/tombell/memoir/pkg/config"
	"github.com/tombell/memoir/pkg/datastore"
	"github.com/tombell/memoir/services"
)

const showHelpText = `usage: memoir-trackshows show [<args>]

  --config     Path to .env.toml configuration file
  --tracklist  Name of the tracklist to export

Special options:

  --help       Show this message, then exit
`

func show() error {
	cmd := flag.NewFlagSet("show", flag.ExitOnError)
	cmd.Usage = usage(showHelpText)

	cfgpath := cmd.String("config", ".env.dev.toml", "")
	tracklist := cmd.String("tracklist", "", "")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("cmd parse failed: %w", err)
	}

	if *tracklist == "" {
		cmd.Usage()
	}

	if !cmd.Parsed() {
		return nil
	}

	cfg, err := config.Load(*cfgpath)
	if err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	store, err := datastore.New(cfg.DB)
	if err != nil {
		return err
	}
	defer store.Close()

	svc := services.Services{
		DataStore: store,
	}

	tl, err := svc.GetTracklistByName(*tracklist)
	if err != nil {
		return err
	}

	t := tabby.New()
	t.AddHeader("ID", "Name", "Artist", "BPM", "Genre", "Key")

	for _, track := range tl.Tracks {
		t.AddLine(track.ID, track.Name, track.Artist, track.BPM, track.Genre, track.Key)
	}

	t.Print()

	return nil
}
