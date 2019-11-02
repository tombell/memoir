package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cheynewallace/tabby"

	"github.com/tombell/memoir/pkg/config"
	"github.com/tombell/memoir/pkg/datastore"
	"github.com/tombell/memoir/pkg/services"
)

const listHelpText = `usage: memoir-tracklists list [<args>]

  --config  Path to .env.toml configuration file

Special options:

  --help    Show this message, then exit
`

func list() error {
	cmd := flag.NewFlagSet("list", flag.ExitOnError)
	cmd.Usage = usage(listHelpText)

	cfgpath := cmd.String("config", ".env.dev.toml", "")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("cmd parse failed: %w", err)
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

	tracklists, err := svc.GetTracklists()
	if err != nil {
		return err
	}

	t := tabby.New()
	t.AddHeader("ID", "Name", "Date")

	for _, tracklist := range tracklists {
		t.AddLine(tracklist.ID, tracklist.Name, tracklist.Date.Format("2006-01-02"))
	}

	t.Print()

	return nil
}
