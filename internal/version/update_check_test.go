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
	// (override URL via a package-level var for testability — see note below)

	notice, err := checkForUpdateWithURL("v0.1.0", srv.URL)
	assert.NoError(t, err)
	assert.Contains(t, notice, "v9.99.0")
}

func TestCheckForUpdate_AlreadyLatest(t *testing.T) {
	current := GetVersion() // always reflects the real built version
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, `{"tag_name":"%s"}`, current)
	}))
	defer srv.Close()

	notice, err := checkForUpdateWithURL(current, srv.URL)
	assert.NoError(t, err)
	assert.Empty(t, notice)
}

func TestCheckForUpdate_OptOut(t *testing.T) {
	t.Setenv("KOSLI_NO_UPDATE_CHECK", "1")
	notice, _ := CheckForUpdate("v0.1.0")
	assert.Empty(t, notice)
}

func TestCheckForUpdate_NetworkError(t *testing.T) {
	notice, err := checkForUpdateWithURL("v0.1.0", "http://localhost:1") // nothing listening
	assert.NoError(t, err)                                               // must be silent
	assert.Empty(t, notice)
}
