package github

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"
)

type GithubTestSuite struct {
	suite.Suite
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *GithubTestSuite) TestNewGithubClientFromToken() {
	for _, t := range []struct {
		name    string
		token   string
		baseURL string
	}{
		{
			name:  "when provided a token, a client is created.",
			token: "some_fake_token",
		},
		{
			name:    "when baseURL and token are provided, a client is created.",
			token:   "some_fake_token",
			baseURL: "https://github.example.com",
		},
	} {
		suite.Run(t.name, func() {
			client, err := NewGithubClientFromToken(context.Background(), t.token, t.baseURL, false)
			require.NoErrorf(suite.T(), err, "was NOT expecting error but got: %s", err)
			require.NotNilf(suite.T(), client, "client should not be nil")
		})
	}
}

func (suite *GithubTestSuite) TestExtractRepoName() {
	for _, t := range []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "full repo name (including org) is separated",
			input: "kosli-dev/cli",
			want:  "cli",
		},
		{
			name:  "short repo name is returned as is",
			input: "cli",
			want:  "cli",
		},
	} {
		suite.Run(t.name, func() {
			repo := extractRepoName(t.input)
			require.Equalf(suite.T(), t.want, repo, "expected %s but got %s", t.want, repo)
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGithubTestSuite(t *testing.T) {
	suite.Run(t, new(GithubTestSuite))
}

// graphqlNullPRResponse is a minimal valid GraphQL response where the PR is null.
// PREvidenceByPRNumber returns nil, nil in this case.
const graphqlNullPRResponse = `{"data":{"repository":{"pullRequest":null}}}`

func newRetryTestServer(t *testing.T, failCount int) (*httptest.Server, *int) {
	t.Helper()
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/graphql" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		attempts++
		if attempts <= failCount {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, graphqlNullPRResponse)
	}))
	return ts, &attempts
}

func newRetryConfig(serverURL string, sleepFn func(time.Duration)) *GithubConfig {
	return &GithubConfig{
		Token:      "fake-token",
		BaseURL:    serverURL,
		Org:        "test-org",
		Repository: "test-repo",
		Sleep:      sleepFn,
	}
}

func TestPREvidenceByPRNumber_RetriesOnGraphQLError(t *testing.T) {
	ts, attempts := newRetryTestServer(t, 2)
	defer ts.Close()

	var sleptDurations []time.Duration
	config := newRetryConfig(ts.URL, func(d time.Duration) { sleptDurations = append(sleptDurations, d) })

	pr, err := config.PREvidenceByPRNumber(1)
	require.NoError(t, err)
	require.Nil(t, pr)
	require.Equal(t, 3, *attempts, "should have retried twice before succeeding")
	require.Len(t, sleptDurations, 2, "should have slept between retries")
}

func TestPREvidenceByPRNumber_ReturnsErrorAfterAllRetriesExhausted(t *testing.T) {
	ts, attempts := newRetryTestServer(t, 999)
	defer ts.Close()

	config := newRetryConfig(ts.URL, func(time.Duration) {})

	_, err := config.PREvidenceByPRNumber(1)
	require.Error(t, err)
	require.Equal(t, 4, *attempts, "should have made 1 initial attempt + 3 retries")
}

func TestRedactAuthHeader(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{
			name: "scheme + long token keeps last 4",
			in:   "Bearer ghp_abcdef1234567890ABCD",
			want: "Bearer ***ABCD",
		},
		{
			name: "scheme + short token is fully redacted",
			in:   "token ab",
			want: "token ***",
		},
		{
			name: "scheme + 4-char token is fully redacted",
			in:   "Bearer abcd",
			want: "Bearer ***",
		},
		{
			name: "no scheme, long value keeps last 4",
			in:   "nospaceshort",
			want: "***hort",
		},
		{
			name: "no scheme, short value is fully redacted",
			in:   "abc",
			want: "***",
		},
		{
			name: "empty value is fully redacted",
			in:   "",
			want: "***",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, redactAuthHeader(tc.in))
		})
	}
}

// fakeRoundTripper lets us inject an arbitrary response/error pair into
// debugTransport without needing a real network. Used to assert behaviour
// on transport-level errors, including the case where some middleboxes
// return both a non-nil response and an error.
type fakeRoundTripper struct {
	resp *http.Response
	err  error
}

func (f *fakeRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return f.resp, f.err
}

// TestNewGithubClientFromTokenDebugChain pins the transport chain shape
// when debug=true: oauth2.Transport must be at the top and debugTransport
// must be its Base. If this order is reversed, the Authorization header
// added by oauth2 won't appear in the debug dump — which is the whole
// reason customers turn debug on for auth failures.
func TestNewGithubClientFromTokenDebugChain(t *testing.T) {
	client, err := NewGithubClientFromToken(context.Background(), "fake-token", "", true)
	require.NoError(t, err)
	require.NotNil(t, client)

	httpClient := client.Client()
	require.NotNil(t, httpClient.Transport)

	oauth2Tr, ok := httpClient.Transport.(*oauth2.Transport)
	require.True(t, ok, "expected *oauth2.Transport at top of chain so it adds Authorization before debug logs")

	_, ok = oauth2Tr.Base.(*debugTransport)
	require.True(t, ok, "expected *debugTransport as oauth2.Transport.Base when debug=true")
}

// TestDebugTransport_LogsResponseOnTransportError covers the case where a
// middlebox (corporate proxy, HTTP/2 layer) returns an error AND a non-nil
// response. Before this fix the response — which carries the actual
// diagnostic body — was silently dropped.
func TestDebugTransport_LogsResponseOnTransportError(t *testing.T) {
	var buf bytes.Buffer
	resp := &http.Response{
		StatusCode: 407,
		Status:     "407 Proxy Authentication Required",
		Header:     http.Header{"Proxy-Authenticate": []string{"Basic realm=\"corp\""}},
		Body:       io.NopCloser(strings.NewReader("authentication required by proxy")),
	}
	tr := &debugTransport{
		base: &fakeRoundTripper{resp: resp, err: errors.New("authenticationrequired")},
		out:  &buf,
	}
	req, err := http.NewRequest("POST", "https://api.github.com/graphql", strings.NewReader(`{"query":"x"}`))
	require.NoError(t, err)

	gotResp, gotErr := tr.RoundTrip(req)
	require.Error(t, gotErr)
	require.NotNil(t, gotResp)

	out := buf.String()
	require.Contains(t, out, "transport error: authenticationrequired")
	require.Contains(t, out, "407 Proxy Authentication Required")
	require.Contains(t, out, "authentication required by proxy")
}

// TestDebugTransport_LogsProxyWhenSet verifies that when a proxy is
// configured for the request, the proxy URL is logged on the request
// line. This is a strong diagnostic for the customer's case where curl
// works but the CLI gets a transport-level rejection.
func TestDebugTransport_LogsProxyWhenSet(t *testing.T) {
	proxyURL, err := url.Parse("http://corp-proxy.example.com:3128")
	require.NoError(t, err)

	var buf bytes.Buffer
	tr := &debugTransport{
		base: &fakeRoundTripper{resp: &http.Response{
			StatusCode: 200,
			Status:     "200 OK",
			Header:     http.Header{},
			Body:       io.NopCloser(strings.NewReader("{}")),
		}},
		out:       &buf,
		proxyFunc: func(*http.Request) (*url.URL, error) { return proxyURL, nil },
	}
	req, err := http.NewRequest("POST", "https://api.github.com/graphql", nil)
	require.NoError(t, err)

	_, err = tr.RoundTrip(req)
	require.NoError(t, err)

	require.Contains(t, buf.String(), "<via proxy http://corp-proxy.example.com:3128>")
}

// TestDebugTransport_LogsProxyLookupError verifies that when proxyFunc
// itself fails (e.g. malformed HTTP_PROXY env var) we surface the error
// in the dump rather than silently swallowing it. This is the debug
// transport — logging more is its job.
func TestDebugTransport_LogsProxyLookupError(t *testing.T) {
	var buf bytes.Buffer
	tr := &debugTransport{
		base: &fakeRoundTripper{resp: &http.Response{
			StatusCode: 200,
			Status:     "200 OK",
			Header:     http.Header{},
			Body:       io.NopCloser(strings.NewReader("{}")),
		}},
		out: &buf,
		proxyFunc: func(*http.Request) (*url.URL, error) {
			return nil, errors.New("invalid proxy URL: missing scheme")
		},
	}
	req, err := http.NewRequest("POST", "https://api.github.com/graphql", nil)
	require.NoError(t, err)

	_, err = tr.RoundTrip(req)
	require.NoError(t, err)

	require.Contains(t, buf.String(), "<proxy lookup error: invalid proxy URL: missing scheme>")
}

// TestDebugTransport_RedactsProxyUserinfo ensures credentials embedded in
// the proxy URL (e.g. http://user:pass@proxy) are not leaked into debug
// output. url.URL.Redacted() replaces the password with "xxxxx".
func TestDebugTransport_RedactsProxyUserinfo(t *testing.T) {
	proxyURL, err := url.Parse("http://user:secret@corp-proxy.example.com:3128")
	require.NoError(t, err)

	var buf bytes.Buffer
	tr := &debugTransport{
		base: &fakeRoundTripper{resp: &http.Response{
			StatusCode: 200,
			Status:     "200 OK",
			Header:     http.Header{},
			Body:       io.NopCloser(strings.NewReader("{}")),
		}},
		out:       &buf,
		proxyFunc: func(*http.Request) (*url.URL, error) { return proxyURL, nil },
	}
	req, err := http.NewRequest("POST", "https://api.github.com/graphql", nil)
	require.NoError(t, err)

	_, err = tr.RoundTrip(req)
	require.NoError(t, err)

	out := buf.String()
	require.NotContains(t, out, "secret")
	require.Contains(t, out, "user:xxxxx@corp-proxy.example.com")
}

func TestRedactSensitiveHeader(t *testing.T) {
	for _, tc := range []struct {
		name       string
		headerName string
		value      string
		want       string
	}{
		{
			name:       "authorization is redacted via redactAuthHeader",
			headerName: "Authorization",
			value:      "Bearer ghp_abcdef1234567890ABCD",
			want:       "Bearer ***ABCD",
		},
		{
			name:       "authorization match is case-insensitive",
			headerName: "AUTHORIZATION",
			value:      "Bearer ghp_abcdef1234567890ABCD",
			want:       "Bearer ***ABCD",
		},
		{
			name:       "cookie is fully redacted",
			headerName: "Cookie",
			value:      "session=abc; csrf=xyz",
			want:       "***",
		},
		{
			name:       "set-cookie is fully redacted",
			headerName: "Set-Cookie",
			value:      "session=abc; Path=/; HttpOnly",
			want:       "***",
		},
		{
			name:       "proxy-authorization is fully redacted",
			headerName: "Proxy-Authorization",
			value:      "Basic dXNlcjpwYXNz",
			want:       "***",
		},
		{
			name:       "non-sensitive header is returned unchanged",
			headerName: "Content-Type",
			value:      "application/json",
			want:       "application/json",
		},
		{
			name:       "x-oauth-scopes is intentionally not redacted",
			headerName: "X-OAuth-Scopes",
			value:      "repo, read:org",
			want:       "repo, read:org",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, redactSensitiveHeader(tc.headerName, tc.value))
		})
	}
}
