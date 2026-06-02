package sonar_test

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/sonar"
)

// fakeSonar is a SonarQube stand-in for auth tests. It records the Authorization
// header of every request (in order) and gates responses by which scheme a given
// server version accepts: SonarQube Server < 10.0 accepts only Basic (Bearer -> 401),
// while >= 10.0 / Cloud accept Bearer.
type fakeSonar struct {
	mu            sync.Mutex
	authSeq       []string
	paths         []string
	acceptsBearer bool
	acceptsBasic  bool
	forceStatus   int      // if non-zero, every request returns this status + forceBody
	forceBody     string   // body served when forceStatus is set
	ceStatuses    []string // successive /api/ce/task Task.Status values (last repeats); default "SUCCESS"
}

func (f *fakeSonar) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.authSeq = append(f.authSeq, r.Header.Get("Authorization"))
		f.paths = append(f.paths, r.URL.Path)
		forceStatus, forceBody := f.forceStatus, f.forceBody
		f.mu.Unlock()

		// A forced status models a server-side condition (a 5xx, a structured 403, ...)
		// that is independent of the auth scheme, so it bypasses the scheme gate.
		if forceStatus != 0 {
			w.WriteHeader(forceStatus)
			_, _ = io.WriteString(w, forceBody)
			return
		}

		auth := r.Header.Get("Authorization")
		accepted := (strings.HasPrefix(auth, "Bearer ") && f.acceptsBearer) ||
			(strings.HasPrefix(auth, "Basic ") && f.acceptsBasic)
		if !accepted {
			// Pre-10.0 SonarQube rejects an unsupported scheme with a non-JSON 401:
			// the exact shape that surfaced the misleading error for ADCB.
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, "Unauthorized")
			return
		}
		switch r.URL.Path {
		case "/api/ce/task":
			_ = json.NewEncoder(w).Encode(sonar.TaskResponse{
				Task: sonar.Task{TaskID: "AYx", ComponentName: "differ", ComponentKey: "my-project", AnalysisID: "AN1", Status: f.nextCEStatus()},
			})
		case "/api/project_analyses/search":
			_ = json.NewEncoder(w).Encode(sonar.ProjectAnalyses{
				Analyses: []sonar.Analysis{{Key: "AN1", Date: "2026-01-01T00:00:00+0000", Revision: "abc123"}},
			})
		case "/api/qualitygates/project_status":
			_ = json.NewEncoder(w).Encode(sonar.QualityGateResponse{
				ProjectStatus: sonar.ProjectStatus{Status: "OK", Conditions: []sonar.Conditions{}},
			})
		default:
			http.NotFound(w, r)
		}
	}
}

// nextCEStatus returns the next scripted /api/ce/task status (the last value repeats),
// defaulting to SUCCESS so single-call tests need not configure it.
func (f *fakeSonar) nextCEStatus() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.ceStatuses) == 0 {
		return "SUCCESS"
	}
	s := f.ceStatuses[0]
	if len(f.ceStatuses) > 1 {
		f.ceStatuses = f.ceStatuses[1:]
	}
	return s
}

func (f *fakeSonar) authHeaders() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.authSeq)
}

func bearerAttempts(auths []string) int {
	n := 0
	for _, a := range auths {
		if strings.HasPrefix(a, "Bearer ") {
			n++
		}
	}
	return n
}

func wantBasicHeader(token string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(token+":"))
}

func discardLogger() *logger.Logger {
	return logger.NewLogger(io.Discard, io.Discard, false)
}

// TestGetSonarResults_Pre10Server_FallsBackToBasic proves the ADCB fix: against a
// SonarQube Server < 10.0 (which rejects Bearer with 401 and accepts only Basic),
// GetSonarResults must transparently retry with Basic and succeed, caching Basic for
// the rest of the run. Fails before the Bearer/Basic fallback exists.
func TestGetSonarResults_Pre10Server_FallsBackToBasic(t *testing.T) {
	fake := &fakeSonar{acceptsBearer: false, acceptsBasic: true} // pre-10.0: Basic only
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), srv.URL+"/api/ce/task?id=AYx", "", "", "", "", 5)
	res, err := sc.GetSonarResults(discardLogger())
	if err != nil {
		t.Fatalf("expected success against pre-10 server via Basic fallback, got error: %v", err)
	}
	if res.QualityGate == nil || res.QualityGate.Status != "OK" {
		t.Fatalf("expected quality gate status OK, got %+v", res.QualityGate)
	}

	auths := fake.authHeaders()
	if len(auths) == 0 || !strings.HasPrefix(auths[0], "Bearer ") {
		t.Fatalf("expected the first attempt to be Bearer, got %v", auths)
	}
	if c := bearerAttempts(auths); c != 1 {
		t.Errorf("expected exactly one Bearer attempt then cached Basic, got %d Bearer attempts in %v", c, auths)
	}
	if want := wantBasicHeader("tok"); !slices.Contains(auths, want) {
		t.Errorf("expected a Basic attempt %q, got %v", want, auths)
	}
}

func (f *fakeSonar) requestPaths() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.paths)
}

// TestGetSonarResults_BearerServer_NoBasicSent is the regression sentinel for
// SonarQube Cloud and Server >= 10.0: when Bearer is accepted, no Basic request is
// ever sent, so the fallback cannot regress those (the common) targets.
func TestGetSonarResults_BearerServer_NoBasicSent(t *testing.T) {
	fake := &fakeSonar{acceptsBearer: true, acceptsBasic: true}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), srv.URL+"/api/ce/task?id=AYx", "", "", "", "", 5)
	res, err := sc.GetSonarResults(discardLogger())
	if err != nil {
		t.Fatalf("expected success against a Bearer-capable server, got error: %v", err)
	}
	if res.QualityGate == nil || res.QualityGate.Status != "OK" {
		t.Fatalf("expected quality gate status OK, got %+v", res.QualityGate)
	}
	for _, a := range fake.authHeaders() {
		if strings.HasPrefix(a, "Basic ") {
			t.Errorf("Bearer was accepted, so no Basic request should be sent; got %v", fake.authHeaders())
			break
		}
	}
}

// TestGetSonarResults_InvalidToken_TriesBothThenErrors verifies that when neither
// scheme is accepted the CLI tries Bearer then Basic exactly once each on the first
// endpoint and then fails, without proceeding to later endpoints.
func TestGetSonarResults_InvalidToken_TriesBothThenErrors(t *testing.T) {
	fake := &fakeSonar{acceptsBearer: false, acceptsBasic: false}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), srv.URL+"/api/ce/task?id=AYx", "", "", "", "", 5)
	_, err := sc.GetSonarResults(discardLogger())
	if err == nil {
		t.Fatal("expected an error when neither auth scheme is accepted")
	}
	auths := fake.authHeaders()
	if len(auths) != 2 {
		t.Fatalf("expected exactly 2 attempts (Bearer then Basic), got %d: %v", len(auths), auths)
	}
	if !strings.HasPrefix(auths[0], "Bearer ") || !strings.HasPrefix(auths[1], "Basic ") {
		t.Errorf("expected attempt order [Bearer, Basic], got %v", auths)
	}
	for _, p := range fake.requestPaths() {
		if p != "/api/ce/task" {
			t.Errorf("expected only /api/ce/task to be attempted before failing, also hit %q", p)
		}
	}
	if msg := err.Error(); !strings.Contains(msg, "HTTP 401") || strings.Contains(msg, "please check your API token") {
		t.Errorf("expected a scheme-aware HTTP 401 error, not the old generic message; got: %v", err)
	}
}

// TestGetSonarResults_ServerError_NoFallback verifies a 5xx is surfaced as a server
// error and never triggers the Basic fallback (which would double-load a struggling
// server and mask the outage).
func TestGetSonarResults_ServerError_NoFallback(t *testing.T) {
	fake := &fakeSonar{acceptsBearer: true, acceptsBasic: true, forceStatus: 503, forceBody: "Service Unavailable"}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), srv.URL+"/api/ce/task?id=AYx", "", "", "", "", 5)
	_, err := sc.GetSonarResults(discardLogger())
	if err == nil {
		t.Fatal("expected an error on HTTP 503")
	}
	if !strings.Contains(err.Error(), "503") {
		t.Errorf("expected the error to mention HTTP 503, got: %v", err)
	}
	for _, a := range fake.authHeaders() {
		if strings.HasPrefix(a, "Basic ") {
			t.Errorf("must not fall back to Basic on a 5xx; got %v", fake.authHeaders())
			break
		}
	}
}

// TestGetSonarResults_StructuredForbidden_NoFallback verifies that a 403 carrying a
// structured SonarQube error is surfaced verbatim and does not trigger a Basic
// fallback: a decodable error means the scheme was understood, so it is a permissions
// problem, not an auth-scheme problem.
func TestGetSonarResults_StructuredForbidden_NoFallback(t *testing.T) {
	fake := &fakeSonar{acceptsBearer: true, acceptsBasic: true, forceStatus: 403, forceBody: `{"errors":[{"msg":"Insufficient privileges"}]}`}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), srv.URL+"/api/ce/task?id=AYx", "", "", "", "", 5)
	_, err := sc.GetSonarResults(discardLogger())
	if err == nil {
		t.Fatal("expected an error on a structured 403")
	}
	if !strings.Contains(err.Error(), "Insufficient privileges") {
		t.Errorf("expected the real SonarQube error to be surfaced, got: %v", err)
	}
	for _, a := range fake.authHeaders() {
		if strings.HasPrefix(a, "Basic ") {
			t.Errorf("must not fall back to Basic on a structured 403; got %v", fake.authHeaders())
			break
		}
	}
}

// TestGetSonarResults_TokenWhitespaceTrimmed verifies a token with a trailing newline
// (e.g. read from a secret file) is trimmed before use, so it still authenticates.
func TestGetSonarResults_TokenWhitespaceTrimmed(t *testing.T) {
	fake := &fakeSonar{acceptsBearer: false, acceptsBasic: true} // pre-10.0: Basic only
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok\n", t.TempDir(), srv.URL+"/api/ce/task?id=AYx", "", "", "", "", 5)
	res, err := sc.GetSonarResults(discardLogger())
	if err != nil {
		t.Fatalf("expected success with a trimmed token, got error: %v", err)
	}
	if res.QualityGate == nil || res.QualityGate.Status != "OK" {
		t.Fatalf("expected quality gate status OK, got %+v", res.QualityGate)
	}
	for _, a := range fake.authHeaders() {
		if strings.Contains(a, "\n") {
			t.Errorf("Authorization header must not contain the untrimmed newline; got %q", a)
		}
	}
	if want := wantBasicHeader("tok"); !slices.Contains(fake.authHeaders(), want) {
		t.Errorf("expected a Basic header for the trimmed token %q, got %v", want, fake.authHeaders())
	}
}

// TestGetSonarResults_Pre10PollLoop_ResolvesSchemeOncePerRun verifies that when the CE
// task is still processing, the --max-wait poll loop reuses the cached Basic scheme:
// the Bearer->Basic fallback happens exactly once for the whole run, not per poll.
func TestGetSonarResults_Pre10PollLoop_ResolvesSchemeOncePerRun(t *testing.T) {
	fake := &fakeSonar{acceptsBearer: false, acceptsBasic: true, ceStatuses: []string{"PENDING", "SUCCESS"}}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), srv.URL+"/api/ce/task?id=AYx", "", "", "", "", 3)
	res, err := sc.GetSonarResults(discardLogger())
	if err != nil {
		t.Fatalf("expected success after a PENDING poll, got error: %v", err)
	}
	if res.QualityGate == nil || res.QualityGate.Status != "OK" {
		t.Fatalf("expected quality gate status OK, got %+v", res.QualityGate)
	}
	if c := bearerAttempts(fake.authHeaders()); c != 1 {
		t.Errorf("expected exactly one Bearer attempt for the whole run (no re-probe per poll), got %d in %v", c, fake.authHeaders())
	}
}

// TestGetSonarResults_Forbidden_NonJSON_RendersActualStatus verifies the HTTP status
// in the error message is the real response code (here 403, not a hard-coded 401),
// and that a 403 does not trigger a Basic fallback.
func TestGetSonarResults_Forbidden_NonJSON_RendersActualStatus(t *testing.T) {
	fake := &fakeSonar{acceptsBearer: true, acceptsBasic: true, forceStatus: 403, forceBody: "Forbidden"}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), srv.URL+"/api/ce/task?id=AYx", "", "", "", "", 5)
	_, err := sc.GetSonarResults(discardLogger())
	if err == nil {
		t.Fatal("expected an error on a 403 with a non-JSON body")
	}
	if !strings.Contains(err.Error(), "HTTP 403") {
		t.Errorf("expected the message to render the actual status HTTP 403, got: %v", err)
	}
	if strings.Contains(err.Error(), "HTTP 401") {
		t.Errorf("status must not be hard-coded to 401; got: %v", err)
	}
	for _, a := range fake.authHeaders() {
		if strings.HasPrefix(a, "Basic ") {
			t.Errorf("a 403 must not trigger a Basic fallback; got %v", fake.authHeaders())
			break
		}
	}
}
