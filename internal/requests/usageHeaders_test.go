package requests

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/stretchr/testify/require"
)

func recordingServer(t *testing.T) (*httptest.Server, *[]http.Header) {
	t.Helper()
	var seen []http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Header.Clone())
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &seen
}

func newTestClient(t *testing.T) *Client {
	t.Helper()
	client, err := NewKosliClient("", 0, false, logger.NewStandardLogger())
	require.NoError(t, err)
	return client
}

var testUsageHeaders = map[string]string{
	"X-Kosli-Client":        "cli/v0.0.0-test",
	"X-Kosli-Operation":     "attest generic",
	"X-Kosli-Invocation-Id": "0b6c2b5e-9a8e-4c43-8d2b-3f0c3f6a1f11",
	"X-Kosli-Cli-Flags":     "flow,name",
	"X-Kosli-Ci":            "github",
}

func TestUsageHeadersAreSentToTheKosliHost(t *testing.T) {
	kosli, seen := recordingServer(t)
	client := newTestClient(t)
	client.SetUsageHeaders(kosli.URL, testUsageHeaders)

	for range 2 {
		_, err := client.Do(&RequestParams{Method: http.MethodGet, URL: kosli.URL + "/api/v2/flows/acme"})
		require.NoError(t, err)
	}

	require.Len(t, *seen, 2)
	for _, headers := range *seen {
		for name, want := range testUsageHeaders {
			require.Equal(t, want, headers.Get(name), name)
		}
		require.Contains(t, headers.Get("User-Agent"), "Kosli/")
	}
}

func TestUsageHeadersAreNotSentToAnotherHost(t *testing.T) {
	kosli, _ := recordingServer(t)
	other, seen := recordingServer(t)
	client := newTestClient(t)
	client.SetUsageHeaders(kosli.URL, testUsageHeaders)

	_, err := client.Do(&RequestParams{Method: http.MethodGet, URL: other.URL + "/v2/registry/manifests/latest"})
	require.NoError(t, err)

	require.Len(t, *seen, 1)
	for name := range testUsageHeaders {
		require.Empty(t, (*seen)[0].Values(name), name)
	}
}

func TestUsageHeadersAreNotSentWhenNeverSet(t *testing.T) {
	kosli, seen := recordingServer(t)
	client := newTestClient(t)

	_, err := client.Do(&RequestParams{Method: http.MethodGet, URL: kosli.URL})
	require.NoError(t, err)

	for name := range testUsageHeaders {
		require.Empty(t, (*seen)[0].Values(name), name)
	}
}

func TestUsageHeadersAreDisabledByAnUnparseableKosliHost(t *testing.T) {
	kosli, seen := recordingServer(t)
	client := newTestClient(t)
	client.SetUsageHeaders(kosli.URL, testUsageHeaders)
	client.SetUsageHeaders("://not a url", testUsageHeaders)

	_, err := client.Do(&RequestParams{Method: http.MethodGet, URL: kosli.URL})
	require.NoError(t, err)

	for name := range testUsageHeaders {
		require.Empty(t, (*seen)[0].Values(name), name)
	}
}

func TestSameHost(t *testing.T) {
	for _, tt := range []struct {
		a, b string
		want bool
	}{
		{"https://app.kosli.com", "https://app.kosli.com/api/v2/flows/acme", true},
		{"https://APP.kosli.com", "https://app.kosli.com/api", true},
		{"https://app.kosli.com", "https://app.kosli.com:443/api", true},
		{"http://localhost:8001", "http://localhost:8001/api", true},
		{"http://localhost:8001", "http://localhost:8002/api", false},
		{"https://app.kosli.com", "http://app.kosli.com/api", false},
		{"https://app.kosli.com", "https://app.us.kosli.com/api", false},
		{"https://app.kosli.com", "https://bucket.s3.amazonaws.com/app.kosli.com", false},
	} {
		t.Run(tt.a+" vs "+tt.b, func(t *testing.T) {
			a, err := url.Parse(tt.a)
			require.NoError(t, err)
			b, err := url.Parse(tt.b)
			require.NoError(t, err)
			require.Equal(t, tt.want, sameHost(a, b))
		})
	}
}
