package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

func endpoint(path string) string {
	return fmt.Sprintf("%s%s", Config.Url, path)
}

func authReq(method string, url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-type", contentType)
	req.Header.Set("Authorization", fmt.Sprintf("Key %s", Config.Key))
	resp, err := http.DefaultClient.Do(req)
	return resp, err
}

func ApiGet[T any](path string, to T) (T, error) {
	var res T
	resp, err := authReq(
		"GET",
		endpoint(path),
		"application/json",
		nil,
	)
	if err != nil {
		return res, err
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	return res, err
}

func ApiPost[T any](path string, data any, to T) (T, error) {
	var res T
	bslc, err := json.Marshal(data)
	if err != nil {
		return res, err
	}
	resp, err := authReq(
		"POST",
		endpoint(path),
		"application/json",
		bytes.NewReader(bslc),
	)
	if err != nil {
		return res, err
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	return res, err
}

func count(queryString string) (int, error) {
	resp, err := ApiPost("/search/count", SearchQuery{Query: queryString}, CountResp{})
	return resp.Count, err
}

func AssetOverview(domain string) error {
	var resp map[string]any
	resp, err := ApiGet(fmt.Sprintf("/overview/asset?domain=%s", domain), resp)
	if err != nil {
		return err
	}
	Output <- resp
	return nil
}

func Search(page int, queryString string) error {
	var searchPage = func(p int) error {
		query := SearchQuery{Page: p, Query: queryString}
		resp, err := ApiPost("/search", query, any(1))
		if err != nil {
			return err
		}
		Output <- resp
		return nil
	}

	count, err := count(queryString)
	if err != nil {
		return err
	}

	pages := (count / 100) + 1
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
