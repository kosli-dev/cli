package sonar_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
