package sonar

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// authScheme selects how the SonarQube API token is presented to the server.
type authScheme int

const (
	// schemeAuto tries Bearer first and falls back to Basic on a 401 from a
	// self-hosted SonarQube Server < 10.0, which does not accept Bearer tokens.
	schemeAuto authScheme = iota
	schemeBearer
	schemeBasic
)

// isSonarCloudHost reports whether host belongs to SonarQube Cloud. Cloud always
// accepts Bearer and must never be sent a Basic request (organization/analysis
// tokens stopped accepting Basic auth in May 2026), so we never fall back for it.
func isSonarCloudHost(host string) bool {
	host = strings.ToLower(host)
	if i := strings.IndexByte(host, ':'); i >= 0 {
		host = host[:i]
	}
	return host == "sonarcloud.io" || strings.HasSuffix(host, ".sonarcloud.io")
}

func bearerHeaderValue(token string) string { return "Bearer " + token }

// basicHeaderValue presents the token as the HTTP Basic username with an empty
// password, which is how SonarQube Server < 10.0 accepts a token.
func basicHeaderValue(token string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(token+":"))
}

// authTransport authenticates SonarQube requests, resolving Bearer vs Basic once
// per run and caching the result. In schemeAuto against a self-hosted Server it
// tries Bearer and, on a 401, transparently retries the same request with Basic.
// It overrides any Authorization header set by the caller, so the request builders
// remain unchanged.
type authTransport struct {
	base  http.RoundTripper
	token string
	mode  authScheme

	mu       sync.Mutex
	decided  bool
	resolved authScheme
}

func (a *authTransport) baseRT() http.RoundTripper {
	if a.base != nil {
		return a.base
	}
	return http.DefaultTransport
}

// send clones req (a RoundTripper must not mutate its input), sets the scheme's
// Authorization header, and sends it.
func (a *authTransport) send(req *http.Request, scheme authScheme) (*http.Response, error) {
	r := req.Clone(req.Context())
	if scheme == schemeBasic {
		r.Header.Set("Authorization", basicHeaderValue(a.token))
	} else {
		r.Header.Set("Authorization", bearerHeaderValue(a.token))
	}
	return a.baseRT().RoundTrip(r)
}

func (a *authTransport) cache(scheme authScheme) {
	a.mu.Lock()
	if !a.decided {
		a.decided = true
		a.resolved = scheme
	}
	a.mu.Unlock()
}

func (a *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	a.mu.Lock()
	decided, resolved := a.decided, a.resolved
	a.mu.Unlock()

	// Scheme already resolved for this run: reuse it, no re-probing.
	if decided {
		return a.send(req, resolved)
	}

	switch {
	case a.mode == schemeBearer:
		return a.sendAndCache(req, schemeBearer)
	case a.mode == schemeBasic:
		return a.sendAndCache(req, schemeBasic)
	case isSonarCloudHost(req.URL.Host):
		// Cloud is always Bearer and must never receive a Basic request.
		return a.sendAndCache(req, schemeBearer)
	}

	// schemeAuto against a self-hosted Server: try Bearer, fall back to Basic on a
	// 401 (Server < 10.0 does not accept Bearer tokens). A 5xx or other status is
	// not a scheme rejection, so it is returned as-is without a fallback.
	resp, err := a.send(req, schemeBearer)
	if err != nil {
		return nil, err
	}
	// A redirect says nothing about the scheme: the client re-enters RoundTrip for
	// the next hop, and the response there decides.
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		return resp, nil
	}
	if resp.StatusCode == http.StatusUnauthorized {
		drainAndClose(resp)
		resp, err = a.send(req, schemeBasic)
		if err != nil {
			return nil, err
		}
		a.cache(schemeBasic)
		return resp, nil
	}
	a.cache(schemeBearer)
	return resp, nil
}

func (a *authTransport) sendAndCache(req *http.Request, scheme authScheme) (*http.Response, error) {
	resp, err := a.send(req, scheme)
	if err != nil {
		return nil, err
	}
	a.cache(scheme)
	return resp, nil
}

func drainAndClose(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

const (
	// maxSonarRedirects bounds how many redirects the Sonar client follows. Go's
	// default of ten is more than any real SonarQube deployment needs.
	maxSonarRedirects = 5

	// sonarClientTimeout bounds one request end to end, redirects and the Basic
	// retry included, so a host that accepts the connection and never answers
	// cannot hold the run. SonarQube API responses are small JSON documents, so
	// this is generous for a healthy server.
	sonarClientTimeout = 60 * time.Second
)

// sonarRedirectPolicy keeps the API token on the host the user configured.
// authTransport attaches the token to every request it sends, including the
// requests Go issues while following redirects, so the stdlib's own header
// stripping never applies. Instead of following a redirect without the token
// (every SonarQube endpoint needs it, so that would only fail later and less
// clearly), a redirect to another host or from https to http is refused outright.
func sonarRedirectPolicy(req *http.Request, via []*http.Request) error {
	// via holds the requests already sent, so it is one longer than the redirects followed.
	if len(via) > maxSonarRedirects {
		return fmt.Errorf("stopped after %d redirects", maxSonarRedirects)
	}
	prev := via[len(via)-1].URL
	if canonicalHost(req.URL) != canonicalHost(prev) {
		return fmt.Errorf("cross-host redirect from %s to %s refused: the SonarQube API token is only sent to the configured host.\n"+
			"This usually means SonarQube redirected an unauthenticated request to a login page, or the configured server URL is not the instance's canonical URL. "+
			"If %s is the SonarQube API, point --sonar-server-url (or the scanner's sonar.host.url, which report-task.txt records) at it directly",
			prev.Host, req.URL.Host, req.URL.Host)
	}
	if prev.Scheme == "https" && req.URL.Scheme != "https" {
		return fmt.Errorf("redirect from https to http on %s refused: the SonarQube API token would be sent in plain text", prev.Host)
	}
	return nil
}

// canonicalHost lowercases the hostname and drops the scheme's default port, so a
// reverse proxy that emits an explicit :443 in Location still counts as the same host.
func canonicalHost(u *url.URL) string {
	host := strings.ToLower(u.Hostname())
	port := u.Port()
	if port == "" || (u.Scheme == "https" && port == "443") || (u.Scheme == "http" && port == "80") {
		return host
	}
	return host + ":" + port
}

// newAuthedClient builds an HTTP client that authenticates SonarQube requests with
// the given token, presenting it as Bearer (SonarQube Cloud and Server >= 10.0) and
// falling back to Basic for a self-hosted Server < 10.0. The token is trimmed of
// surrounding whitespace (e.g. a trailing newline from a secret file). Redirects
// are followed only within the configured host, see sonarRedirectPolicy.
func newAuthedClient(token string, mode authScheme) *http.Client {
	return &http.Client{
		Transport:     &authTransport{token: strings.TrimSpace(token), mode: mode},
		CheckRedirect: sonarRedirectPolicy,
		Timeout:       sonarClientTimeout,
	}
}

// sonarResponseError turns a SonarQube response that could not be parsed as the
// expected JSON into an actionable error, based on the HTTP status. It replaces the
// previous catch-all "please check your API token" message, which misdiagnosed
// everything (including the self-hosted Server < 10.0 auth-scheme case) as a bad token.
func sonarResponseError(statusCode int) error {
	switch {
	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden:
		return fmt.Errorf("SonarQube rejected the request (HTTP %d): the API token is invalid or does not have permission to read this project.\nThe Kosli CLI uses Bearer authentication for SonarQube Cloud and Server 10.0 and later, and Basic authentication for older self-hosted Servers", statusCode)
	case statusCode >= 500:
		return fmt.Errorf("SonarQube returned HTTP %d: the server may be unavailable or experiencing problems; please try again later", statusCode)
	default:
		return fmt.Errorf("SonarQube returned an unexpected response (HTTP %d) that could not be parsed; check that the SonarQube server URL is correct and that the API token has the required permissions", statusCode)
	}
}
