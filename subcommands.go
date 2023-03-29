package main

import (
	"fmt"
	"sync"
)

func SearchPage(page int, query string) (any, error) {
	req := SearchReq{Page: page, Query: query}
	status, resp, err := ApiPost("/search/distinct-subdomains", req, any(1))
	if status != 200 {
		return resp, fmt.Errorf("Request error, status: {}", status)
	}
	return resp, err
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
	status, resp, err := ApiPost("/search/count", SearchReq{Query: queryString}, CountResp{})
	if status != 200 {
		return resp.Count, fmt.Errorf("Request error, status: {}", status)
	}
	return resp.Count, err
}

func Search(page int, queryString string) error {
	if page >= 0 {
		resp, err := SearchPage(page, queryString)
		if err == nil {
			StdoutJson <- resp
		}
		return err
	} else {
		return SearchAllPages(queryString, func(resp any, err error) {
			if err == nil {
				StdoutJson <- resp
			}
		})
	}
}

func Subdomains(root string) error {
	var seen sync.Map
	var unique = func(it string) bool {
		_, present := seen.Load(it)
		if present {
			return false
		}
		seen.Store(it, true)
		return true
	}

	return SearchAllPages(root, func(resp any, err error) {
		if err == nil {
			r, ok := resp.([]any); if ok {
				for _, row := range r {
					res, ok := row.(map[string]any)["subdomain"]; if ok {
						if unique(res.(string)) {
							Stdout <- res.(string)
						}
					}
				}
			}
		}
	})
}

func Metadata(queryString string) error {
	query := SearchReq{Page: 1, Query: queryString}
	status, resp, err := ApiPost("/search/distinct-subdomains/metadata", query, any(1))
	if err != nil {
		return err
	}
	if status != 200 {
		Stderr <- fmt.Errorf("Request error, status: {}", status).Error()
	}
	StdoutJson <- resp
	return nil
}

func Overview(domain string) error {
	var respT map[string]any
	var sources = map[string]func(string) (int, any, error){
		"asset": func(domain string) (int, any, error) {
			return ApiGet(fmt.Sprintf("/overview/asset?domain=%s", domain), respT)
		},
		"service": func(domain string) (int, any, error) {
			return ApiGet(fmt.Sprintf("/overview/services?domain=%s", domain), respT)
		},
		"risk": func(domain string) (int, any, error) {
			return ApiGet(fmt.Sprintf("/overview/risks?domain=%s", domain), respT)
		},
	}

	// concurrently evaluate the above data
	var wg sync.WaitGroup
	var tmp = sync.Map{}
	wg.Add(len(sources))
	for k, v := range sources {
		go func(key string, src func(string) (int, any, error)) {
			defer wg.Done()
			status, ov, err := src(domain)
			if status != 200 && err == nil {
				Stderr <- fmt.Errorf("Request error, status: {}", status).Error()
				return;
			}
			if err != nil {
				Stderr <- err.Error()
				return;
			}
			tmp.Store(key, ov)
		}(k, v)
	}
	wg.Wait()

	// sync map to map[string]any
	var resp = OverviewResp{}
	tmp.Range(func(k any, v any) bool {
		resp[k.(string)] = v
		return true
	})

	StdoutJson <- resp
	return nil
}
