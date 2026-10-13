package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

// fakeRegoPolicyRead records each read and answers with body and status.
type fakeRegoPolicyRead struct {
	body        string
	status      int
	escapedPath string
	query       url.Values
	calls       int
}

func newFakeRegoPolicyRead(t *testing.T, body string) (*httptest.Server, *fakeRegoPolicyRead) {
	t.Helper()
	fake := &fakeRegoPolicyRead{body: body, status: http.StatusOK}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fake.calls++
		fake.escapedPath = r.URL.EscapedPath()
		fake.query = r.URL.Query()
		w.WriteHeader(fake.status)
		_, _ = fmt.Fprint(w, fake.body)
	}))
	t.Cleanup(server.Close)
	return server, fake
}

// withoutOrgAndApiToken leaves no org or API token in the environment or a config file.
func withoutOrgAndApiToken(t *testing.T) string {
	t.Helper()
	for _, key := range []string{"KOSLI_ORG", "KOSLI_API_TOKEN"} {
		t.Setenv(key, "")
		_ = os.Unsetenv(key)
	}
	return "--config-file " + filepath.Join(t.TempDir(), "absent.yml")
}
