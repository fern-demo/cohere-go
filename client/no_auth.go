package client

import (
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
// It is equivalent to passing option.WithToken("") and exists so the intent is explicit at the
// call site. The CO_API_KEY environment variable cannot reintroduce the header either way.
func NewClientWithoutAuth(opts ...option.RequestOption) *Client {
	// Applied last so no caller-supplied token can override it.
	return NewClient(append(opts, option.WithToken(""))...)
}
