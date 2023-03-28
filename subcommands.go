package main

import (
	"fmt"
	"sync"
)

func SearchPage(page int, query string) (any, error) {
	req := SearchReq{Page: page, Query: query}
	return ApiPost("/search/distinct-subdomains", req, any(1))
}

func SearchAllPages(query string, callback func(any, error)) error {
	c, err := CountSearch(query)
	if err != nil {return err}
	pages := (c / 100) + 1
	var wg sync.WaitGroup
	for i := 1; i <= pages; i++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			callback(SearchPage(p, query))
		}(i)
	}
	wg.Wait()
	return nil
}

func CountSearch(queryString string) (int, error) {
	resp, err := ApiPost("/search/count", SearchReq{Query: queryString}, CountResp{})
	return resp.Count, err
}

func Search(page int, queryString string) error {
	if page >= 0 {
		resp, err := SearchPage(page, queryString)
		if err == nil {
			JsonOutput <- resp
		}
		return err
	} else {
		return SearchAllPages(queryString, func(resp any, err error) {
			if err == nil {
				JsonOutput <- resp
			}
		})
	}
}

func Subdomains(root string) error {
	return SearchAllPages(root, func(resp any, err error) {
		if err == nil {
			for _, row := range resp.([]any) {
				res, ok := row.(map[string]any)["subdomain"]; if ok {
					Output <- res.(string)
				}
			}
		}
	})
}

func Metadata(queryString string) error {
	query := SearchReq{Page: 1, Query: queryString}
	resp, err := ApiPost("/search/distinct-subdomains/metadata", query, any(1))
	if err != nil {
		return err
	}
	JsonOutput <- resp
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

	JsonOutput <- resp
	return nil
}
