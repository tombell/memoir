package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tombell/memoir"
)

const (
	helpText = `usage: memoir-tracklists <command> [<args>]

Commands:
  list       List all tracklists, most recent to oldest
  export     Export a tracklist to a simple "artist - title" list
  delete     Delete a tracklist

Special options:
  --help     Show this message, then exit
  --version  Show the version number, then exit
`
)

var (
	version = flag.Bool("version", false, "")
)

func usage(text string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, text)
		os.Exit(2)
	}
}

func main() {
	flag.Usage = usage(helpText)
	flag.Parse()

	if *version {
		fmt.Fprintf(os.Stdout, "memoir-tracklists %s (%s)\n", memoir.Version, memoir.Commit)
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		flag.Usage()
	}

	logger := log.New(os.Stderr, "", 0)

	var err error

	switch os.Args[1] {
	case "list", "l":
		err = listTracklists()
	case "export", "e":
		err = exportTracklist()
	case "delete", "d":
		err = deleteTracklist()
	default:
		err = fmt.Errorf("%q is not a valid command", os.Args[1])
	}

	if err != nil {
		logger.Fatalf("error: %s\n", err)
	}
}
