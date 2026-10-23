package sonar_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/kosli-dev/cli/internal/sonar"
)

// The responses below are trimmed copies of what sonarcloud.io returned on 23 Oct 2026
// for the public project cyber-dojo_differ, whose branch run-workflow-196 is short-lived.
// The branch name here has a slash, as the customer's feature branches do.
const (
	shortBranch       = "feature/kosli-integration"
	shortLatestCommit = "6860082827803e7c2dabef20d9941de77d4eda23"
	shortAnalysisDate = "2026-10-20T07:06:05+0000"
	longBranch        = "release/uat"
	unlistedBranch    = "no-such-branch"
	// badGateway in branchCommits makes that branch listing fail.
	badGateway = "502"
)

type fakeSonarCloudBranches struct {
	mu            sync.Mutex
	paths         []string
	qualityGateQs []url.Values
	// branchCommits is the short branch's latest commit at each branch listing in
	// turn, to simulate a newer push being processed between two calls.
	branchCommits []string
}

func (f *fakeSonarCloudBranches) nextBranchCommit() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.branchCommits) == 0 {
		return shortLatestCommit
	}
	commit := f.branchCommits[0]
	if len(f.branchCommits) > 1 {
		f.branchCommits = f.branchCommits[1:]
	}
	return commit
}

func (f *fakeSonarCloudBranches) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		f.mu.Lock()
		f.paths = append(f.paths, r.URL.Path)
		if r.URL.Path == "/api/qualitygates/project_status" {
			f.qualityGateQs = append(f.qualityGateQs, q)
		}
		f.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/project_analyses/search":
			switch q.Get("branch") {
			case shortBranch:
				w.WriteHeader(http.StatusBadRequest)
				_, _ = fmt.Fprintf(w, `{"errors":[{"msg":"Branch '%s' is not of type LONG"}]}`, shortBranch)
				return
			case unlistedBranch:
				w.WriteHeader(http.StatusNotFound)
				_, _ = fmt.Fprintf(w, `{"errors":[{"msg":"Component '%s' on branch '%s' not found"}]}`, revProjectKey, unlistedBranch)
				return
			}
			_, _ = fmt.Fprintf(w, `{"paging":{"pageIndex":1,"pageSize":500,"total":0},"analyses":[]}`)
		case "/api/project_branches/list":
			next := f.nextBranchCommit()
			if next == badGateway {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusBadGateway)
				_, _ = fmt.Fprint(w, "<html><body>502 Bad Gateway</body></html>")
				return
			}
			commit := fmt.Sprintf(`,"commit":{"sha":%q}`, next)
			if strings.HasSuffix(commit, `""}`) {
				commit = ""
			}
			_, _ = fmt.Fprintf(w, `{"branches":[
				{"name":"main","isMain":true,"type":"LONG","analysisDate":"2026-10-23T06:59:29+0000","commit":{"sha":"ce9bf5fdfb6e10b5d2de1b324fd03483e8b67173"}},
				{"name":%q,"isMain":false,"type":"LONG","analysisDate":"2026-10-08T06:59:29+0000","commit":{"sha":"aaaa"}},
				{"name":%q,"isMain":false,"type":"SHORT","mergeBranch":"main","analysisDate":%q%s}]}`,
				longBranch, shortBranch, shortAnalysisDate, commit)
		case "/api/ce/activity":
			// SonarCloud refuses this to a token without project administration.
			w.WriteHeader(http.StatusForbidden)
			_, _ = fmt.Fprint(w, `{"errors":[{"msg":"Insufficient privileges"}]}`)
		case "/api/qualitygates/project_status":
			if q.Get("analysisId") == "" && q.Get("branch") == "" && q.Get("pullRequest") == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = fmt.Fprint(w, `{"errors":[{"msg":"Either 'analysisId', 'projectId' or 'projectKey' must be provided"}]}`)
				return
			}
			_, _ = fmt.Fprint(w, `{"projectStatus":{"status":"OK","conditions":[{"status":"OK","metricKey":"new_reliability_rating","comparator":"GT","periodIndex":1,"errorThreshold":"1","actualValue":"1"}],"periods":[],"ignoredConditions":false}}`)
		default:
			http.NotFound(w, r)
		}
	}
}

func (f *fakeSonarCloudBranches) called(path string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, p := range f.paths {
		if p == path {
			return true
		}
	}
	return false
}

// SonarCloud serves api/project_analyses/search only for long-lived branches, so a
// scan on a short-lived branch is found through the branch's latest analysis.
func TestGetSonarResults_ShortLivedBranch_LatestCommit(t *testing.T) {
	fake := &fakeSonarCloudBranches{}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), "", revProjectKey, srv.URL, shortLatestCommit, "", shortBranch, 5)
	results, err := sc.GetSonarResults(discardLogger())
	if err != nil {
		t.Fatalf("expected the short-lived branch scan to be found, got error: %v", err)
	}
	if results.Revision != shortLatestCommit || results.AnalysedAt != shortAnalysisDate {
		t.Errorf("expected revision %s analysed at %s, got %s at %s", shortLatestCommit, shortAnalysisDate, results.Revision, results.AnalysedAt)
	}
	if results.Branch == nil || results.Branch.Name != shortBranch || results.Branch.Type != "SHORT" {
		t.Errorf("expected branch %s of type SHORT, got %+v", shortBranch, results.Branch)
	}
	if results.QualityGate == nil || results.QualityGate.Status != "OK" || len(results.QualityGate.Conditions) != 1 {
		t.Fatalf("expected the branch's quality gate, got %+v", results.QualityGate)
	}
	if len(fake.qualityGateQs) != 1 {
		t.Fatalf("expected one quality gate request, got %d", len(fake.qualityGateQs))
	}
	q := fake.qualityGateQs[0]
	if q.Get("projectKey") != revProjectKey || q.Get("branch") != shortBranch || q.Has("analysisId") {
		t.Errorf("expected the quality gate by project key and branch, got %v", q)
	}
}

// SonarCloud keeps only the latest analysis of a short-lived branch, so any other
// revision cannot be attested: its results are gone, or not processed yet.
func TestGetSonarResults_ShortLivedBranch_OtherRevision(t *testing.T) {
	const otherRevision = "0000000000000000000000000000000000000000"
	fake := &fakeSonarCloudBranches{}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), "", revProjectKey, srv.URL, otherRevision, "", shortBranch, 5)
	_, err := sc.GetSonarResults(discardLogger())
	if err == nil {
		t.Fatal("expected an error: the branch's latest analysis is of another revision")
	}
	for _, want := range []string{"short-lived", shortBranch, shortLatestCommit, otherRevision} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("expected %q in the error, got: %v", want, err)
		}
	}
	if fake.called("/api/qualitygates/project_status") {
		t.Error("expected no quality gate lookup for a revision that was not found")
	}
}

// A SonarQube error on a branch that is not short-lived is not ours to explain,
// so it reaches the user unchanged.
func TestGetSonarResults_SonarErrorOnOtherBranch_Unchanged(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/project_analyses/search":
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, `{"errors":[{"msg":"Component key 'x' not found"}]}`)
		case "/api/project_branches/list":
			// SonarQube Server has no short-lived branches: every branch is of type BRANCH.
			_, _ = fmt.Fprintf(w, `{"branches":[{"name":%q,"type":"BRANCH","commit":{"sha":%q}}]}`, longBranch, revRevision)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), "", revProjectKey, srv.URL, revRevision, "", longBranch, 5)
	_, err := sc.GetSonarResults(discardLogger())
	if err == nil || err.Error() != "SonarQube error: Component key 'x' not found" {
		t.Errorf("expected the SonarQube error unchanged, got: %v", err)
	}
}

// The branch lookup is only a fallback: a search that succeeds costs no extra call.
func TestGetSonarResults_LongBranch_NoBranchLookup(t *testing.T) {
	fake := &fakeSonarProject{}
	branchLookups := 0
	var qualityGateQ url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/project_branches/list":
			branchLookups++
		case "/api/qualitygates/project_status":
			qualityGateQ = r.URL.Query()
		}
		fake.handler()(w, r)
	}))
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), "", revProjectKey, srv.URL, revRevision, "", revFeatureBranch, 5)
	if _, err := sc.GetSonarResults(discardLogger()); err != nil {
		t.Fatalf("expected the long-lived branch scan to be found, got error: %v", err)
	}
	if branchLookups != 0 {
		t.Errorf("expected no api/project_branches/list call, got %d", branchLookups)
	}
	if qualityGateQ.Get("analysisId") != revAnalysisKey || qualityGateQ.Has("branch") {
		t.Errorf("expected the quality gate of analysis %s, got %v", revAnalysisKey, qualityGateQ)
	}
}

// The quality gate of a short-lived branch is its latest one and cannot be asked
// for by commit. If a newer push is processed between finding the revision and
// reading the gate, the gate is the newer commit's, so the attestation must fail.
func TestGetSonarResults_ShortLivedBranch_NewerAnalysisDuringLookup(t *testing.T) {
	const newerCommit = "1111111111111111111111111111111111111111"
	fake := &fakeSonarCloudBranches{branchCommits: []string{shortLatestCommit, newerCommit}}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), "", revProjectKey, srv.URL, shortLatestCommit, "", shortBranch, 5)
	_, err := sc.GetSonarResults(discardLogger())
	if err == nil {
		t.Fatal("expected an error: the quality gate read may be of the newer commit")
	}
	for _, want := range []string{newerCommit, shortLatestCommit} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("expected %q in the error, got: %v", want, err)
		}
	}
}

// A listed short-lived branch without a commit has no analysis we could attest.
func TestGetSonarResults_ShortLivedBranch_NoCommit(t *testing.T) {
	fake := &fakeSonarCloudBranches{branchCommits: []string{""}}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), "", revProjectKey, srv.URL, shortLatestCommit, "", shortBranch, 5)
	_, err := sc.GetSonarResults(discardLogger())
	if err == nil {
		t.Fatal("expected an error for a branch with no analysed commit")
	}
	if !strings.Contains(err.Error(), "short-lived") || !strings.Contains(err.Error(), "no analysis yet") {
		t.Errorf("expected an error saying the short-lived branch has no analysis yet, got: %v", err)
	}
}

// A branch SonarCloud does not list keeps SonarQube's own error.
func TestGetSonarResults_UnlistedBranch_Unchanged(t *testing.T) {
	fake := &fakeSonarCloudBranches{}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), "", revProjectKey, srv.URL, shortLatestCommit, "", unlistedBranch, 5)
	_, err := sc.GetSonarResults(discardLogger())
	want := fmt.Sprintf("SonarQube error: Component '%s' on branch '%s' not found", revProjectKey, unlistedBranch)
	if err == nil || err.Error() != want {
		t.Errorf("expected %q, got: %v", want, err)
	}
}

// Without --sonar-branch there is no branch to look up, so SonarQube's error stands.
func TestGetSonarResults_SonarErrorWithoutBranch_Unchanged(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/project_branches/list" {
			_, _ = fmt.Fprintf(w, `{"branches":[{"name":%q,"type":"SHORT","commit":{"sha":%q}}]}`, shortBranch, revRevision)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"errors":[{"msg":"Component key 'x' not found"}]}`)
	}))
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), "", revProjectKey, srv.URL, revRevision, "", "", 5)
	_, err := sc.GetSonarResults(discardLogger())
	if err == nil || err.Error() != "SonarQube error: Component key 'x' not found" {
		t.Errorf("expected the SonarQube error unchanged, got: %v", err)
	}
}

// If the second listing cannot confirm the branch's latest commit, the gate just
// read may be another commit's, so nothing is attested.
func TestGetSonarResults_ShortLivedBranch_RecheckFails(t *testing.T) {
	fake := &fakeSonarCloudBranches{branchCommits: []string{shortLatestCommit, badGateway}}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), "", revProjectKey, srv.URL, shortLatestCommit, "", shortBranch, 5)
	_, err := sc.GetSonarResults(discardLogger())
	if err == nil {
		t.Fatal("expected an error: the branch's latest commit could not be confirmed after reading the gate")
	}
	if !strings.Contains(err.Error(), shortBranch) {
		t.Errorf("expected the branch in the error, got: %v", err)
	}
}
