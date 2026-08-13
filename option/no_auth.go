package option

import (
	core "github.com/cohere-ai/cohere-go/v2/core"
)

// WithoutToken disables authentication, so that no 'Authorization' header is sent.
//
// Use it to point the client at a proxy or a self-hosted deployment that performs its own
// authentication, e.g. client.NewClient(option.WithoutToken()). Unlike WithToken(""), this also
// prevents the client from falling back to the CO_API_KEY environment variable.
//
// Combine it with WithHTTPClient by passing WithoutToken() last, so that the custom HTTP client is
// the one whose requests have the header removed.
func WithoutToken() *core.NoAuthOption {
	return &core.NoAuthOption{}
}
