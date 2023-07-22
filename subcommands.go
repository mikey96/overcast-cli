package main

import (
	"fmt"
	"sync"
)

func SearchPageSubs(page int, query string) (any, error) {
	req := SearchReq{Page: page, Query: query}
	status, resp, err := ApiPost("/search/distinct-subdomains", req, any(1))
	if status != 200 && err == nil {
		return resp, fmt.Errorf("Request error, status: %d", status)
	}
	return resp, err
}

func SearchAllPagesSubs(query string, callback func(any, error)) error {
	c, err := CountSearchSubs(query)
	if err != nil {
		return err
	}
	pages := (c / 100) + 1
	var wg sync.WaitGroup
	for i := 1; i <= pages; i++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			callback(SearchPageSubs(p, query))
		}(i)
	}
	wg.Wait()
	return nil
}

func CountSearchSubs(queryString string) (int, error) {
	status, resp, err := ApiPost("/search/distinct-subdomains/count", SearchReq{Query: queryString}, CountResp{})
	if status != 200 && err == nil {
		return resp.Count, fmt.Errorf("Request error, status: %d", status)
	}
	return resp.Count, err
}

func SearchPageIPs(page int, query string) (any, error) {
	req := SearchReq{Page: page, Query: query}
	status, resp, err := ApiPost("/search/distinct-ips", req, any(1))
	if status != 200 && err == nil {
		return resp, fmt.Errorf("Request error, status: %d", status)
	}
	return resp, err
}

func SearchAllPagesIPs(query string, callback func(any, error)) error {
	c, err := CountSearchIPs(query)
	if err != nil {
		return err
	}
	pages := (c / 100) + 1
	var wg sync.WaitGroup
	for i := 1; i <= pages; i++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			callback(SearchPageIPs(p, query))
		}(i)
	}
	wg.Wait()
	return nil
}

func CountSearchIPs(queryString string) (int, error) {
	status, resp, err := ApiPost("/search/distinct-ips/count", SearchReq{Query: queryString}, CountResp{})
	if status != 200 && err == nil {
		return resp.Count, fmt.Errorf("Request error, status: %d", status)
	}
	return resp.Count, err
}

func SearchSubs(page int, queryString string) error {
	if page >= 0 {
		resp, err := SearchPageSubs(page, queryString)
		if err != nil {
			Stderr <- err.Error()
			return err
		}
		StdoutJson <- resp
		return err
	} else {
		return SearchAllPagesSubs(queryString, func(resp any, err error) {
			if err != nil {
				Stderr <- err.Error()
				return
			}
			StdoutJson <- resp
		})
	}
}

func SearchIPs(page int, queryString string) error {
	if page >= 0 {
		resp, err := SearchPageIPs(page, queryString)
		if err != nil {
			Stderr <- err.Error()
			return err
		}
		StdoutJson <- resp
		return err
	} else {
		return SearchAllPagesIPs(queryString, func(resp any, err error) {
			if err != nil {
				Stderr <- err.Error()
				return
			}
			StdoutJson <- resp
		})
	}
}

func Subdomains(queryString string) error {
	return SearchAllPagesSubs(queryString, func(resp any, err error) {
		if err != nil {
			Stderr <- err.Error()
			return
		}
		r, ok := resp.([]any)
		if ok {
			for _, row := range r {
				res, ok := row.(map[string]any)["subdomain"]
				if ok {
					fmt.Println(res)
				}
			}
		}
	})
}

func IPs(queryString string) error {
	return SearchAllPagesIPs(queryString, func(resp any, err error) {
		if err != nil {
			Stderr <- err.Error()
			return
		}
		r, ok := resp.([]any)
		if ok {
			for _, row := range r {
				res, ok := row.(map[string]any)["subdomain"]
				if ok {
					fmt.Println(res)
				}
			}
		}
	})
}

func MetadataSubs(queryString string) error {
	query := SearchReq{Page: 1, Query: queryString}
	status, resp, err := ApiPost("/search/distinct-subdomains/metadata", query, any(1))
	fmt.Println(resp)
	if err != nil {
		return err
	}
	if status != 200 {
		Stderr <- fmt.Errorf("Request error, status: %d", status).Error()
	}
	StdoutJson <- resp
	return nil
}

func MetadataIPs(queryString string) error {
	query := SearchReq{Page: 1, Query: queryString}
	status, resp, err := ApiPost("/search/distinct-ips/metadata", query, any(1))
	if err != nil {
		return err
	}
	if status != 200 {
		Stderr <- fmt.Errorf("Request error, status: %d", status).Error()
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
				Stderr <- fmt.Errorf("Request error, status: %d", status).Error()
				return
			}
			if err != nil {
				Stderr <- err.Error()
				return
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
