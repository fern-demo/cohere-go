package client

import (
	core "github.com/cohere-ai/cohere-go/v2/core"
	option "github.com/cohere-ai/cohere-go/v2/option"
)

// NewClientWithoutAuth returns a *Client that sends no 'Authorization' header.
//
// Use it to point the client at a proxy or a self-hosted deployment that performs its own
// authentication:
//
//	co := client.NewClientWithoutAuth(
//		option.WithBaseURL("https://my-proxy.example.com"),
//	)
//
// A custom HTTP client is preserved rather than replaced: its requests have the header removed.
//
//	co := client.NewClientWithoutAuth(
//		option.WithBaseURL("https://my-proxy.example.com"),
//		option.WithHTTPClient(myClient),
//	)
//
// Option order does not matter, and unlike option.WithToken(""), the CO_API_KEY environment
// variable cannot reintroduce the header.
func NewClientWithoutAuth(opts ...option.RequestOption) *Client {
	// Resolve the caller's options first so that any custom HTTP client is wrapped rather than
	// overwritten, then apply the wrapper last so no option ordering can undo it.
	resolved := core.NewRequestOptions(opts...)
	return NewClient(
		append(opts, option.WithHTTPClient(core.NoAuthClient(resolved.HTTPClient)))...,
	)
}
