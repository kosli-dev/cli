package sonar

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestIsSonarCloudHost(t *testing.T) {
	cases := []struct {
		host string
		want bool
	}{
		{"sonarcloud.io", true},
		{"SonarCloud.io", true},
		{"sonarcloud.io:443", true},
		{"api.sonarcloud.io", true},
		{"www.sonarcloud.io", true},
		{"sonarqube.mycorp.local", false},
		{"adcm5929.adcbmis.local:9000", false},
		{"localhost:9000", false},
		{"notsonarcloud.io", false},       // must not match by accident
		{"sonarcloud.io.evil.com", false}, // suffix-spoof must be rejected
	}
	for _, c := range cases {
		if got := isSonarCloudHost(c.host); got != c.want {
			t.Errorf("isSonarCloudHost(%q) = %v, want %v", c.host, got, c.want)
		}
	}
}

// TestSonarRedirectPolicy pins which redirects the Sonar client follows. The
// client-level tests below use httptest servers, which all live on 127.0.0.1 and
// differ only by port, so hostname, case and default-port handling are checked here.
func TestSonarRedirectPolicy(t *testing.T) {
	cases := []struct {
		name    string
		prev    string
		next    string
		wantErr string // "" means the redirect is followed
	}{
		{"different hostname", "https://sonar.example.com/x", "https://evil.example.com/x", "cross-host redirect"},
		{"subdomain of the configured host", "https://sonar.example.com/x", "https://api.sonar.example.com/x", "cross-host redirect"},
		{"different port", "https://sonar.example.com/x", "https://sonar.example.com:8443/x", "cross-host redirect"},
		{"https to http downgrade", "https://sonar.example.com/x", "http://sonar.example.com/x", "https to http"},
		{"hostname differs only by case", "https://sonar.example.com/x", "https://SONAR.Example.com/y", ""},
		{"explicit default https port added", "https://sonar.example.com/x", "https://sonar.example.com:443/x", ""},
		{"explicit default http port dropped", "http://sonar.example.com:80/x", "http://sonar.example.com/x", ""},
		{"http to https upgrade", "http://sonar.example.com/x", "https://sonar.example.com/x", ""},
		{"path rewrite", "https://sonar.example.com/old", "https://sonar.example.com/new", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := sonarRedirectPolicy(mustRequest(t, c.next), []*http.Request{mustRequest(t, c.prev)})
			switch {
			case c.wantErr == "" && err != nil:
				t.Errorf("expected the redirect to be followed, got: %v", err)
			case c.wantErr != "" && err == nil:
				t.Errorf("expected the redirect to be refused with %q, got nil", c.wantErr)
			case c.wantErr != "" && !strings.Contains(err.Error(), c.wantErr):
				t.Errorf("expected error containing %q, got: %v", c.wantErr, err)
			}
		})
	}
}

// TestSonarRedirectPolicy_LimitMessageMatchesRedirectsFollowed: via holds the
// requests already sent, so the message must report the limit, not len(via).
func TestSonarRedirectPolicy_LimitMessageMatchesRedirectsFollowed(t *testing.T) {
	via := make([]*http.Request, 0, maxSonarRedirects+1)
	for range maxSonarRedirects + 1 {
		via = append(via, mustRequest(t, "https://sonar.example.com/loop"))
	}
	err := sonarRedirectPolicy(mustRequest(t, "https://sonar.example.com/loop"), via)
	if err == nil || !strings.Contains(err.Error(), "stopped after 5 redirects") {
		t.Errorf("expected the limit error to name %d redirects, got: %v", maxSonarRedirects, err)
	}
	if err := sonarRedirectPolicy(mustRequest(t, "https://sonar.example.com/loop"), via[:maxSonarRedirects]); err != nil {
		t.Errorf("expected the %dth redirect to still be followed, got: %v", maxSonarRedirects, err)
	}
}

// authRecorder is a redirect target that records the Authorization header of
// every request it receives.
type authRecorder struct {
	hits atomic.Int32
	auth atomic.Value
}

func (r *authRecorder) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		r.auth.Store(req.Header.Get("Authorization"))
		r.hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}
}

func newTestServer(t *testing.T, h http.Handler) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

// TestAuthedClient_CrossHostRedirect_DoesNotLeakToken is the sentinel for
// server#6880: a Sonar host that redirects to another host must not be followed,
// because the transport would otherwise attach the API token to the new host.
func TestAuthedClient_CrossHostRedirect_DoesNotLeakToken(t *testing.T) {
	target := &authRecorder{}
	targetSrv := newTestServer(t, target.handler())
	sonarSrv := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, targetSrv.URL+"/api/ce/task", http.StatusFound)
	}))

	client := newAuthedClient("SUPER-SECRET-SONAR-TOKEN", schemeBearer)
	resp, err := client.Get(sonarSrv.URL + "/api/ce/task")
	if err == nil {
		drainAndClose(resp)
	}
	if n := target.hits.Load(); n != 0 {
		t.Fatalf("redirect target must never be contacted, got %d request(s) with Authorization %q", n, target.auth.Load())
	}
	if err == nil {
		t.Fatal("expected the cross-host redirect to be refused, got a response")
	}
	if !strings.Contains(err.Error(), "cross-host redirect") {
		t.Errorf("expected a cross-host redirect error, got: %v", err)
	}
}

// TestAuthedClient_SameHostRedirect_IsFollowedWithToken keeps ordinary Sonar
// deployments working: a same-host redirect (a path rewrite, a trailing slash) is
// followed and the redirected request is still authenticated.
func TestAuthedClient_SameHostRedirect_IsFollowedWithToken(t *testing.T) {
	target := &authRecorder{}
	mux := http.NewServeMux()
	mux.Handle("/new", target.handler())
	mux.HandleFunc("/old", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/new", http.StatusFound)
	})
	srv := newTestServer(t, mux)

	client := newAuthedClient("tok", schemeBearer)
	resp, err := client.Get(srv.URL + "/old")
	if err != nil {
		t.Fatalf("expected the same-host redirect to be followed, got: %v", err)
	}
	drainAndClose(resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from the redirect target, got %d", resp.StatusCode)
	}
	if got := target.auth.Load(); got != bearerHeaderValue("tok") {
		t.Errorf("expected the redirected request to carry the Bearer token, got %q", got)
	}
}

// TestAuthedClient_RedirectThenUnauthorized_StillFallsBackToBasic: a 3xx on the
// Bearer probe proves nothing about the scheme. On a Server < 10.0 that redirects
// /api/ce/task to /api/ce/task/, the 401 on the redirected hop must still trigger
// the Basic fallback instead of being returned as an invalid token.
func TestAuthedClient_RedirectThenUnauthorized_StillFallsBackToBasic(t *testing.T) {
	target := &authRecorder{}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/ce/task", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/ce/task/", http.StatusFound)
	})
	mux.HandleFunc("/api/ce/task/", func(w http.ResponseWriter, r *http.Request) {
		// Server < 10.0 accepts only Basic.
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Basic ") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		target.handler()(w, r)
	})
	srv := newTestServer(t, mux)

	client := newAuthedClient("tok", schemeAuto)
	resp, err := client.Get(srv.URL + "/api/ce/task")
	if err != nil {
		t.Fatalf("expected the redirected request to succeed via Basic, got: %v", err)
	}
	drainAndClose(resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 after the Basic fallback on the redirected hop, got %d", resp.StatusCode)
	}
	if got := target.auth.Load(); got != basicHeaderValue("tok") {
		t.Errorf("expected the redirected request to be retried with Basic, got %q", got)
	}
}

// TestAuthedClient_RedirectLoop_StopsAtLimit: a Sonar host that redirects to itself
// forever must fail fast rather than spin through Go's default of ten hops.
func TestAuthedClient_RedirectLoop_StopsAtLimit(t *testing.T) {
	var hits atomic.Int32
	srv := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		http.Redirect(w, r, "/loop", http.StatusFound)
	}))

	client := newAuthedClient("tok", schemeBearer)
	resp, err := client.Get(srv.URL + "/loop")
	if err == nil {
		drainAndClose(resp)
		t.Fatal("expected the redirect loop to be refused, got a response")
	}
	if !strings.Contains(err.Error(), "redirects") {
		t.Errorf("expected a redirect-limit error, got: %v", err)
	}
	if n := hits.Load(); n != maxSonarRedirects+1 {
		t.Errorf("expected the initial request plus %d followed redirects, got %d requests", maxSonarRedirects, n)
	}
}

// TestAuthedClient_HangingRedirectTarget_TimesOut: the hop cap bounds how many
// redirects are followed, the client deadline bounds how long the whole chain may
// take, so a host that accepts the connection and never answers cannot hold the
// run forever.
func TestAuthedClient_HangingRedirectTarget_TimesOut(t *testing.T) {
	release := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/old", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/hang", http.StatusFound)
	})
	mux.HandleFunc("/hang", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	})
	srv := newTestServer(t, mux)
	// Cleanup runs LIFO and srv.Close blocks until handlers return, so the parked
	// handler must be released before the server is closed.
	t.Cleanup(func() { close(release) })

	client := newAuthedClient("tok", schemeBearer)
	if client.Timeout != sonarClientTimeout {
		t.Fatalf("expected the client deadline to be %v, got %v", sonarClientTimeout, client.Timeout)
	}
	client.Timeout = 200 * time.Millisecond // keep the test fast; the wiring is asserted above

	start := time.Now()
	resp, err := client.Get(srv.URL + "/old")
	if err == nil {
		drainAndClose(resp)
		t.Fatal("expected the hanging redirect target to time out, got a response")
	}
	var urlErr *url.Error
	if !errors.As(err, &urlErr) || !urlErr.Timeout() {
		t.Errorf("expected a timeout error, got: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("expected the deadline to cut the request short, waited %v", elapsed)
	}
}

func mustRequest(t *testing.T, rawURL string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	return req
}
