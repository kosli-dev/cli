package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeProductUpdates records each request and answers with body and status.
// When listBody is set, a GET is answered with it and a 200 instead.
type fakeProductUpdates struct {
	body        string
	listBody    string
	failOnPost  int // 1-based; that POST is answered with a 403
	posted      []string
	status      int
	method      string
	escapedPath string
	requestBody string
	calls       int
}

func newFakeProductUpdates(t *testing.T, status int, body string) (*httptest.Server, *fakeProductUpdates) {
	t.Helper()
	fake := &fakeProductUpdates{body: body, status: status}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fake.calls++
		fake.method = r.Method
		fake.escapedPath = r.URL.EscapedPath()
		requestBody, _ := io.ReadAll(r.Body)
		fake.requestBody = string(requestBody)
		if r.Method == http.MethodPost {
			fake.posted = append(fake.posted, fake.requestBody)
		}
		if r.Method == http.MethodPost && len(fake.posted) == fake.failOnPost {
			w.WriteHeader(http.StatusForbidden)
			_, _ = fmt.Fprint(w, `{"message":"boom"}`)
			return
		}
		if r.Method == http.MethodGet && fake.listBody != "" {
			_, _ = fmt.Fprint(w, fake.listBody)
			return
		}
		w.WriteHeader(fake.status)
		_, _ = fmt.Fprint(w, fake.body)
	}))
	t.Cleanup(server.Close)
	return server, fake
}
