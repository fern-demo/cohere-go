package tests

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	cohere "github.com/cohere-ai/cohere-go/v2"
	client "github.com/cohere-ai/cohere-go/v2/client"
	option "github.com/cohere-ai/cohere-go/v2/option"
	"github.com/stretchr/testify/require"
)

type recordingHTTPClient struct {
	header http.Header
	calls  int
}

func (r *recordingHTTPClient) Do(req *http.Request) (*http.Response, error) {
	r.calls++
	r.header = req.Header.Clone()
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"text":"hi"}`)),
	}, nil
}

func chatWith(t *testing.T, co *client.Client) {
	t.Helper()
	_, err := co.Chat(context.TODO(), &cohere.ChatRequest{Message: "hi"})
	require.NoError(t, err)
}

// chat issues a request through client.NewClient with a recording HTTP client installed first.
func chat(t *testing.T, opts ...option.RequestOption) http.Header {
	t.Helper()
	recorder := &recordingHTTPClient{}
	co := client.NewClient(append([]option.RequestOption{option.WithHTTPClient(recorder)}, opts...)...)
	chatWith(t, co)
	return recorder.header
}

func TestNewClientWithoutAuthOmitsAuthorizationHeader(t *testing.T) {
	t.Setenv("CO_API_KEY", "env-token")
	recorder := &recordingHTTPClient{}
	co := client.NewClientWithoutAuth(option.WithHTTPClient(recorder))
	chatWith(t, co)
	require.Empty(t, recorder.header.Get("Authorization"))
}

// A token supplied alongside the constructor is still suppressed: the point of the constructor is
// that this client never authenticates.
func TestNewClientWithoutAuthOmitsAuthorizationHeaderEvenWithAToken(t *testing.T) {
	recorder := &recordingHTTPClient{}
	co := client.NewClientWithoutAuth(
		option.WithHTTPClient(recorder),
		option.WithToken("some-token"),
	)
	chatWith(t, co)
	require.Empty(t, recorder.header.Get("Authorization"))
}

// A custom HTTP client must be wrapped, not replaced, so that a caller's proxy, mTLS config,
// timeouts and instrumentation continue to apply.
func TestNewClientWithoutAuthPreservesCustomHTTPClient(t *testing.T) {
	t.Setenv("CO_API_KEY", "env-token")
	recorder := &recordingHTTPClient{}
	co := client.NewClientWithoutAuth(option.WithHTTPClient(recorder))
	chatWith(t, co)
	require.Equal(t, 1, recorder.calls, "the caller's HTTP client must receive the request")
}

// No ordering of the options may reintroduce the header, in either direction.
func TestNewClientWithoutAuthIsOrderIndependent(t *testing.T) {
	t.Setenv("CO_API_KEY", "env-token")
	for _, tt := range []struct {
		name string
		opts func(recorder *recordingHTTPClient) []option.RequestOption
	}{
		{
			name: "http client first",
			opts: func(recorder *recordingHTTPClient) []option.RequestOption {
				return []option.RequestOption{
					option.WithHTTPClient(recorder),
					option.WithToken("some-token"),
				}
			},
		},
		{
			name: "token first",
			opts: func(recorder *recordingHTTPClient) []option.RequestOption {
				return []option.RequestOption{
					option.WithToken("some-token"),
					option.WithHTTPClient(recorder),
				}
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			recorder := &recordingHTTPClient{}
			co := client.NewClientWithoutAuth(tt.opts(recorder)...)
			chatWith(t, co)
			require.Empty(t, recorder.header.Get("Authorization"))
			require.Equal(t, 1, recorder.calls, "the caller's HTTP client must receive the request")
		})
	}
}

// TestEmptyTokenOmitsAuthorizationHeaderWhenEnvironmentIsUnset pins the behavior of
// WithToken("") only for the case where CO_API_KEY is also empty. The t.Setenv call is what makes
// the assertion hold, so it is deliberate rather than incidental: see the test below for what
// WithToken("") does when the environment variable is actually set.
func TestEmptyTokenOmitsAuthorizationHeaderWhenEnvironmentIsUnset(t *testing.T) {
	t.Setenv("CO_API_KEY", "")
	require.Empty(t, chat(t, option.WithToken("")).Get("Authorization"))
}

// TestEmptyTokenFallsBackToEnvironmentVariable documents that WithToken("") does not disable
// authentication in this SDK: client.NewClient replaces an empty token with CO_API_KEY, so the
// header is still sent. NewClientWithoutAuth is the only way to send no header at all.
func TestEmptyTokenFallsBackToEnvironmentVariable(t *testing.T) {
	t.Setenv("CO_API_KEY", "env-token")
	require.Equal(t, "Bearer env-token", chat(t, option.WithToken("")).Get("Authorization"))
}

func TestTokenIsSentWhenProvided(t *testing.T) {
	require.Equal(t, "Bearer some-token", chat(t, option.WithToken("some-token")).Get("Authorization"))
}
