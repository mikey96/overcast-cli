package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var Config = ConfigT{
	Url: "https://api.search.overcast-security.app",
	// this is overwritten by -key flag
	Key: os.Getenv("OVERCAST_API_KEY"),
}

func app() {
	var initKey = func(key string) {
		if key == "" && Config.Key == "" {
			log.Fatalf("No API key provided\nYou can provide one by using the --key flag (priority)\nor OVERCAST_API_KEY environment variable\nGet or update your key here: https://search.overcast-security.app/profile")
		}
		if key != "" {
			Config.Key = key
		}
	}

	var keyFlag = func(mut *flag.FlagSet) *string {
		return mut.String("key", "", "Use API key (uses env if not supplied)")
	}

	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
	searchAll := searchCmd.Bool("all", false, "get all pages")
	searchPage := searchCmd.Int("page", 1, "get page")
	searchKey := keyFlag(searchCmd)

	metadataCmd := flag.NewFlagSet("metadata", flag.ExitOnError)
	metadataKey := keyFlag(metadataCmd)

	overviewCmd := flag.NewFlagSet("overview", flag.ExitOnError)
	overviewKey := keyFlag(overviewCmd)

	subsCmd := flag.NewFlagSet("subs", flag.ExitOnError)
	subsKey := keyFlag(subsCmd)

	if len(os.Args) <= 1 {
		fmt.Print(help())
		return
	}

	switch os.Args[1] {
	case "help":
		fmt.Print(help())
	case "subs":
		subsCmd.Parse(os.Args[2:])
		initKey(*subsKey)
		Subdomains(QueryStringArgs(subsCmd.Args()))
	case "search":
		searchCmd.Parse(os.Args[2:])
		initKey(*searchKey)
		if *searchAll {
			*searchPage = -1
		}
		Search(*searchPage, QueryStringArgs(searchCmd.Args()))
	case "metadata":
		metadataCmd.Parse(os.Args[2:])
		initKey(*metadataKey)
		Metadata(QueryStringArgs(metadataCmd.Args()))
	case "overview":
		overviewCmd.Parse(os.Args[2:])
		initKey(*overviewKey)
		Overview(overviewCmd.Args()[0])
	default:
		fmt.Print(help())
	}
}

func help() string {
	return fmt.Sprintln(
`Subcommands:
  - search
    - Options: -key -all (default: false) -page (default: 1)
    - Arguments: query string
    Search assets

  - subs
    - Options: -key
    - Arguments: root domain or query string
    Get subdomains of a root domain

  - metadata
    - Options: -key
    - Arguments: query string
    Get metadata for a search

  - overview
    - Options: -key
    - Arguments: root domain
    Get data about the assets matching a search query

  - help
    View this page

Links:
  - Get api key  https://search.overcast-security.app/profile
  - Github repo  https://github.com/mikey96/overcast-cli/`)
}

func main() {
	go func() {
		defer CloseWriter()
		app()
	}()
	Writer()
}
