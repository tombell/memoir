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
  list       list all imported tracklist, most recent to oldest
  import     import a CSV tracklist from Serato, or txt tracklist from Rekordbox
  export     export a tracklist to a simple "artist - title" list
  delete     delete a tracklist

Special options:
  --help     show this message, then exit
  --version  show the version number, then exit
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

	switch os.Args[1] {
	case "list", "l":
		if err := listTracklists(logger); err != nil {
			logger.Fatalf("error listing tracklists: %q\n", err)
		}
	case "show", "s":
		if err := showTracklist(logger); err != nil {
			logger.Fatalf("error showing tracklist: %q\n", err)
		}
	case "import", "i":
		if err := importTracklist(logger); err != nil {
			logger.Fatalf("error importing tracklist: %q\n", err)
		}
	case "export", "e":
		if err := exportTracklist(logger); err != nil {
			logger.Fatalf("error exporting tracklist: %q\n", err)
		}
	case "delete", "d":
		if err := deleteTracklist(logger); err != nil {
			logger.Fatalf("error deleting tracklist: %q\n", err)
		}
	default:
		logger.Fatalf("error: %q is not a valid command\n", os.Args[1])
	}
}
