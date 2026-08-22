package jira

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeJiraIssueKey(t *testing.T) {
	tests := []struct {
		name        string
		projectKeys []string
		want        string
		matches     []string // strings that should match the pattern
		nonMatches  []string // strings that should not match the pattern
	}{
		{
			name:        "Empty project keys",
			projectKeys: []string{},
			want:        `\b[A-Z][A-Z0-9]{1,9}-[0-9]+`,
			matches: []string{
				"ABC-123",
				"A1-456",
				"XY-789",
			},
			nonMatches: []string{
				"abc-123", // pattern is uppercase-only; FindJiraIssueKeys uppercases the text before applying it
				"A-123",   // project key too short (only 1 char)
				"1A-123",  // project key starts with a number
				"ABC_123", // wrong separator
				"ABC-",    // missing number
				"-123",    // missing project key
			},
		},
		{
			name:        "With project keys",
			projectKeys: []string{"ABC", "XYZ"},
			want:        `\b(ABC|XYZ)-[0-9]+`,
			matches: []string{
				"ABC-123",
				"XYZ-789",
			},
			nonMatches: []string{
				"xyz-123", // pattern is uppercase-only; text is uppercased by FindJiraIssueKeys before matching
				"ABC_123", // wrong separator
				"ABC-",    // missing number
				"-123",    // missing project key
				"DEF-123", // wrong project key
			},
		},
		{
			name:        "With lowercase project keys they are uppercased in the pattern",
			projectKeys: []string{"abc", "xyz"},
			want:        `\b(ABC|XYZ)-[0-9]+`,
			matches: []string{
				"ABC-123",
				"XYZ-789",
			},
			nonMatches: []string{
				"DEF-123",
			},
		},
		{
			// A real Jira project key cannot contain these, and validateJiraProjectKeys
			// rejects them before they reach here: quoting is what makes "the pattern
			// always compiles" a property of the code rather than a promise its callers
			// have to keep.
			name:        "Project keys carrying regex metacharacters are quoted",
			projectKeys: []string{"a(b", "c[d"},
			want:        `\b(A\(B|C\[D)-[0-9]+`,
			matches: []string{
				"A(B-123",
				"C[D-456",
			},
			nonMatches: []string{
				"AB-123",
				"AXB-123",
			},
		},
		{
			// Keys are trimmed and blank ones dropped, so that --jira-project-key "EX, "
			// cannot widen the pattern: an empty alternative matches any -[0-9]+, a
			// whitespace-only key does the same one column over, and an untrimmed " EX"
			// would report " EX-12" as the key.
			name:        "Blank project keys are dropped and the rest trimmed",
			projectKeys: []string{" EX ", "", " ", "\t"},
			want:        `\b(EX)-[0-9]+`,
			matches: []string{
				"EX-12",
			},
			nonMatches: []string{
				"CVE-2026-41284",
				"-41284",
				"fix -123",
				"XEX-12",
			},
		},
		{
			// Not the default pattern: the caller named projects, so widening to every
			// project would answer a question they did not ask. "" means nothing can
			// match, which FindJiraIssueKeys turns into no keys.
			name:        "Project keys that all drop out match nothing rather than everything",
			projectKeys: []string{"", " ", "\t"},
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := makeJiraIssueKeyPattern(tt.projectKeys)
			if got != tt.want {
				t.Errorf("makeJiraIssueKeyPattern() = %v, want %v", got, tt.want)
			}

			// Only test pattern matching if a pattern is returned
			if got != "" {
				re, err := regexp.Compile(got)
				if err != nil {
					t.Errorf("Invalid regex pattern returned: %v", err)
					return
				}

				// Test matches
				for _, s := range tt.matches {
					if !re.MatchString(s) {
						t.Errorf("Pattern %q should match %q but doesn't", got, s)
					}
				}

				// Test non-matches
				for _, s := range tt.nonMatches {
					if re.MatchString(s) {
						t.Errorf("Pattern %q should NOT match %q but does", got, s)
					}
				}
			}
		})
	}
}

func TestFindJiraIssueKeys(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		projectKeys []string
		want        []string
	}{
		{
			name:        "Jira key alongside CVE identifiers",
			text:        "PROJ-42: Upgrade dependency for CVE-2026-41284 / CVE-2026-42498",
			projectKeys: []string{},
			want:        []string{"PROJ-42"},
		},
		{
			name:        "CVE identifier is not matched",
			text:        "fix CVE-2026-41284",
			projectKeys: []string{},
			want:        nil,
		},
		{
			// The empty key must not widen the pattern to any -[0-9]+, which would
			// report -41284 out of the CVE as a key of project EX.
			name:        "an empty project key alongside a real one finds only the real project's keys",
			text:        "EX-12 fixes CVE-2026-41284",
			projectKeys: []string{"EX", ""},
			want:        []string{"EX-12"},
		},
		{
			// And project keys that all drop out find nothing, rather than falling back
			// to every project and returning PROJ-42.
			name:        "project keys that all drop out find nothing",
			text:        "PROJ-42 fixes CVE-2026-41284",
			projectKeys: []string{"", ""},
			want:        nil,
		},
		{
			name:        "multiple CVE identifiers produce no matches",
			text:        "CVE-2026-41284 and CVE-2025-12345",
			projectKeys: []string{},
			want:        nil,
		},
		{
			name:        "CWE-like multi-segment identifier is not matched",
			text:        "addresses CWE-79-1234",
			projectKeys: []string{},
			want:        nil,
		},
		{
			name:        "multiple valid Jira keys",
			text:        "PROJ-1 and PROJ-2 are related",
			projectKeys: []string{},
			want:        []string{"PROJ-1", "PROJ-2"},
		},
		{
			name:        "Jira key at end of string",
			text:        "fix for PROJ-999",
			projectKeys: []string{},
			want:        []string{"PROJ-999"},
		},
		{
			name:        "standalone Jira key",
			text:        "ABC-123",
			projectKeys: []string{},
			want:        []string{"ABC-123"},
		},
		{
			name:        "with project keys filters to specified projects",
			text:        "PROJ-10: fix for OTHER-789",
			projectKeys: []string{"PROJ"},
			want:        []string{"PROJ-10"},
		},
		{
			name:        "with project keys still rejects CVE-like matches",
			text:        "PROJ-10: Upgrade for CVE-2026-41284",
			projectKeys: []string{"PROJ", "CVE"},
			want:        []string{"PROJ-10"},
		},
		{
			name:        "key appears both standalone and in multi-segment identifier",
			text:        "CVE-2026-41284 and also CVE-2026 standalone",
			projectKeys: []string{},
			want:        []string{"CVE-2026"},
		},
		{
			name:        "duplicate keys are deduplicated",
			text:        "PROJ-42 and PROJ-42 again",
			projectKeys: []string{},
			want:        []string{"PROJ-42"},
		},
		{
			name:        "empty text returns nil",
			text:        "",
			projectKeys: []string{},
			want:        nil,
		},
		{
			name:        "lowercase key is matched and returned uppercase",
			text:        "tDIE-12419:Update test9882.ts",
			projectKeys: []string{},
			want:        []string{"TDIE-12419"},
		},
		{
			name:        "lowercase prefix does not produce phantom substring match",
			text:        "tDIE-12419 is the only ticket",
			projectKeys: []string{},
			want:        []string{"TDIE-12419"},
		},
		{
			name:        "lowercase key alongside CVE is matched and CVE is filtered",
			text:        "proj-42 fixes CVE-2026-41284",
			projectKeys: []string{},
			want:        []string{"PROJ-42"},
		},
		{
			name:        "mixed-case duplicate keys are deduplicated",
			text:        "tDIE-12419 and TDIE-12419",
			projectKeys: []string{},
			want:        []string{"TDIE-12419"},
		},
		{
			name:        "CVE-like lowercase occurrence first does not suppress standalone uppercase occurrence",
			text:        "tdie-12419-5 foo TDIE-12419",
			projectKeys: []string{},
			want:        []string{"TDIE-12419"},
		},
		{
			name: "prose word matching Jira key format is treated as a candidate alongside a real key",
			// "note-1" matches the pattern and becomes NOTE-1; if NOTE-1 does not exist in
			// Jira, the attestation will be non-compliant even though PROJ-42 is a valid
			// reference. Use --jira-project-key to restrict matching to known project keys.
			text:        "see note-1 for context, fixes PROJ-42",
			projectKeys: []string{},
			want:        []string{"NOTE-1", "PROJ-42"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindJiraIssueKeys(tt.text, tt.projectKeys)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestGetJiraIssueInfo covers how a lookup is classified. The abort paths are part of the
// contract and are asserted here: a returned error means the attest command stops before
// reporting the attestation, so a case that returns no error today must not start doing so.
func TestGetJiraIssueInfo(t *testing.T) {
	for _, tt := range []struct {
		name string
		// handler serves the issue endpoint. When nil the server is closed before the
		// lookup, so the request fails at the transport level.
		handler    http.HandlerFunc
		wantStatus LookupStatus
		wantExists bool
		wantErr    []string // substrings of the returned error; empty means no error
		wantReason []string // substrings of LookupReason
		wantDebug  string
	}{
		{
			name: "an existing issue is found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"key":"EX-1"}`))
			},
			wantStatus: IssueFound,
			wantExists: true,
			wantDebug:  "status 200",
		},
		{
			name: "a 404 with no evidence of rejected credentials is a missing issue",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-AUSERNAME", "user@example.com")
				w.Header().Set("X-Seraph-LoginReason", "OK")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"errorMessages":["Issue does not exist"]}`))
			},
			wantStatus: IssueMissing,
			wantDebug:  "status 404",
		},
		{
			name: "a 404 handled as anonymous is unverified, not missing",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-AUSERNAME", "anonymous")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"errorMessages":["Issue does not exist"]}`))
			},
			wantStatus: LookupUnverified,
			wantReason: []string{"username user@example.com and API token", "anonymous", "may have expired"},
			wantDebug:  "status 404",
		},
		{
			name: "a 404 with a failed login reason is unverified, not missing",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Seraph-LoginReason", "AUTHENTICATED_FAILED")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"errorMessages":["Issue does not exist"]}`))
			},
			wantStatus: LookupUnverified,
			wantReason: []string{"AUTHENTICATED_FAILED", "may have expired"},
			wantDebug:  "status 404",
		},
		{
			name: "a 401 aborts, naming the base URL and the credentials",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"errorMessages":["Client must be authenticated"]}`))
			},
			wantErr:   []string{"EX-1", "401", "username user@example.com and API token", "may have expired"},
			wantDebug: "status 401",
		},
		{
			name: "a 403 aborts, naming the base URL and the credentials",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
			},
			wantErr:   []string{"403", "may have expired"},
			wantDebug: "status 403",
		},
		{
			name: "a 500 aborts without blaming the credentials",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr:   []string{"500"},
			wantDebug: "status 500",
		},
		{
			name:       "a transport failure is unverified rather than silently missing",
			handler:    nil,
			wantStatus: LookupUnverified,
			wantReason: []string{"no response from Jira at"},
			wantDebug:  "no response",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.handler(w, r)
			}))
			if tt.handler == nil {
				server.Close()
			} else {
				defer server.Close()
			}

			debug := new(bytes.Buffer)
			jc := NewJiraConfig(server.URL, "user@example.com", "token", "")
			result, err := jc.GetJiraIssueInfo("EX-1", "", logger.NewLogger(io.Discard, debug, true))

			require.NotNil(t, result)
			assert.Equal(t, "EX-1", result.IssueID)
			assert.Contains(t, debug.String(), tt.wantDebug)

			// asserted for the abort cases too: a caller that inspects the result after an
			// error must not read it as a found issue
			assert.Equal(t, tt.wantStatus, result.LookupStatus)
			assert.Equal(t, tt.wantExists, result.IssueExists)

			if len(tt.wantErr) > 0 {
				require.Error(t, err)
				for _, want := range tt.wantErr {
					assert.Contains(t, err.Error(), want)
				}
				assert.NotContains(t, err.Error(), "\n", "an error built from a Jira response body must stay on one line")
				return
			}

			require.NoError(t, err)
			for _, want := range tt.wantReason {
				assert.Contains(t, result.LookupReason, want)
			}
			if len(tt.wantReason) == 0 {
				assert.Empty(t, result.LookupReason)
			}
		})
	}
}

// TestGetJiraIssueInfoWithoutLogger covers a caller that passes no logger. The parameter is
// new, so nil is the easy mistake to make on an exported function, and a lookup must report
// its outcome rather than panic.
func TestGetJiraIssueInfoWithoutLogger(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"key":"EX-1"}`))
	}))
	defer server.Close()

	jc := NewJiraConfig(server.URL, "user@example.com", "token", "")
	result, err := jc.GetJiraIssueInfo("EX-1", "", nil)

	require.NoError(t, err)
	assert.Equal(t, IssueFound, result.LookupStatus)
	assert.True(t, result.IssueExists)
}

// TestGetJiraIssueInfoFlattensResponseBody guards the message enrichment against Jira
// answering with a non-JSON body: NewJiraError embeds the whole body in the error it
// builds, and a login page must not end up in a log line.
func TestGetJiraIssueInfoFlattensResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("<html>\n<body>\n" + strings.Repeat("login page ", 200) + "\n</body>\n</html>"))
	}))
	defer server.Close()

	jc := NewJiraConfig(server.URL, "user@example.com", "token", "")
	_, err := jc.GetJiraIssueInfo("EX-1", "", logger.NewLogger(io.Discard, io.Discard, false))

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "\n")
	assert.Less(t, len(err.Error()), 500)
}

func BenchmarkFindJiraIssueKeys(b *testing.B) {
	text := "EX-1 fixes the regression reported in EX-2, see also branch bugfix/EX-3"
	b.Run("default pattern", func(b *testing.B) {
		for b.Loop() {
			FindJiraIssueKeys(text, nil)
		}
	})
	b.Run("project key pattern", func(b *testing.B) {
		for b.Loop() {
			FindJiraIssueKeys(text, []string{"EX"})
		}
	})
}
