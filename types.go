package main

import "fmt"

type ConfigT struct {
	Url string
	Key string
}

type OverviewResp map[string]any

type SearchReq struct {
	Page  int    `json:"page"`
	Query string `json:"query"`
}

type MetadataResp map[string]map[string]any

type CountResp struct {
	Count int `json:"count"`
}

func QueryStringArgs(args []string) string {
	var qstr string
	for _, arg := range args {
		qstr = fmt.Sprintf("%s %s", qstr, arg)
	}
	return qstr
}
