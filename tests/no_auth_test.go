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

func chatWith(t *testing.T, co *client.Client, recorder *recordingHTTPClient) http.Header {
	t.Helper()
	_, err := co.Chat(context.TODO(), &cohere.ChatRequest{Message: "hi"})
	require.NoError(t, err)
	require.Equal(t, 1, recorder.calls, "the caller's HTTP client must receive the request")
	return recorder.header
}

// chat issues a request through client.NewClient with a recording HTTP client installed first.
func chat(t *testing.T, opts ...option.RequestOption) http.Header {
	t.Helper()
	recorder := &recordingHTTPClient{}
	co := client.NewClient(append([]option.RequestOption{option.WithHTTPClient(recorder)}, opts...)...)
	return chatWith(t, co, recorder)
}

// An explicitly empty token means "send no Authorization header", and is distinct from never
// setting a token at all. The environment variable cannot reintroduce the header.
func TestEmptyTokenOmitsAuthorizationHeader(t *testing.T) {
	t.Setenv("CO_API_KEY", "env-token")
	require.Empty(t, chat(t, option.WithToken("")).Get("Authorization"))
}

// The counterpart of the test above: omitting the option entirely still falls back to CO_API_KEY.
// These two together are what make WithToken("") meaningful rather than a no-op.
func TestOmittingTheTokenStillFallsBackToEnvironmentVariable(t *testing.T) {
	t.Setenv("CO_API_KEY", "env-token")
	require.Equal(t, "Bearer env-token", chat(t).Get("Authorization"))
}

func TestEmptyTokenOmitsAuthorizationHeaderWhenEnvironmentIsUnset(t *testing.T) {
	t.Setenv("CO_API_KEY", "")
	require.Empty(t, chat(t, option.WithToken("")).Get("Authorization"))
}

// An empty token wins regardless of where it appears in the option list.
func TestEmptyTokenIsOrderIndependent(t *testing.T) {
	t.Setenv("CO_API_KEY", "env-token")
	require.Empty(t, chat(t, option.WithToken("real-token"), option.WithToken("")).Get("Authorization"))
}

func TestNewClientWithoutAuthOmitsAuthorizationHeader(t *testing.T) {
	t.Setenv("CO_API_KEY", "env-token")
	recorder := &recordingHTTPClient{}
	co := client.NewClientWithoutAuth(option.WithHTTPClient(recorder))
	require.Empty(t, chatWith(t, co, recorder).Get("Authorization"))
}

// The constructor suppresses auth even if a token is supplied alongside it.
func TestNewClientWithoutAuthOmitsAuthorizationHeaderEvenWithAToken(t *testing.T) {
	recorder := &recordingHTTPClient{}
	co := client.NewClientWithoutAuth(
		option.WithHTTPClient(recorder),
		option.WithToken("some-token"),
	)
	require.Empty(t, chatWith(t, co, recorder).Get("Authorization"))
}

// A custom HTTP client must still be the one issuing requests: suppressing auth must not swap out
// a caller's proxy, mTLS config, timeouts or instrumentation. chatWith asserts recorder.calls.
func TestNewClientWithoutAuthPreservesCustomHTTPClient(t *testing.T) {
	t.Setenv("CO_API_KEY", "env-token")
	recorder := &recordingHTTPClient{}
	co := client.NewClientWithoutAuth(option.WithHTTPClient(recorder))
	chatWith(t, co, recorder)
}

func TestTokenIsSentWhenProvided(t *testing.T) {
	require.Equal(t, "Bearer some-token", chat(t, option.WithToken("some-token")).Get("Authorization"))
}

// Per-request auth suppression is NOT supported: the client-level Authorization header is already
// in place by the time request options are merged, and MergeHeaders cannot clear it. Pinned here
// so the limitation is explicit rather than a surprise. Build a separate client with
// NewClientWithoutAuth instead.
func TestPerRequestEmptyTokenDoesNotSuppressClientLevelAuth(t *testing.T) {
	t.Setenv("CO_API_KEY", "env-token")
	recorder := &recordingHTTPClient{}
	co := client.NewClient(option.WithHTTPClient(recorder))
	_, err := co.Chat(context.TODO(), &cohere.ChatRequest{Message: "hi"}, option.WithToken(""))
	require.NoError(t, err)
	require.Equal(t, "Bearer env-token", recorder.header.Get("Authorization"))
}
