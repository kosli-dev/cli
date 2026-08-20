package sonar_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/sonar"
)

// bufferLogger returns a logger whose warnings can be asserted on. Warn writes to
// the error stream.
func bufferLogger() (*logger.Logger, *bytes.Buffer) {
	stderr := &bytes.Buffer{}
	return logger.NewLogger(&bytes.Buffer{}, stderr, false), stderr
}

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
	// omitMatchingTask serves activity without the task for our analysis, as a
	// bounded recent-activity list eventually does.
	omitMatchingTask bool
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
			tasks := []sonar.Task{
				{
					TaskID:        "DECOY",
					ComponentName: "customer project",
					ComponentKey:  revProjectKey,
					AnalysisID:    "SOME_OTHER_ANALYSIS",
					Status:        "FAILED",
					Branch:        revMainBranch,
					BranchType:    "LONG",
				},
			}
			if !f.omitMatchingTask {
				tasks = append(tasks, sonar.Task{
					TaskID:        revTaskID,
					ComponentName: "customer project",
					ComponentKey:  revProjectKey,
					AnalysisID:    revAnalysisKey,
					Status:        "SUCCESS",
					Branch:        f.taskBranch,
					BranchType:    f.taskBranchType,
				})
			}
			_ = json.NewEncoder(w).Encode(sonar.ActivityResponse{Tasks: tasks})
		case "/api/project_pull_requests/list":
			_ = json.NewEncoder(w).Encode(sonar.PullRequestsResponse{
				PullRequests: []sonar.PullRequestInfo{{
					Key:          revPullRequest,
					Branch:       "feature/whatever",
					AnalysisDate: revAnalysisDate,
					Commit:       sonar.PRCommit{SHA: revRevision},
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

	log, stderr := bufferLogger()
	sc := sonar.NewSonarConfig("tok", t.TempDir(), "", revProjectKey, srv.URL, revRevision, "", revFeatureBranch, 5)
	results, err := sc.GetSonarResults(log)
	if err != nil {
		t.Fatalf("expected the scan on %s to be found, got error: %v", revFeatureBranch, err)
	}
	// The happy path warns about nothing: neither an unmatched task nor an ignored
	// flag, both of which are warned about elsewhere.
	if stderr.Len() != 0 {
		t.Errorf("expected no warnings on the happy path, got stderr: %q", stderr.String())
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

// TestGetSonarResults_BranchIgnoredForPullRequest is the invalid state the guard
// exists to prevent: a PR scan is not a branch scan, so an attestation must not
// carry both. The CLI blocks the flag combination, but nothing stopped the
// library from producing it — and the CE task only clears the branch when it
// matches a task, which api/ce/activity (a bounded recent-activity list) need
// not contain.
func TestGetSonarResults_BranchIgnoredForPullRequest(t *testing.T) {
	fake := &fakeBranchSonar{}
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

// TestGetSonarResults_TaskWithoutBranch_KeepsSuppliedBranch pins which side wins
// when the CE task reports no branch at all — SonarQube omits it for main-branch
// tasks, and older self-hosted Servers omit it more widely. The task is
// authoritative for the branch type, but it must not delete the branch the user
// gave us to find the analysis with.
func TestGetSonarResults_TaskWithoutBranch_KeepsSuppliedBranch(t *testing.T) {
	fake := &fakeBranchSonar{taskBranch: "", taskBranchType: ""}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	sc := sonar.NewSonarConfig("tok", t.TempDir(), "", revProjectKey, srv.URL, revRevision, "", revFeatureBranch, 5)
	results, err := sc.GetSonarResults(discardLogger())
	if err != nil {
		t.Fatalf("expected the scan to be found, got error: %v", err)
	}

	if results.Branch == nil || results.Branch.Name != revFeatureBranch {
		t.Fatalf("expected the supplied branch %q to survive a task that reports none, got %+v", revFeatureBranch, results.Branch)
	}
	if results.Branch.Type != "" {
		t.Errorf("expected no branch type when the task reports none, got %q", results.Branch.Type)
	}
	if results.TaskID != revTaskID {
		t.Errorf("expected the task to still be matched, got TaskID %q", results.TaskID)
	}
}

// TestGetSonarResults_NoMatchingTask_Warns covers the silent gap: api/ce/activity
// cannot be scoped to a branch and is a bounded recent-activity list, so it may
// simply not hold the task. The attestation is then published with no task ID
// and no scan status, which used to happen without a word.
func TestGetSonarResults_NoMatchingTask_Warns(t *testing.T) {
	fake := &fakeBranchSonar{taskBranch: revFeatureBranch, taskBranchType: "LONG", omitMatchingTask: true}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	log, stderr := bufferLogger()
	sc := sonar.NewSonarConfig("tok", t.TempDir(), "", revProjectKey, srv.URL, revRevision, "", revFeatureBranch, 5)
	results, err := sc.GetSonarResults(log)
	if err != nil {
		t.Fatalf("expected the scan to be found, got error: %v", err)
	}

	if results.TaskID != "" {
		t.Errorf("expected no task ID when no task matched, got %q", results.TaskID)
	}
	if !strings.Contains(stderr.String(), "no SonarQube compute engine task") {
		t.Errorf("expected a warning that no task matched, got stderr: %q", stderr.String())
	}
	// The branch still has to survive: it is what found the analysis.
	if results.Branch == nil || results.Branch.Name != revFeatureBranch {
		t.Errorf("expected the supplied branch to survive, got %+v", results.Branch)
	}
}

// TestGetSonarResults_BranchIgnoredOnCETaskPath_Warns covers the flag being a
// no-op: identified by report-task.txt or --sonar-ce-task-url, the branch comes
// from the scan task and --sonar-branch reaches nothing. Silently ignoring it is
// the same class of unexplained outcome this flag exists to remove.
func TestGetSonarResults_BranchIgnoredOnCETaskPath_Warns(t *testing.T) {
	fake := &fakeSonar{acceptsBearer: true, acceptsBasic: true}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	log, stderr := bufferLogger()
	sc := sonar.NewSonarConfig("tok", t.TempDir(), srv.URL+"/api/ce/task?id=AYx", "", "", "", "", revFeatureBranch, 5)
	if _, err := sc.GetSonarResults(log); err != nil {
		t.Fatalf("expected the CE task path to succeed, got error: %v", err)
	}

	if !strings.Contains(stderr.String(), "--sonar-branch is ignored") {
		t.Errorf("expected a warning that --sonar-branch is ignored on this path, got stderr: %q", stderr.String())
	}
}
