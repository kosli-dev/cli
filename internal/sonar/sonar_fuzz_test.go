package sonar_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/kosli-dev/cli/internal/sonar"
)

// FuzzBranchParam checks two properties of branch handling on the
// project_analyses/search request, for any branch name at all:
//
//  1. round-trip — the branch SonarQube receives is byte-for-byte the branch the
//     user supplied. Branch names routinely contain characters that are special in
//     a query string (release/uat, feature/ABC-123, names with spaces).
//  2. no query-parameter injection — a branch name cannot alter another parameter.
//     A branch called "x&project=other" must not change which project is searched.
//
// url.Values.Encode() gives us both. The point of the fuzz test is to keep them:
// a later refactor to string concatenation would reintroduce the injection, and
// this is what would catch it.
func FuzzBranchParam(f *testing.F) {
	seeds := []string{
		"release/uat",       // the branch from #1116 — a slash is the common case
		"feature/ABC-123_x", // typical Jira-derived branch name
		"main",              // plain
		"master",            //
		"a b",               // space
		"a&b=c",             // separator and assignment
		"x&project=other",   // the injection attempt this test exists for
		"a#b",               // fragment
		"a?b",               // query start
		"a%2Fb",             // already percent-encoded: must not be double-decoded
		"a+b",               // plus, which decodes to a space if mishandled
		"ünïcøde",           // non-ASCII
		"",                  // unset
		strings.Repeat("longer", 200),
	}
	for _, s := range seeds {
		f.Add(s)
	}

	// One server for the whole run, not one per input: a fuzz worker executes
	// thousands of inputs, and a server each exhausts the ephemeral port range.
	var (
		mu            sync.Mutex
		gotBranch     string
		gotProject    string
		branchPresent bool
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		mu.Lock()
		gotBranch = q.Get("branch")
		gotProject = q.Get("project")
		_, branchPresent = q["branch"]
		mu.Unlock()
		_ = json.NewEncoder(w).Encode(sonar.ProjectAnalyses{
			Analyses: []sonar.Analysis{
				{Key: revAnalysisKey, Date: revAnalysisDate, Revision: revRevision},
			},
		})
	}))
	defer server.Close()

	f.Fuzz(func(t *testing.T, branch string) {
		mu.Lock()
		gotBranch, gotProject, branchPresent = "", "", false
		mu.Unlock()

		sonarResults := &sonar.SonarResults{ServerUrl: server.URL, Branch: &sonar.Branch{Name: branch}}
		project := &sonar.Project{Key: revProjectKey}

		if _, err := sonar.GetProjectAnalysisFromRevision(http.DefaultClient, sonarResults, project, revRevision, discardLogger()); err != nil {
			t.Fatalf("branch %q: unexpected error: %v", branch, err)
		}

		mu.Lock()
		sentBranch, sentProject, sawBranch := gotBranch, gotProject, branchPresent
		mu.Unlock()

		if sentProject != revProjectKey {
			t.Errorf("branch %q altered the project parameter: got %q, want %q", branch, sentProject, revProjectKey)
		}
		if branch == "" {
			if sawBranch {
				t.Errorf("empty branch must not be sent at all, got branch=%q", sentBranch)
			}
			return
		}
		if !sawBranch {
			t.Fatalf("branch %q was not sent", branch)
		}
		if sentBranch != branch {
			t.Errorf("branch was not round-tripped: sent %q, SonarQube received %q", branch, sentBranch)
		}
	})
}
