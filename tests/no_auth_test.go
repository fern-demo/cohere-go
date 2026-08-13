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
}

func (r *recordingHTTPClient) Do(req *http.Request) (*http.Response, error) {
	r.header = req.Header.Clone()
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"text":"hi"}`)),
	}, nil
}

func chat(t *testing.T, opts ...option.RequestOption) http.Header {
	t.Helper()
	recorder := &recordingHTTPClient{}
	co := client.NewClient(append([]option.RequestOption{option.WithHTTPClient(recorder)}, opts...)...)
	_, err := co.Chat(context.TODO(), &cohere.ChatRequest{Message: "hi"})
	require.NoError(t, err)
	return recorder.header
}

func TestWithoutTokenOmitsAuthorizationHeader(t *testing.T) {
	t.Setenv("CO_API_KEY", "env-token")
	require.Empty(t, chat(t, option.WithoutToken()).Get("Authorization"))
	require.Empty(t, chat(t, option.WithToken("some-token"), option.WithoutToken()).Get("Authorization"))
}

func TestEmptyTokenOmitsAuthorizationHeader(t *testing.T) {
	t.Setenv("CO_API_KEY", "")
	require.Empty(t, chat(t, option.WithToken("")).Get("Authorization"))
}

func TestTokenIsSentWhenProvided(t *testing.T) {
	require.Equal(t, "Bearer some-token", chat(t, option.WithToken("some-token")).Get("Authorization"))
}
