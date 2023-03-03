package main

import (
	"fmt"
	"sync"
)

func Search(page int, queryString string) error {
	var count = func(queryString string) (int, error) {
		resp, err := ApiPost("/search/count", SearchReq{Query: queryString}, CountResp{})
		return resp.Count, err
	}

	var searchPage = func(p int) error {
		query := SearchReq{Page: p, Query: queryString}
		resp, err := ApiPost("/search", query, any(1))
		if err != nil {
			return err
		}
		Output <- resp
		return nil
	}

	c, err := count(queryString)
	if err != nil {
		return err
	}

	pages := (c / 100) + 1
	if page >= 0 {
		if page > pages {
			return fmt.Errorf("Last page is %d", pages)
		}
		return searchPage(page)
	}

	var wg sync.WaitGroup
	for i := 1; i <= pages; i++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			searchPage(p)
		}(i)
	}
	wg.Wait()
	return nil
}

func Metadata(queryString string) error {
	query := SearchReq{Page: 1, Query: queryString}
	resp, err := ApiPost("/search/distinct-subdomains/metadata", query, any(1))
	if err != nil {
		return err
	}
	Output <- resp
	return nil
}

func Overview(domain string) error {
	var respT map[string]any
	var sources = map[string]func(string) (any, error){
		"asset": func(domain string) (any, error) {
			return ApiGet(fmt.Sprintf("/overview/asset?domain=%s", domain), respT)
		},
		"service": func(domain string) (any, error) {
			return ApiGet(fmt.Sprintf("/overview/services?domain=%s", domain), respT)
		},
		"risk": func(domain string) (any, error) {
			return ApiGet(fmt.Sprintf("/overview/risks?domain=%s", domain), respT)
		},
	}

	// concurrently evaluate the above data
	var wg sync.WaitGroup
	var tmp = sync.Map{}
	wg.Add(len(sources))
	for k, v := range sources {
		go func(key string, src func(string) (any, error)) {
			defer wg.Done()
			ov, err := src(domain)
			if err == nil {
				tmp.Store(key, ov)
			}
		}(k, v)
	}
	wg.Wait()

	// sync map to map[string]any
	var resp = OverviewResp{}
	tmp.Range(func(k any, v any) bool {
		resp[k.(string)] = v
		return true
	})

	Output <- resp
	return nil
}
