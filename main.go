package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

type ConfigT struct {
	Url string
	Key string
}

type SearchQuery struct {
	Page  int    `json:"page"`
	Query string `json:"query"`
}

type CountResp struct {
	Count int `json:"count"`
}

var Config = ConfigT{
	Url: "http://localhost:8000",
	Key: os.Getenv("OVERCAST_API_KEY"),
}

var Output = make(chan any)

func writer() {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for res := range Output {
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func app() {
	defer close(Output)

	var initKey = func(key string) {
		log.Println(key, Config.Key)
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

	switch os.Args[1] {
	case "search":
		searchCmd.Parse(os.Args[2:])
		initKey(*searchKey)
		if *searchAll {
			*searchPage = -1
		}
		err := Search(*searchPage, searchCmd.Args()[0])
		if err != nil {
			log.Println(err)
		}
	case "overview":
		overviewCmd.Parse(os.Args[2:])
		initKey(*overviewKey)
		err := AssetOverview(overviewCmd.Args()[0])
		if err != nil {
			log.Println(err)
		}
	default:
		log.Fatalf("subcommands: search, overview")
	}
}

func main() {
	go app()
	writer()
}
