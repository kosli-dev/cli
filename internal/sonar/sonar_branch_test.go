package sonar_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/kosli-dev/cli/internal/sonar"
)

// fakeBranchSonar is a SonarQube stand-in for the project-key/revision path of
// issue #1116. It models the customer's project: the analysis for the revision
// exists only on release/uat, so api/project_analyses/search returns it only when
// the request is scoped to that branch — exactly as SonarQube behaves.
type fakeBranchSonar struct {
	mu             sync.Mutex
	searchBranches []string // branch param of each project_analyses/search request
	searchProjects []string // project param of each project_analyses/search request
	taskBranch     string   // branch reported on the ce/activity task
	taskBranchType string
}

func (f *fakeBranchSonar) handler() http.HandlerFunc {
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
			// The activity list spans every branch of the project, so it also holds
			// tasks for other analyses. The decoy comes first: the task we want has to
			// be selected by analysis ID, not by being the only one there.
			_ = json.NewEncoder(w).Encode(sonar.ActivityResponse{
				Tasks: []sonar.Task{
					{
						TaskID:        "DECOY",
						ComponentName: "customer project",
						ComponentKey:  revProjectKey,
						AnalysisID:    "SOME_OTHER_ANALYSIS",
						Status:        "FAILED",
						Branch:        revMainBranch,
						BranchType:    "LONG",
					},
					{
						TaskID:        revTaskID,
						ComponentName: "customer project",
						ComponentKey:  revProjectKey,
						AnalysisID:    revAnalysisKey,
						Status:        "SUCCESS",
						Branch:        f.taskBranch,
						BranchType:    f.taskBranchType,
					},
				},
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

func (f *fakeBranchSonar) searches() ([]string, []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.searchBranches...), append([]string(nil), f.searchProjects...)
}

// TestGetSonarResults_ProjectKeyPath_WithBranch is the end-to-end wiring for
// --sonar-branch: given the project key, the revision and the branch, the whole
// project-key path succeeds against a project whose analysis is not on the main
// branch, and the branch reaches the attestation payload.
func TestGetSonarResults_ProjectKeyPath_WithBranch(t *testing.T) {
	fake := &fakeBranchSonar{taskBranch: revFeatureBranch, taskBranchType: "LONG"}
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
	if results.Branch == nil || results.Branch.Name != revFeatureBranch {
		t.Errorf("expected branch %q in the attestation payload, got %+v", revFeatureBranch, results.Branch)
	}
	if results.AnalysedAt != revAnalysisDate {
		t.Errorf("expected AnalysedAt=%q, got %q", revAnalysisDate, results.AnalysedAt)
	}
	if results.Revision != revRevision {
		t.Errorf("expected Revision=%q, got %q", revRevision, results.Revision)
	}
	if results.QualityGate == nil || results.QualityGate.Status != "OK" {
		t.Fatalf("expected quality gate OK, got %+v", results.QualityGate)
	}
	if results.TaskID != revTaskID {
		t.Errorf("expected the task for analysis %s, got TaskID %q", revAnalysisKey, results.TaskID)
	}
	if results.Status != "SUCCESS" {
		t.Errorf("expected status SUCCESS from the matched task, got %q", results.Status)
	}
	// The CE task is authoritative once found, and it is the only source of the
	// branch type — the flag supplies a name only.
	if results.Branch.Type != "LONG" {
		t.Errorf("expected branch type from the CE task, got %q", results.Branch.Type)
	}
}

// TestGetSonarResults_ProjectKeyPath_WithoutBranch is the other half of the same
// contract: against the same server, omitting the branch still fails. That is what
// makes the test above evidence that the flag is what fixes #1116, rather than the
// fake being lenient.
func TestGetSonarResults_ProjectKeyPath_WithoutBranch(t *testing.T) {
	fake := &fakeBranchSonar{taskBranch: revFeatureBranch, taskBranchType: "LONG"}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), "", revProjectKey, srv.URL, revRevision, "", "", 5)
	if _, err := sc.GetSonarResults(discardLogger()); err == nil {
		t.Fatal("expected the unscoped search to fail, as it does for the customer in #1116")
	}

	branches, _ := fake.searches()
	if len(branches) != 1 || branches[0] != "" {
		t.Errorf("expected one unscoped search, got %v", branches)
	}
}

// TestGetSonarResults_BranchIgnoredForPullRequest documents that a PR scan is not a
// branch scan: with --pull-request the branch is not used to scope any search.
// The two flags are mutually exclusive at the CLI, so this only pins the library.
func TestGetSonarResults_BranchIgnoredForPullRequest(t *testing.T) {
	fake := &fakeBranchSonar{}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), "", revProjectKey, srv.URL, "", "42", revFeatureBranch, 5)
	// The PR lookup is not served by this fake, so this fails; what matters is that
	// no branch-scoped analyses search was made on the way.
	_, _ = sc.GetSonarResults(discardLogger())

	if branches, _ := fake.searches(); len(branches) != 0 {
		t.Errorf("expected no project_analyses/search on the pull-request path, got %v", branches)
	}
}
