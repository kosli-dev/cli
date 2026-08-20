package sonar_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

// The fixtures below are real responses from the customer instance in issue #1116.
// The project's main branch (master) has never been analysed; the analysis lives on
// release/uat. The slash in the branch name is deliberate — see FuzzBranchParam.
const (
	revProjectKey    = "customer-project"
	revMainBranch    = "master"
	revFeatureBranch = "release/uat"
	revAnalysisKey   = "AaAeAfTdP27JeOuKOycd"
	revAnalysisDate  = "2026-08-20T11:09:48+0400"
	revRevision      = "8700f236fe2fd6c3e2dc5bf33c7e5f3aa8fd3dee"
)

// branchScopedAnalysesServer serves api/project_analyses/search the way SonarQube
// does for the issue #1116 project: unscoped (or scoped to master) the result is
// empty, and the analysis is only returned when the search is scoped to
// release/uat. It records the branch parameter of the last request.
func branchScopedAnalysesServer(t *testing.T, received *string, present *bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/project_analyses/search" {
			http.NotFound(w, r)
			return
		}
		*received = r.URL.Query().Get("branch")
		_, *present = r.URL.Query()["branch"]

		resp := sonar.ProjectAnalyses{Analyses: []sonar.Analysis{}}
		if *received == revFeatureBranch {
			resp.Analyses = []sonar.Analysis{
				{Key: revAnalysisKey, Date: revAnalysisDate, Revision: revRevision},
			}
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

// TestGetProjectAnalysisFromRevision_PassesBranch is the issue #1116 regression:
// when a branch is known, GetProjectAnalysisFromRevision must scope the search to
// it, otherwise SonarQube only searches the main branch and the analysis is never
// found. This is the same defect as #861 in the sibling code path.
func TestGetProjectAnalysisFromRevision_PassesBranch(t *testing.T) {
	var receivedBranch string
	var branchParamPresent bool
	server := branchScopedAnalysesServer(t, &receivedBranch, &branchParamPresent)
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
	if receivedBranch != revFeatureBranch {
		t.Errorf("expected branch=%q to be forwarded to SonarQube, got %q", revFeatureBranch, receivedBranch)
	}
	if analysisID != revAnalysisKey {
		t.Errorf("expected analysis ID %q, got %q", revAnalysisKey, analysisID)
	}
	if sonarResults.AnalysedAt != revAnalysisDate {
		t.Errorf("expected AnalysedAt=%q, got %q", revAnalysisDate, sonarResults.AnalysedAt)
	}
}

// TestGetProjectAnalysisFromRevision_NoBranch pins the backwards-compatible
// behaviour: with no branch known, the branch parameter must be absent from the
// request, not present and empty — an empty branch is not the same query to
// SonarQube as no branch at all.
func TestGetProjectAnalysisFromRevision_NoBranch(t *testing.T) {
	var receivedBranch string
	var branchParamPresent bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBranch = r.URL.Query().Get("branch")
		_, branchParamPresent = r.URL.Query()["branch"]
		resp := sonar.ProjectAnalyses{
			Analyses: []sonar.Analysis{
				{Key: revAnalysisKey, Date: revAnalysisDate, Revision: revRevision},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	sonarResults := &sonar.SonarResults{ServerUrl: server.URL}
	project := &sonar.Project{Key: revProjectKey}

	if _, err := sonar.GetProjectAnalysisFromRevision(http.DefaultClient, sonarResults, project, revRevision, discardLogger()); err != nil {
		t.Fatalf("GetProjectAnalysisFromRevision returned error: %v", err)
	}
	if branchParamPresent {
		t.Errorf("expected no branch param when Branch is nil, got branch=%q", receivedBranch)
	}
}

// TestGetProjectAnalysisFromRevision_EmptyBranchNameNotSent covers the same
// contract for a branch that is set but empty (the flag defaults to ""): still no
// branch parameter.
func TestGetProjectAnalysisFromRevision_EmptyBranchNameNotSent(t *testing.T) {
	var receivedBranch string
	var branchParamPresent bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBranch = r.URL.Query().Get("branch")
		_, branchParamPresent = r.URL.Query()["branch"]
		resp := sonar.ProjectAnalyses{
			Analyses: []sonar.Analysis{
				{Key: revAnalysisKey, Date: revAnalysisDate, Revision: revRevision},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	sonarResults := &sonar.SonarResults{ServerUrl: server.URL, Branch: &sonar.Branch{}}
	project := &sonar.Project{Key: revProjectKey}

	if _, err := sonar.GetProjectAnalysisFromRevision(http.DefaultClient, sonarResults, project, revRevision, discardLogger()); err != nil {
		t.Fatalf("GetProjectAnalysisFromRevision returned error: %v", err)
	}
	if branchParamPresent {
		t.Errorf("expected no branch param for an empty branch name, got branch=%q", receivedBranch)
	}
}

// TestGetProjectAnalysisFromRevision_MainBranchHasNoAnalysis reproduces the
// customer's failure exactly: without the branch, the search comes back empty and
// the command fails even though the analysis exists on release/uat.
func TestGetProjectAnalysisFromRevision_MainBranchHasNoAnalysis(t *testing.T) {
	var receivedBranch string
	var branchParamPresent bool
	server := branchScopedAnalysesServer(t, &receivedBranch, &branchParamPresent)
	defer server.Close()

	project := &sonar.Project{Key: revProjectKey}

	// Scoped to the main branch: nothing there, so this must fail.
	mainResults := &sonar.SonarResults{ServerUrl: server.URL, Branch: &sonar.Branch{Name: revMainBranch}}
	if _, err := sonar.GetProjectAnalysisFromRevision(http.DefaultClient, mainResults, project, revRevision, discardLogger()); err == nil {
		t.Fatal("expected an error when the main branch has no analysis for the revision")
	}

	// Scoped to the branch the scan actually ran on: found.
	branchResults := &sonar.SonarResults{ServerUrl: server.URL, Branch: &sonar.Branch{Name: revFeatureBranch}}
	analysisID, err := sonar.GetProjectAnalysisFromRevision(http.DefaultClient, branchResults, project, revRevision, discardLogger())
	if err != nil {
		t.Fatalf("expected the analysis to be found on %s, got error: %v", revFeatureBranch, err)
	}
	if analysisID != revAnalysisKey {
		t.Errorf("expected analysis ID %q, got %q", revAnalysisKey, analysisID)
	}
}

// TestGetProjectAnalysisFromRevision_WrongRevisionOnBranch guards against the
// branch scoping loosening the revision match: an analysis on the right branch
// with a different revision is not the one we asked for.
func TestGetProjectAnalysisFromRevision_WrongRevisionOnBranch(t *testing.T) {
	var receivedBranch string
	var branchParamPresent bool
	server := branchScopedAnalysesServer(t, &receivedBranch, &branchParamPresent)
	defer server.Close()

	sonarResults := &sonar.SonarResults{
		ServerUrl: server.URL,
		Branch:    &sonar.Branch{Name: revFeatureBranch},
	}
	project := &sonar.Project{Key: revProjectKey}

	_, err := sonar.GetProjectAnalysisFromRevision(http.DefaultClient, sonarResults, project, "0000000000000000000000000000000000000000", discardLogger())
	if err == nil {
		t.Fatal("expected an error when no analysis on the branch matches the revision")
	}
	if sonarResults.AnalysedAt != "" {
		t.Errorf("expected AnalysedAt to stay empty on no match, got %q", sonarResults.AnalysedAt)
	}
}

// TestGetProjectAnalysisFromRevision_NoBranchErrorSuggestsFlag covers the other
// half of #1116: an empty search result is indistinguishable from a permissions
// problem, which is what cost the customer several days. When no branch was given,
// the error has to say that only the main branch was searched.
func TestGetProjectAnalysisFromRevision_NoBranchErrorSuggestsFlag(t *testing.T) {
	var receivedBranch string
	var branchParamPresent bool
	server := branchScopedAnalysesServer(t, &receivedBranch, &branchParamPresent)
	defer server.Close()

	sonarResults := &sonar.SonarResults{ServerUrl: server.URL}
	project := &sonar.Project{Key: revProjectKey}

	_, err := sonar.GetProjectAnalysisFromRevision(http.DefaultClient, sonarResults, project, revRevision, discardLogger())
	if err == nil {
		t.Fatal("expected an error when the main branch has no analysis for the revision")
	}
	if !strings.Contains(err.Error(), "--sonar-branch") {
		t.Errorf("expected the error to point at --sonar-branch, got: %v", err)
	}
	if !strings.Contains(err.Error(), "main branch") {
		t.Errorf("expected the error to say only the main branch was searched, got: %v", err)
	}
}

// TestGetProjectAnalysisFromRevision_BranchErrorNamesBranch is the same contract
// once a branch has been supplied: the error must name the branch that was
// searched, and must not suggest a flag that is already in use.
func TestGetProjectAnalysisFromRevision_BranchErrorNamesBranch(t *testing.T) {
	var receivedBranch string
	var branchParamPresent bool
	server := branchScopedAnalysesServer(t, &receivedBranch, &branchParamPresent)
	defer server.Close()

	sonarResults := &sonar.SonarResults{ServerUrl: server.URL, Branch: &sonar.Branch{Name: revMainBranch}}
	project := &sonar.Project{Key: revProjectKey}

	_, err := sonar.GetProjectAnalysisFromRevision(http.DefaultClient, sonarResults, project, revRevision, discardLogger())
	if err == nil {
		t.Fatal("expected an error when the searched branch has no analysis for the revision")
	}
	if !strings.Contains(err.Error(), revMainBranch) {
		t.Errorf("expected the error to name branch %q, got: %v", revMainBranch, err)
	}
	if strings.Contains(err.Error(), "--sonar-branch") {
		t.Errorf("expected no --sonar-branch suggestion when a branch was given, got: %v", err)
	}
}
