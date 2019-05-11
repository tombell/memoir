package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/cheynewallace/tabby"

	"github.com/tombell/memoir/datastore"
	"github.com/tombell/memoir/services"
)

const (
	listHelpText = `usage: memoir-tracklists list [<args>]

  --db    connection string for connecting to the database

Special options:
  --help  show this message, then exit
`
	showHelpText = `usage: memoir-tracklists show [<args>]

  --db         connection string for connecting to the database
  --tracklist  name for the tracklist to show

Special options:
  --help       show this message, then exit
`

	importHelpText = `usage: memoir-tracklists import [<args>] <path to file>

  --db         connection string for connecting to the database
  --tracklist  name for the tracklist being imported

Special options:
  --help  show this message, then exit
`

	exportHelpText = `usage: memoir-tracklists export [<args>]

  --db         connection string for connecting to the database
  --tracklist  name of the tracklist to export

Special options:
  --help  show this message, then exit
`

	deleteHelpText = `usage: memoir-tracklists delete [<args>]

  --db         connection string for connecting to the database
  --tracklist  name of the tracklist to delete

Special options:
  --help  show this message, then exit
`
)

func listTracklists(logger *log.Logger) error {
	cmd := flag.NewFlagSet("list", flag.ExitOnError)
	cmd.Usage = usage(listHelpText)
	dsn := cmd.String("db", "", "")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		return err
	}

	if *dsn == "" {
		cmd.Usage()
	}

	if !cmd.Parsed() {
		return nil
	}

	store, err := datastore.New(*dsn)
	if err != nil {
		return err
	}
	defer store.Close()

	svc := services.Services{
		Logger:    logger,
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

func showTracklist(logger *log.Logger) error {
	cmd := flag.NewFlagSet("show", flag.ExitOnError)
	cmd.Usage = usage(showHelpText)
	dsn := cmd.String("db", "", "")
	tracklist := cmd.String("tracklist", "", "")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		return err
	}

	if *dsn == "" || *tracklist == "" {
		cmd.Usage()
	}

	if !cmd.Parsed() {
		return nil
	}

	store, err := datastore.New(*dsn)
	if err != nil {
		return err
	}
	defer store.Close()

	svc := services.Services{
		Logger:    logger,
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

func importTracklist(logger *log.Logger) error {
	cmd := flag.NewFlagSet("import", flag.ExitOnError)
	cmd.Usage = usage(importHelpText)
	dsn := cmd.String("db", "", "")
	tracklist := cmd.String("tracklist", "", "")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		return err
	}

	if *dsn == "" || *tracklist == "" {
		cmd.Usage()
	}

	if !cmd.Parsed() {
		return nil
	}

	args := cmd.Args()
	if len(args) != 1 {
		cmd.Usage()
	}

	records, err := parseSeratoExport(args[0])
	if err != nil {
		return err
	}

	date, err := time.Parse(dateTimeFormat, records[0][0])
	if err != nil {
		return err
	}

	store, err := datastore.New(*dsn)
	if err != nil {
		return err
	}
	defer store.Close()

	svc := services.Services{
		Logger:    logger,
		DataStore: store,
	}

	if _, err := svc.ImportTracklist(*tracklist, date, records[1:]); err != nil {
		return err
	}

	return nil
}

func exportTracklist(logger *log.Logger) error {
	cmd := flag.NewFlagSet("export", flag.ExitOnError)
	cmd.Usage = usage(exportHelpText)
	dsn := cmd.String("db", "", "")
	tracklist := cmd.String("tracklist", "", "")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		return err
	}

	if *dsn == "" || *tracklist == "" {
		cmd.Usage()
	}

	if !cmd.Parsed() {
		return nil
	}

	store, err := datastore.New(*dsn)
	if err != nil {
		return err
	}
	defer store.Close()

	svc := services.Services{
		Logger:    logger,
		DataStore: store,
	}

	return svc.ExportTracklist(*tracklist, os.Stdout)
}

func deleteTracklist(logger *log.Logger) error {
	cmd := flag.NewFlagSet("delete", flag.ExitOnError)
	cmd.Usage = usage(deleteHelpText)
	dsn := cmd.String("db", "", "")
	tracklist := cmd.String("tracklist", "", "")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		return err
	}

	if *dsn == "" || *tracklist == "" {
		cmd.Usage()
	}

	if !cmd.Parsed() {
		return nil
	}

	store, err := datastore.New(*dsn)
	if err != nil {
		return err
	}
	defer store.Close()

	svc := services.Services{
		Logger:    logger,
		DataStore: store,
	}

	return svc.RemoveTracklist(*tracklist)
}
