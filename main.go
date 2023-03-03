package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"os"
)

var Config = ConfigT{
	Url: "http://localhost:8000",
	// this is overwritten by -key flag
	Key: os.Getenv("OVERCAST_API_KEY"),
}

var Output = make(chan any)

func writer() {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for res := range Output {
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Println(err)
		}
	}
}

func app() {
	defer close(Output)

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

	overviewCmd := flag.NewFlagSet("overview", flag.ExitOnError)
	overviewKey := keyFlag(overviewCmd)

	metadataCmd := flag.NewFlagSet("metadata", flag.ExitOnError)
	metadataKey := keyFlag(metadataCmd)

	switch os.Args[1] {
	case "search":
		searchCmd.Parse(os.Args[2:])
		initKey(*searchKey)
		if *searchAll {
			*searchPage = -1
		}
		err := Search(*searchPage, QueryStringArgs(searchCmd.Args()))
		if err != nil {
			log.Println(err)
		}
	case "metadata":
		metadataCmd.Parse(os.Args[2:])
		initKey(*metadataKey)
		err := Metadata(QueryStringArgs(metadataCmd.Args()))
		if err != nil {
			log.Println(err)
		}
	case "overview":
		overviewCmd.Parse(os.Args[2:])
		initKey(*overviewKey)
		err := Overview(overviewCmd.Args()[0])
		if err != nil {
			log.Println(err)
		}
	default:
		log.Println("subcommands: search, overview")
	}
}

func main() {
	go app()
	writer()
}
