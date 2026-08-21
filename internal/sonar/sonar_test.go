package sonar_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/kosli-dev/cli/internal/sonar"
)

func TestPlaceholder(t *testing.T) {
	// We use the tool cover for coverage, but if there is no _test.go file, then
	// Go use the tool covdata. At the same time, they removed covdata as a precompiled
	// binary in the distribution. This made the coverage calculation fail for some of us.
}

// TestGetProjectAnalysisFromAnalysisID_PassesBranch verifies that when the scan ran on
// a non-default branch, GetProjectAnalysisFromAnalysisID forwards the branch name to
// SonarQube's api/project_analyses/search. Without it, SonarQube only returns analyses
// for the main branch and the analysis-ID lookup fails (issue #861).
func TestGetProjectAnalysisFromAnalysisID_PassesBranch(t *testing.T) {
	const (
		wantAnalysisID = "AYxxxxxxxxxxxxxxxxxx"
		wantBranch     = "release/11.2.0"
		wantDate       = "2026-05-06T19:00:00+0000"
		wantRevision   = "abc1234def5678"
	)

	var receivedBranch string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/project_analyses/search" {
			http.NotFound(w, r)
			return
		}
		receivedBranch = r.URL.Query().Get("branch")
		resp := sonar.ProjectAnalyses{}
		if receivedBranch == wantBranch {
			resp.Analyses = []sonar.Analysis{
				{Key: wantAnalysisID, Date: wantDate, Revision: wantRevision},
			}
		} else {
			// Simulate SonarQube's default behaviour: returns main-branch analyses only.
			resp.Analyses = []sonar.Analysis{
				{Key: "MAIN_ANALYSIS_KEY", Date: "2026-05-01T00:00:00+0000", Revision: "deadbeef"},
			}
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	sonarResults := &sonar.SonarResults{
		ServerUrl: server.URL,
		Branch:    &sonar.Branch{Name: wantBranch, Type: "LONG"},
	}
	project := &sonar.Project{Key: "my-project"}

	err := sonar.GetProjectAnalysisFromAnalysisID(http.DefaultClient, sonarResults, project, wantAnalysisID)
	if err != nil {
		t.Fatalf("GetProjectAnalysisFromAnalysisID returned error: %v", err)
	}
	if receivedBranch != wantBranch {
		t.Errorf("expected branch=%q to be forwarded to SonarQube, got %q", wantBranch, receivedBranch)
	}
	if sonarResults.AnalysedAt != wantDate {
		t.Errorf("expected AnalysedAt=%q, got %q", wantDate, sonarResults.AnalysedAt)
	}
	if sonarResults.Revision != wantRevision {
		t.Errorf("expected Revision=%q, got %q", wantRevision, sonarResults.Revision)
	}
}

// TestGetProjectAnalysisFromAnalysisID_NoBranch verifies that when no branch was
// recorded on the CE task (main-branch scan), no branch param is sent — preserving
// existing behaviour.
func TestGetProjectAnalysisFromAnalysisID_NoBranch(t *testing.T) {
	const wantAnalysisID = "MAIN_ANALYSIS_KEY"

	var receivedBranch string
	var branchParamPresent bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBranch = r.URL.Query().Get("branch")
		_, branchParamPresent = r.URL.Query()["branch"]
		resp := sonar.ProjectAnalyses{
			Analyses: []sonar.Analysis{
				{Key: wantAnalysisID, Date: "2026-05-01T00:00:00+0000", Revision: "deadbeef"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	sonarResults := &sonar.SonarResults{ServerUrl: server.URL}
	project := &sonar.Project{Key: "my-project"}

	err := sonar.GetProjectAnalysisFromAnalysisID(http.DefaultClient, sonarResults, project, wantAnalysisID)
	if err != nil {
		t.Fatalf("GetProjectAnalysisFromAnalysisID returned error: %v", err)
	}
	if branchParamPresent {
		t.Errorf("expected no branch param when Branch is nil, got branch=%q", receivedBranch)
	}
}

// Fixtures from the customer instance in issue #1116: the project's main branch
// has never been analysed, and the analysis lives on release/uat. The slash in the
// branch name is deliberate — it is special in a query string.
const (
	revProjectKey    = "customer-project"
	revFeatureBranch = "release/uat"
	revAnalysisKey   = "AaAeAfTdP27JeOuKOycd"
	revTaskID        = "AaAeAfTdP27JeOuKOyce"
	revAnalysisDate  = "2026-08-20T11:09:48+0400"
	revRevision      = "8700f236fe2fd6c3e2dc5bf33c7e5f3aa8fd3dee"
	revPullRequest   = "42"
)

// fakeSonarProject serves the endpoints the project-key path uses, the way
// SonarQube behaves for the issue #1116 project: the analysis for the revision is
// returned only when the search is scoped to release/uat.
type fakeSonarProject struct {
	mu             sync.Mutex
	searchBranches []string
	searchProjects []string
}

func (f *fakeSonarProject) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/project_analyses/search":
			branch := r.URL.Query().Get("branch")
			f.mu.Lock()
			f.searchBranches = append(f.searchBranches, branch)
			f.searchProjects = append(f.searchProjects, r.URL.Query().Get("project"))
			f.mu.Unlock()

			resp := sonar.ProjectAnalyses{Analyses: []sonar.Analysis{}}
			if branch == revFeatureBranch {
				resp.Analyses = []sonar.Analysis{
					{Key: revAnalysisKey, Date: revAnalysisDate, Revision: revRevision},
				}
			}
			_ = json.NewEncoder(w).Encode(resp)
		case "/api/ce/activity":
			// The activity list spans every branch of the project, so the decoy comes
			// first: the task must be selected by analysis ID, not by being the only one.
			_ = json.NewEncoder(w).Encode(sonar.ActivityResponse{Tasks: []sonar.Task{
				{TaskID: "DECOY", AnalysisID: "SOME_OTHER_ANALYSIS", Status: "FAILED", ComponentKey: revProjectKey},
				{TaskID: revTaskID, AnalysisID: revAnalysisKey, Status: "SUCCESS", ComponentKey: revProjectKey,
					ComponentName: "customer project", Branch: revFeatureBranch, BranchType: "LONG"},
			}})
		case "/api/project_pull_requests/list":
			_ = json.NewEncoder(w).Encode(sonar.PullRequestsResponse{
				PullRequests: []sonar.PullRequestInfo{{
					Key: revPullRequest, AnalysisDate: revAnalysisDate,
					Commit: sonar.PRCommit{SHA: revRevision},
				}},
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

func (f *fakeSonarProject) searches() ([]string, []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.searchBranches), slices.Clone(f.searchProjects)
}

// TestGetProjectAnalysisFromRevision_PassesBranch is the issue #1116 regression:
// without the branch, SonarQube searches only the main branch and the analysis is
// never found. This is the same defect as #861 in the sibling code path.
func TestGetProjectAnalysisFromRevision_PassesBranch(t *testing.T) {
	fake := &fakeSonarProject{}
	server := httptest.NewServer(fake.handler())
	defer server.Close()

	sonarResults := &sonar.SonarResults{
		ServerUrl: server.URL,
		Branch:    &sonar.Branch{Name: revFeatureBranch},
	}
	project := &sonar.Project{Key: revProjectKey}

	analysisID, err := sonar.GetProjectAnalysisFromRevision(http.DefaultClient, sonarResults, project, revRevision, discardLogger())
	if err != nil {
		t.Fatalf("GetProjectAnalysisFromRevision returned error: %v", err)
	}
	if analysisID != revAnalysisKey {
		t.Errorf("expected analysis ID %q, got %q", revAnalysisKey, analysisID)
	}
	if sonarResults.AnalysedAt != revAnalysisDate {
		t.Errorf("expected AnalysedAt=%q, got %q", revAnalysisDate, sonarResults.AnalysedAt)
	}
}

// TestGetProjectAnalysisFromRevision_BranchParam pins what reaches SonarQube for a
// given branch: the value arrives exactly as supplied, an unset branch sends no
// branch parameter at all (not an empty one), and a branch name cannot alter
// another query parameter.
func TestGetProjectAnalysisFromRevision_BranchParam(t *testing.T) {
	cases := []struct {
		name        string
		branch      *sonar.Branch
		wantBranch  string
		wantPresent bool
	}{
		{name: "branch forwarded verbatim", branch: &sonar.Branch{Name: revFeatureBranch}, wantBranch: revFeatureBranch, wantPresent: true},
		{name: "no branch known sends no parameter", branch: nil, wantPresent: false},
		{name: "empty branch name sends no parameter", branch: &sonar.Branch{}, wantPresent: false},
		{name: "branch cannot inject another parameter", branch: &sonar.Branch{Name: "x&project=other"}, wantBranch: "x&project=other", wantPresent: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var gotBranch, gotProject string
			var present bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotBranch = r.URL.Query().Get("branch")
				gotProject = r.URL.Query().Get("project")
				_, present = r.URL.Query()["branch"]
				_ = json.NewEncoder(w).Encode(sonar.ProjectAnalyses{Analyses: []sonar.Analysis{
					{Key: revAnalysisKey, Date: revAnalysisDate, Revision: revRevision},
				}})
			}))
			defer server.Close()

			sonarResults := &sonar.SonarResults{ServerUrl: server.URL, Branch: c.branch}
			project := &sonar.Project{Key: revProjectKey}
			_, err := sonar.GetProjectAnalysisFromRevision(http.DefaultClient, sonarResults, project, revRevision, discardLogger())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if present != c.wantPresent {
				t.Errorf("branch parameter present=%t, want %t (got %q)", present, c.wantPresent, gotBranch)
			}
			if c.wantPresent && gotBranch != c.wantBranch {
				t.Errorf("SonarQube received branch %q, want %q", gotBranch, c.wantBranch)
			}
			if gotProject != revProjectKey {
				t.Errorf("project parameter altered: got %q, want %q", gotProject, revProjectKey)
			}
		})
	}
}

// TestGetProjectAnalysisFromRevision_NotFoundError covers the other half of #1116:
// an empty result is indistinguishable from a permissions problem, which is what
// cost the customer several days while their token was correct throughout.
func TestGetProjectAnalysisFromRevision_NotFoundError(t *testing.T) {
	cases := []struct {
		name       string
		branch     *sonar.Branch
		wantInErr  string
		wantNotErr string
	}{
		{name: "no branch given points at the flag", branch: nil, wantInErr: "--sonar-branch"},
		{name: "branch given names the branch", branch: &sonar.Branch{Name: revFeatureBranch}, wantInErr: revFeatureBranch, wantNotErr: "--sonar-branch"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(sonar.ProjectAnalyses{Analyses: []sonar.Analysis{}})
			}))
			defer server.Close()

			sonarResults := &sonar.SonarResults{ServerUrl: server.URL, Branch: c.branch}
			project := &sonar.Project{Key: revProjectKey}
			_, err := sonar.GetProjectAnalysisFromRevision(http.DefaultClient, sonarResults, project, revRevision, discardLogger())
			if err == nil {
				t.Fatal("expected an error when no analysis matches")
			}
			if !strings.Contains(err.Error(), c.wantInErr) {
				t.Errorf("expected %q in the error, got: %v", c.wantInErr, err)
			}
			if c.wantNotErr != "" && strings.Contains(err.Error(), c.wantNotErr) {
				t.Errorf("did not expect %q in the error, got: %v", c.wantNotErr, err)
			}
		})
	}
}

// TestGetSonarResults_ProjectKeyPath_WithBranch is the end-to-end wiring for
// --sonar-branch: the whole project-key path succeeds against a project whose
// analysis is not on the main branch.
func TestGetSonarResults_ProjectKeyPath_WithBranch(t *testing.T) {
	fake := &fakeSonarProject{}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), "", revProjectKey, srv.URL, revRevision, "", revFeatureBranch, 5)
	results, err := sc.GetSonarResults(discardLogger())
	if err != nil {
		t.Fatalf("expected the scan on %s to be found, got error: %v", revFeatureBranch, err)
	}

	branches, projects := fake.searches()
	if len(branches) != 1 || branches[0] != revFeatureBranch {
		t.Errorf("expected one search scoped to branch %q, got %v", revFeatureBranch, branches)
	}
	if len(projects) != 1 || projects[0] != revProjectKey {
		t.Errorf("expected project %q on the search, got %v", revProjectKey, projects)
	}
	if results.Revision != revRevision {
		t.Errorf("expected Revision=%q, got %q", revRevision, results.Revision)
	}
	if results.TaskID != revTaskID {
		t.Errorf("expected the task for analysis %s, got TaskID %q", revAnalysisKey, results.TaskID)
	}
	if results.QualityGate == nil || results.QualityGate.Status != "OK" {
		t.Fatalf("expected quality gate OK, got %+v", results.QualityGate)
	}
	// Type can only come from the matched CE task: the flag sets a name and no type.
	if results.Branch == nil || results.Branch.Name != revFeatureBranch || results.Branch.Type != "LONG" {
		t.Errorf("expected branch %s with the type from the CE task, got %+v", revFeatureBranch, results.Branch)
	}
}

// TestGetSonarResults_ProjectKeyPath_WithoutBranch is the other half of the same
// contract: against the same fake, omitting the branch still fails. That is what
// makes the test above evidence about SonarQube rather than a lenient fake.
func TestGetSonarResults_ProjectKeyPath_WithoutBranch(t *testing.T) {
	fake := &fakeSonarProject{}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), "", revProjectKey, srv.URL, revRevision, "", "", 5)
	_, err := sc.GetSonarResults(discardLogger())
	if err == nil {
		t.Fatal("expected the unscoped search to fail, as it does for the customer in #1116")
	}
	if !strings.Contains(err.Error(), "--sonar-branch") {
		t.Errorf("expected the error to point at --sonar-branch, got: %v", err)
	}

	if branches, _ := fake.searches(); len(branches) != 1 || branches[0] != "" {
		t.Errorf("expected one unscoped search, got %v", branches)
	}
}

// TestGetSonarResults_BranchIgnoredForPullRequest is the invalid state the guard
// prevents: a pull request scan is not a branch scan, so an attestation must not
// carry both. The CLI blocks the flag combination; nothing stopped the library.
func TestGetSonarResults_BranchIgnoredForPullRequest(t *testing.T) {
	fake := &fakeSonarProject{}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), "", revProjectKey, srv.URL, "", revPullRequest, revFeatureBranch, 5)
	results, err := sc.GetSonarResults(discardLogger())
	if err != nil {
		t.Fatalf("expected the pull-request scan to be found, got error: %v", err)
	}

	if branches, _ := fake.searches(); len(branches) != 0 {
		t.Errorf("expected no project_analyses/search on the pull-request path, got %v", branches)
	}
	if results.PullRequest != revPullRequest {
		t.Errorf("expected pull request %q, got %q", revPullRequest, results.PullRequest)
	}
	if results.Branch != nil {
		t.Errorf("expected no branch alongside a pull request, got %+v", results.Branch)
	}
}
