package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	kosliErrors "github.com/kosli-dev/cli/internal/errors"
	"github.com/stretchr/testify/assert"
)

// TestExitCodeScenarios verifies that each documented exit code is produced
// by the right class of error. Uses executeCommandC so no binary build is needed.
//
// Codes covered:
//
//	0 — success
//	1 — compliance/policy violation
//	2 — server unreachable or 5xx
//	3 — invalid API token (401/403)
//	4 — CLI usage error (missing flags, unknown flag, wrong arg count)
func TestExitCodeScenarios(t *testing.T) {
	assertCode := func(t *testing.T, cmd string, want int) {
		t.Helper()
		_, _, err := executeCommandC(cmd)
		got := kosliErrors.ExitCodeFor(err)
		assert.Equal(t, want, got, "command: %s\nerror: %v", cmd, err)
	}

	// ── exit 0: success ───────────────────────────────────────────────────────

	t.Run("exit 0: kosli version succeeds", func(t *testing.T) {
		assertCode(t, "version", kosliErrors.ExitOK)
	})

	// ── exit 1: compliance / policy violation ─────────────────────────────────

	t.Run("exit 1: evaluate trail with deny-all policy", func(t *testing.T) {
		srv := newStaticJSONServer(t, http.StatusOK, `{"name":"my-trail","compliance_status":{}}`)
		cmd := fmt.Sprintf(
			"evaluate trail my-trail --flow my-flow --policy testdata/policies/deny-all.rego --host %s --org test-org --api-token secret",
			srv.URL,
		)
		assertCode(t, cmd, kosliErrors.ExitCompliance)
	})

	// ── exit 2: server unreachable / 5xx ──────────────────────────────────────

	t.Run("exit 2: server returns 500", func(t *testing.T) {
		srv := newStaticJSONServer(t, http.StatusInternalServerError, `{"message":"internal error"}`)
		cmd := fmt.Sprintf(
			"list environments --host %s --org test-org --api-token secret --max-api-retries 0",
			srv.URL,
		)
		assertCode(t, cmd, kosliErrors.ExitServer)
	})

	t.Run("exit 2: server unreachable (bad host)", func(t *testing.T) {
		assertCode(t,
			"list environments --host http://localhost:19999 --org test-org --api-token secret --max-api-retries 0",
			kosliErrors.ExitServer,
		)
	})

	// ── exit 3: invalid API token ─────────────────────────────────────────────

	t.Run("exit 3: server returns 401", func(t *testing.T) {
		srv := newStaticJSONServer(t, http.StatusUnauthorized, `{"message":"unauthorized"}`)
		cmd := fmt.Sprintf(
			"list environments --host %s --org test-org --api-token bad-token",
			srv.URL,
		)
		assertCode(t, cmd, kosliErrors.ExitConfig)
	})

	t.Run("exit 3: server returns 403", func(t *testing.T) {
		srv := newStaticJSONServer(t, http.StatusForbidden, `{"message":"forbidden"}`)
		cmd := fmt.Sprintf(
			"list environments --host %s --org test-org --api-token bad-token",
			srv.URL,
		)
		assertCode(t, cmd, kosliErrors.ExitConfig)
	})

	// ── exit 4: CLI usage errors ──────────────────────────────────────────────

	t.Run("exit 4: unknown flag", func(t *testing.T) {
		assertCode(t, "version --no-such-flag", kosliErrors.ExitUsage)
	})

	t.Run("exit 4: cobra required flag not set (--flow, --policy)", func(t *testing.T) {
		assertCode(t,
			"evaluate trail my-trail --host http://localhost:8001 --org test-org --api-token secret",
			kosliErrors.ExitUsage,
		)
	})

	t.Run("exit 4: wrong number of arguments", func(t *testing.T) {
		// evaluate trail requires exactly 1 positional argument
		assertCode(t,
			"evaluate trail --host http://localhost:8001 --org test-org --api-token secret",
			kosliErrors.ExitUsage,
		)
	})
}

// newStaticJSONServer starts a test HTTP server that always responds with the
// given status code and JSON body, and closes itself when the test ends.
func newStaticJSONServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = fmt.Fprint(w, body)
	}))
	t.Cleanup(srv.Close)
	return srv
}
