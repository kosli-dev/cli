package sonar_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/kosli-dev/cli/internal/sonar"
)

const (
	shortTaskID     = "AaFnjwU6BCVWgkk-qOYR"
	shortAnalysisID = "99176d9b-f4f0-4dc4-8dc6-e40d15f3ad97"
)

// fakeSonarCloudCETask adds the CE endpoints to fakeSonarCloudBranches. The task is
// the scan of the short-lived branch; api/ce/component answers with currentTaskID as
// the project's latest finished task (none when empty), and a queued task with
// queuedStatus when that is set, as SonarCloud does to a token with Browse.
type fakeSonarCloudCETask struct {
	branches      fakeSonarCloudBranches
	currentTaskID string
	queuedStatus  string
	// queuedBranch is the queued task's branch: ours when empty, none when "-",
	// pull request 7 when "pr".
	queuedBranch         string
	ceComponentForbidden bool
	mu                   sync.Mutex
	paths                []string
}

func (f *fakeSonarCloudCETask) handler() http.HandlerFunc {
	branches := f.branches.handler()
	return func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.paths = append(f.paths, r.URL.Path)
		f.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/ce/task":
			_, _ = fmt.Fprintf(w, `{"task":{"organization":"cyber-dojo","id":%q,"type":"REPORT","componentKey":%q,"componentName":"customer project","componentQualifier":"TRK","analysisId":%q,"status":"SUCCESS","submittedAt":"2026-10-20T07:06:20+0000","branch":%q,"branchType":"SHORT"}}`,
				shortTaskID, revProjectKey, shortAnalysisID, shortBranch)
		case "/api/ce/component":
			if f.ceComponentForbidden {
				w.WriteHeader(http.StatusForbidden)
				_, _ = fmt.Fprint(w, `{"errors":[{"msg":"Insufficient privileges"}]}`)
				return
			}
			queue := ""
			if f.queuedStatus != "" {
				branch := fmt.Sprintf(`,"branch":%q,"branchType":"SHORT"`, shortBranch)
				switch f.queuedBranch {
				case "":
				case "-":
					branch = ""
				case "pr":
					branch = `,"pullRequest":"7"`
				default:
					branch = fmt.Sprintf(`,"branch":%q,"branchType":"SHORT"`, f.queuedBranch)
				}
				queue = fmt.Sprintf(`{"id":"AaFnz-queued","type":"REPORT","componentKey":%q,"status":%q%s}`, revProjectKey, f.queuedStatus, branch)
			}
			current := ""
			if f.currentTaskID != "" {
				current = fmt.Sprintf(`,"current":{"id":%q,"type":"REPORT","componentKey":%q,"analysisId":"another-analysis","status":"SUCCESS","branch":"main","branchType":"LONG"}`, f.currentTaskID, revProjectKey)
			}
			_, _ = fmt.Fprintf(w, `{"queue":[%s]%s}`, queue, current)
		default:
			branches(w, r)
		}
	}
}

func (f *fakeSonarCloudCETask) calls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.paths)
}

// newCETaskConfig finds the scan the way the scanner leaves it: through
// report-task.txt, or through --sonar-ce-task-url with no report-task.txt.
func newCETaskConfig(t *testing.T, serverURL string, viaReportTask bool) *sonar.SonarConfig {
	t.Helper()
	dir := t.TempDir()
	ceTaskURL := serverURL + "/api/ce/task?id=" + shortTaskID
	if viaReportTask {
		report := fmt.Sprintf("serverUrl=%s\ndashboardUrl=%s/dashboard?id=%s\nceTaskUrl=%s\n", serverURL, serverURL, revProjectKey, ceTaskURL)
		if err := os.WriteFile(filepath.Join(dir, "report-task.txt"), []byte(report), 0o600); err != nil {
			t.Fatal(err)
		}
		ceTaskURL = ""
	}
	// The revision is the CI commit the command defaults to, which this path ignores.
	return sonar.NewSonarConfig("tok", dir, ceTaskURL, "", "", revRevision, "", "", 5)
}

// SonarCloud refuses api/project_analyses/search for a short-lived branch, and lists
// only the branch's latest analysis. That is our scan's when no task of the project
// finished after ours, which api/ce/component tells a token with Browse.
func TestGetSonarResults_CETask_ShortLivedBranch(t *testing.T) {
	for name, viaReportTask := range map[string]bool{"report-task.txt": true, "--sonar-ce-task-url": false} {
		t.Run(name, func(t *testing.T) {
			fake := &fakeSonarCloudCETask{currentTaskID: shortTaskID}
			srv := httptest.NewServer(fake.handler())
			defer srv.Close()

			results, err := newCETaskConfig(t, srv.URL, viaReportTask).GetSonarResults(discardLogger())
			if err != nil {
				t.Fatalf("expected the short-lived branch scan to be found, got error: %v", err)
			}
			if results.Revision != shortLatestCommit || results.AnalysedAt != shortAnalysisDate {
				t.Errorf("expected revision %s analysed at %s, got %s at %s", shortLatestCommit, shortAnalysisDate, results.Revision, results.AnalysedAt)
			}
			if results.TaskID != shortTaskID || results.Branch == nil || results.Branch.Name != shortBranch || results.Branch.Type != "SHORT" {
				t.Errorf("expected task %s on branch %s of type SHORT, got task %s, branch %+v", shortTaskID, shortBranch, results.TaskID, results.Branch)
			}
			if len(fake.branches.qualityGateQs) != 1 || fake.branches.qualityGateQs[0].Get("analysisId") != shortAnalysisID || fake.branches.qualityGateQs[0].Has("branch") {
				t.Errorf("expected one quality gate request for analysis %s, got %v", shortAnalysisID, fake.branches.qualityGateQs)
			}
			// A task finishing between the two reads must be caught, so the task check
			// has to come after the branch listing.
			calls := fake.calls()
			listed, checked := slices.Index(calls, "/api/project_branches/list"), slices.Index(calls, "/api/ce/component")
			if listed < 0 || checked < listed {
				t.Errorf("expected api/ce/component after api/project_branches/list, got %v", calls)
			}
		})
	}
}

// Another task finished after ours, so the branch's latest analysis may not be ours.
func TestGetSonarResults_CETask_ShortLivedBranch_NotLatestTask(t *testing.T) {
	fake := &fakeSonarCloudCETask{currentTaskID: "AaFnz-a-later-task"}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	_, err := newCETaskConfig(t, srv.URL, true).GetSonarResults(discardLogger())
	if err == nil {
		t.Fatal("expected an error: a later task means the branch's latest analysis may be another commit's")
	}
	for _, want := range []string{"short-lived", shortBranch, shortTaskID, "--sonar-project-key", "--sonar-revision"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("expected %q in the error, got: %v", want, err)
		}
	}
	if len(fake.branches.qualityGateQs) != 0 {
		t.Error("expected no quality gate lookup")
	}
}

// SonarQube saves a scan's analysis before it records the task as finished, so a task
// in progress may already have replaced the branch's latest analysis.
func TestGetSonarResults_CETask_ShortLivedBranch_TaskInProgress(t *testing.T) {
	fake := &fakeSonarCloudCETask{currentTaskID: shortTaskID, queuedStatus: "IN_PROGRESS"}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	_, err := newCETaskConfig(t, srv.URL, true).GetSonarResults(discardLogger())
	if err == nil || !strings.Contains(err.Error(), "being processed") {
		t.Errorf("expected an error: a task in progress may have replaced the branch's latest analysis, got: %v", err)
	}
}

// SonarQube runs one task of a project at a time, so while a task of another branch
// is in progress, none of ours can be, and our branch's latest analysis stands.
func TestGetSonarResults_CETask_ShortLivedBranch_OtherBranchTaskInProgress(t *testing.T) {
	fake := &fakeSonarCloudCETask{currentTaskID: shortTaskID, queuedStatus: "IN_PROGRESS", queuedBranch: "renovate/other"}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	results, err := newCETaskConfig(t, srv.URL, true).GetSonarResults(discardLogger())
	if err != nil || results.Revision != shortLatestCommit {
		t.Errorf("expected revision %s, got %+v, error: %v", shortLatestCommit, results, err)
	}
}

// A pull request scan never changes a branch's latest analysis.
func TestGetSonarResults_CETask_ShortLivedBranch_PullRequestTaskInProgress(t *testing.T) {
	fake := &fakeSonarCloudCETask{currentTaskID: shortTaskID, queuedStatus: "IN_PROGRESS", queuedBranch: "pr"}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	results, err := newCETaskConfig(t, srv.URL, true).GetSonarResults(discardLogger())
	if err != nil || results.Revision != shortLatestCommit {
		t.Errorf("expected revision %s, got %+v, error: %v", shortLatestCommit, results, err)
	}
}

// A task in progress that names no branch or pull request may be on ours.
func TestGetSonarResults_CETask_ShortLivedBranch_UnknownBranchTaskInProgress(t *testing.T) {
	fake := &fakeSonarCloudCETask{currentTaskID: shortTaskID, queuedStatus: "IN_PROGRESS", queuedBranch: "-"}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	_, err := newCETaskConfig(t, srv.URL, true).GetSonarResults(discardLogger())
	if err == nil || !strings.Contains(err.Error(), "being processed") {
		t.Errorf("expected an error: the task in progress may be on our branch, got: %v", err)
	}
}

// A pending task has not started, so it cannot have changed the branch yet.
func TestGetSonarResults_CETask_ShortLivedBranch_TaskPending(t *testing.T) {
	fake := &fakeSonarCloudCETask{currentTaskID: shortTaskID, queuedStatus: "PENDING"}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	results, err := newCETaskConfig(t, srv.URL, true).GetSonarResults(discardLogger())
	if err != nil || results.Revision != shortLatestCommit {
		t.Errorf("expected revision %s, got %+v, error: %v", shortLatestCommit, results, err)
	}
}

func TestGetSonarResults_CETask_ShortLivedBranch_NoFinishedTask(t *testing.T) {
	fake := &fakeSonarCloudCETask{}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	_, err := newCETaskConfig(t, srv.URL, true).GetSonarResults(discardLogger())
	if err == nil || !strings.Contains(err.Error(), shortTaskID) {
		t.Errorf("expected an error naming task %s, got: %v", shortTaskID, err)
	}
}

func TestGetSonarResults_CETask_ShortLivedBranch_NoCommit(t *testing.T) {
	fake := &fakeSonarCloudCETask{currentTaskID: shortTaskID, branches: fakeSonarCloudBranches{branchCommits: []string{""}}}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	_, err := newCETaskConfig(t, srv.URL, true).GetSonarResults(discardLogger())
	if err == nil || !strings.Contains(err.Error(), "short-lived") || !strings.Contains(err.Error(), "no analysis yet") {
		t.Errorf("expected an error saying the short-lived branch has no analysis yet, got: %v", err)
	}
}

func TestGetSonarResults_CETask_ShortLivedBranch_TaskCheckFails(t *testing.T) {
	fake := &fakeSonarCloudCETask{ceComponentForbidden: true}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	_, err := newCETaskConfig(t, srv.URL, true).GetSonarResults(discardLogger())
	if err == nil || !strings.Contains(err.Error(), "Insufficient privileges") {
		t.Errorf("expected SonarQube's refusal of api/ce/component, got: %v", err)
	}
	if len(fake.branches.qualityGateQs) != 0 {
		t.Error("expected no quality gate lookup")
	}
}

// When the branch can't be listed, SonarQube's own error stands.
func TestGetSonarResults_CETask_ShortLivedBranch_ListingFails(t *testing.T) {
	fake := &fakeSonarCloudCETask{currentTaskID: shortTaskID, branches: fakeSonarCloudBranches{branchCommits: []string{badGateway}}}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	_, err := newCETaskConfig(t, srv.URL, true).GetSonarResults(discardLogger())
	want := fmt.Sprintf("SonarQube error: Branch '%s' is not of type LONG", shortBranch)
	if err == nil || err.Error() != want {
		t.Errorf("expected %q, got: %v", want, err)
	}
	if slices.Contains(fake.calls(), "/api/ce/component") {
		t.Error("expected no task check without a short-lived branch")
	}
}

// A CE task with no branch is the main branch's: there is no short-lived branch to
// look up, so SonarQube's error stands.
func TestGetSonarResults_CETask_NoBranch_SonarErrorUnchanged(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/ce/task":
			_, _ = fmt.Fprintf(w, `{"task":{"id":%q,"componentKey":%q,"analysisId":%q,"status":"SUCCESS"}}`, shortTaskID, revProjectKey, shortAnalysisID)
		case "/api/project_analyses/search":
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, `{"errors":[{"msg":"Component key 'x' not found"}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	_, err := newCETaskConfig(t, srv.URL, false).GetSonarResults(discardLogger())
	if err == nil || err.Error() != "SonarQube error: Component key 'x' not found" {
		t.Errorf("expected the SonarQube error unchanged, got: %v", err)
	}
}
