package core

import (
	http "net/http"
)

// NoAuthOption implements the RequestOption interface, and removes the
// 'Authorization' header from every request issued by the client.
//
// Use option.WithoutToken() to construct this option.
type NoAuthOption struct{}

func (n *NoAuthOption) applyRequestOptions(opts *RequestOptions) {
	opts.Token = ""
	if _, ok := opts.HTTPClient.(*noAuthHTTPClient); ok {
		return
	}
	opts.HTTPClient = &noAuthHTTPClient{delegate: opts.HTTPClient}
}

// noAuthHTTPClient strips the 'Authorization' header before delegating the request.
//
// The header is removed at request time (rather than by leaving the token empty) so that the
// CO_API_KEY environment variable, which the generated clients fall back to, cannot reintroduce it.
type noAuthHTTPClient struct {
	delegate HTTPClient
}

func (c *noAuthHTTPClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Del("Authorization")
	if c.delegate == nil {
		return http.DefaultClient.Do(req)
	}
	return c.delegate.Do(req)
}
