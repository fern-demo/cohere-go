package core

import (
	http "net/http"
)

// NoAuthClient returns an HTTPClient that removes the 'Authorization' header from every request
// before delegating to client.
//
// The header is stripped at request time rather than by leaving the token empty, because
// client.NewClient re-populates an empty token from the CO_API_KEY environment variable. A nil
// client delegates to http.DefaultClient, matching the generated caller's own default.
//
// Wrapping is idempotent: passing an already-wrapped client returns it unchanged.
//
// Use client.NewClientWithoutAuth rather than calling this directly.
func NoAuthClient(client HTTPClient) HTTPClient {
	if _, ok := client.(*noAuthHTTPClient); ok {
		return client
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &noAuthHTTPClient{delegate: client}
}

type noAuthHTTPClient struct {
	delegate HTTPClient
}

func (c *noAuthHTTPClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Del("Authorization")
	return c.delegate.Do(req)
}
