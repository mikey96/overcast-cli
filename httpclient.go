package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

func ApiGet[T any](path string, to T) (int, T, error) {
	var res T
	resp, err := authReq(
		"GET",
		endpoint(path),
		"application/json",
		nil,
	)
	if err != nil {
		return resp.StatusCode, res, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, res, err
	}

	err = json.NewDecoder(bytes.NewReader(body)).Decode(&res)
	if err != nil {
		var message MessageResp
		err = json.NewDecoder(bytes.NewReader(body)).Decode(&message)
		// Stderr <- message.Message
		return resp.StatusCode, res, errors.New(message.Message)
	}
	return resp.StatusCode, res, err
}

func ApiPost[T any](path string, data any, to T) (int, T, error) {
	var res T
	bslc, err := json.Marshal(data)
	if err != nil {
		return 400, res, err
	}
	resp, err := authReq(
		"POST",
		endpoint(path),
		"application/json",
		bytes.NewReader(bslc),
	)
	if err != nil {
		return resp.StatusCode, res, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, res, err
	}

	err = json.NewDecoder(bytes.NewReader(body)).Decode(&res)
	if err != nil || resp.StatusCode != 200 {
		var message MessageResp
		err = json.NewDecoder(bytes.NewReader(body)).Decode(&message)
		// Stderr <- message.Message
		return resp.StatusCode, res, errors.New(message.Message)
	}
	return resp.StatusCode, res, err
}
