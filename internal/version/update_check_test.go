package version

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckForUpdate_NewVersionAvailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"tag_name":"v9.99.0"}`)
	}))
	defer srv.Close()

	notice, err := checkForUpdateWithURL("v0.1.0", srv.URL)
	assert.NoError(t, err)
	assert.Contains(t, notice, "v9.99.0")
}

func TestCheckForUpdate_AlreadyLatest(t *testing.T) {
	current := "v9.99.0"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"tag_name":"%s"}`, current)
	}))
	defer srv.Close()

	notice, err := checkForUpdateWithURL(current, srv.URL)
	assert.NoError(t, err)
	assert.Empty(t, notice)
}

func TestCheckForUpdate_NewerThanLatest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"tag_name":"v1.0.0"}`)
	}))
	defer srv.Close()

	// User has a newer version — should NOT show an update notice
	notice, err := checkForUpdateWithURL("v2.0.0", srv.URL)
	assert.NoError(t, err)
	assert.Empty(t, notice)
}

func TestCheckForUpdate_OptOut(t *testing.T) {
	t.Setenv("KOSLI_NO_UPDATE_CHECK", "1")
	notice, _ := CheckForUpdate("v0.1.0")
	assert.Empty(t, notice)
}

func TestCheckForUpdate_DevBuild(t *testing.T) {
	// dev builds should be skipped without any HTTP call
	notice, err := checkForUpdateWithURL("main", "http://should-not-be-called")
	assert.NoError(t, err)
	assert.Empty(t, notice)
}

func TestCheckForUpdate_NetworkError(t *testing.T) {
	notice, err := checkForUpdateWithURL("v0.1.0", "http://localhost:1") // nothing listening
	assert.NoError(t, err)                                               // must be silent
	assert.Empty(t, notice)
}

func TestCheckForUpdate_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `not json`)
	}))
	defer srv.Close()

	notice, err := checkForUpdateWithURL("v0.1.0", srv.URL)
	assert.NoError(t, err)
	assert.Empty(t, notice)
}

func TestCheckForUpdate_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	notice, err := checkForUpdateWithURL("v0.1.0", srv.URL)
	assert.NoError(t, err)
	assert.Empty(t, notice)
}
