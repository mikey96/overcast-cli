package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var Config = ConfigT{
	Url: "https://api.enterprise.overcast-security.app",
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

	searchCmd := flag.NewFlagSet("subdomains-search", flag.ExitOnError)
	searchKey := keyFlag(searchCmd)

	metadataCmd := flag.NewFlagSet("subdomains-metadata", flag.ExitOnError)
	metadataKey := keyFlag(metadataCmd)

	/*verviewCmd := flag.NewFlagSet("overview", flag.ExitOnError)
	overviewKey := keyFlag(overviewCmd)*/

	subsCmd := flag.NewFlagSet("subdomains", flag.ExitOnError)
	subsKey := keyFlag(subsCmd)

	ipsSearchCmd := flag.NewFlagSet("ips-search", flag.ExitOnError)
	ipsSearchKey := keyFlag(ipsSearchCmd)

	ipsMetadataCmd := flag.NewFlagSet("ips-metadata", flag.ExitOnError)
	ipMetadataKey := keyFlag(ipsMetadataCmd)

	/*verviewCmd := flag.NewFlagSet("overview", flag.ExitOnError)
	overviewKey := keyFlag(overviewCmd)*/

	ipsCmd := flag.NewFlagSet("ips", flag.ExitOnError)
	ipsKey := keyFlag(ipsCmd)

	if len(os.Args) <= 1 {
		fmt.Print(help())
		return
	}

	switch os.Args[1] {
	case "help":
		fmt.Print(help())
	case "subdomains":
		subsCmd.Parse(os.Args[2:])
		initKey(*subsKey)
		Subdomains(QueryStringArgs(subsCmd.Args()))
	case "subdomains-search":
		searchCmd.Parse(os.Args[2:])
		initKey(*searchKey)
		SearchSubs(-1, QueryStringArgs(searchCmd.Args()))
	case "subdomains-metadata":
		metadataCmd.Parse(os.Args[2:])
		initKey(*metadataKey)
		MetadataSubs(QueryStringArgs(metadataCmd.Args()))
	case "ips":
		ipsCmd.Parse(os.Args[2:])
		initKey(*ipsKey)
		IPs(QueryStringArgs(ipsCmd.Args()))
	case "ips-search":
		ipsSearchCmd.Parse(os.Args[2:])
		initKey(*ipsSearchKey)
		SearchIPs(-1, QueryStringArgs(ipsSearchCmd.Args()))
	case "ips-metadata":
		ipsMetadataCmd.Parse(os.Args[2:])
		initKey(*ipMetadataKey)
		MetadataIPs(QueryStringArgs(ipsMetadataCmd.Args()))
	/*case "overview":
	overviewCmd.Parse(os.Args[2:])
	initKey(*overviewKey)
	Overview(overviewCmd.Args()[0])*/
	default:
		fmt.Print(help())
	}
}

func help() string {
	return fmt.Sprintln(
		`Subcommands:
  - search-subdomains
    - Options: -key
    - Arguments: query string
    Get a list of assets

  - subdomains
    - Options: -key
    - Arguments: query string
    Get subdomains in a single list.

  - metadata
    - Options: -key
    - Arguments: query string
    Get metadata for a search

  - help
    View this page

Links:
  - Get api key  https://enterprise.overcast-security.app/account
  - Github repo  https://github.com/mikey96/overcast-cli/`)
}

func main() {
	go func() {
		defer CloseWriter()
		app()
	}()
	Writer()
}
