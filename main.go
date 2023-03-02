package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

var Output = make(chan any)

type SearchQuery struct {
	Page  int    `json:"page"`
	Query string `json:"query"`
}

type CountResp struct {
	Count int `json:"count"`
}

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
	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
	searchAll := flag.Bool("all", false, "get all pages")
	searchPage := flag.Int("page", 1, "get page")

	overviewCmd := flag.NewFlagSet("overview", flag.ExitOnError)
	//overviewOrg := flag.Bool("org", false, "organization overview (replaces domain)")

	switch os.Args[1] {
	case "search":
		searchCmd.Parse(os.Args[2:])
		if *searchAll {
			*searchPage = -1
		}
		err := Search(*searchPage, searchCmd.Args()[0])
		if err != nil {
			log.Println(err)
		}
	case "overview":
		overviewCmd.Parse(os.Args[2:])
		err := AssetOverview(overviewCmd.Args()[0])
		if err != nil {
			log.Println(err)
		}
	}

	close(Output)
}

func main() {
	go app()
	writer()
}
