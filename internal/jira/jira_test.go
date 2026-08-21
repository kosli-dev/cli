package jira

import (
	"regexp"
	"testing"

	"github.com/maxcnunes/httpfake"
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

// TestGetJiraIssueInfo pins which Jira responses mean "the issue does not exist"
// and which mean "we could not find out". Only a 404 is an answer about the issue:
// every other failure has to reach the caller as an error, because reporting it as
// a missing issue makes `attest jira --assert` fail a pipeline for a reason that
// has nothing to do with the commit's Jira references.
func TestGetJiraIssueInfo(t *testing.T) {
	const issueID = "EX-1"

	t.Run("an existing issue is reported as existing", func(t *testing.T) {
		fake := httpfake.New()
		defer fake.Close()
		fake.NewHandler().
			Get("/rest/api/2/issue/" + issueID).
			Reply(200).
			BodyString(`{"id":"10001","key":"EX-1","fields":{"summary":"a summary"}}`)

		jc := NewJiraConfig(fake.Server.URL, "user@example.com", "token", "")
		result, err := jc.GetJiraIssueInfo(issueID, "summary")
		require.NoError(t, err)
		assert.True(t, result.IssueExists)
		assert.Equal(t, issueID, result.IssueID)
	})

	t.Run("a 404 is the one status that means the issue does not exist", func(t *testing.T) {
		fake := httpfake.New()
		defer fake.Close()
		fake.NewHandler().
			Get("/rest/api/2/issue/" + issueID).
			Reply(404).
			BodyString(`{"errorMessages":["Issue does not exist or you do not have permission to see it."],"errors":{}}`)

		jc := NewJiraConfig(fake.Server.URL, "user@example.com", "token", "")
		result, err := jc.GetJiraIssueInfo(issueID, "")
		require.NoError(t, err)
		assert.False(t, result.IssueExists)
	})

	// A stale API token is what this whole classification is for: Jira answers an
	// expired credential with 401, and swallowing it reported every referenced issue
	// as missing until someone recycled the token.
	t.Run("an expired credential is an error, not a missing issue", func(t *testing.T) {
		fake := httpfake.New()
		defer fake.Close()
		fake.NewHandler().
			Get("/rest/api/2/issue/" + issueID).
			Reply(401).
			BodyString(`{"errorMessages":["Client must be authenticated to access this resource."],"errors":{}}`)

		jc := NewJiraConfig(fake.Server.URL, "user@example.com", "token", "")
		result, err := jc.GetJiraIssueInfo(issueID, "")
		require.Error(t, err)
		assert.False(t, result.IssueExists)
		assert.Contains(t, err.Error(), "authenticate")
		// the message has to name the credential to check and where it was used
		assert.Contains(t, err.Error(), "user@example.com")
		assert.Contains(t, err.Error(), fake.Server.URL)
	})

	t.Run("a credential without browse permission is an error, not a missing issue", func(t *testing.T) {
		fake := httpfake.New()
		defer fake.Close()
		fake.NewHandler().
			Get("/rest/api/2/issue/" + issueID).
			Reply(403).
			BodyString(`{"errorMessages":["You do not have the permission to see the specified issue."],"errors":{}}`)

		jc := NewJiraConfig(fake.Server.URL, "user@example.com", "token", "")
		result, err := jc.GetJiraIssueInfo(issueID, "")
		require.Error(t, err)
		assert.False(t, result.IssueExists)
		assert.Contains(t, err.Error(), "permission")
		assert.Contains(t, err.Error(), issueID)
	})

	t.Run("a server error is an error, not a missing issue", func(t *testing.T) {
		fake := httpfake.New()
		defer fake.Close()
		fake.NewHandler().
			Get("/rest/api/2/issue/" + issueID).
			Reply(500).
			BodyString(`{"errorMessages":["Internal server error"],"errors":{}}`)

		jc := NewJiraConfig(fake.Server.URL, "user@example.com", "token", "")
		result, err := jc.GetJiraIssueInfo(issueID, "")
		require.Error(t, err)
		assert.False(t, result.IssueExists)
	})

	// The transport failure is the other half of the same bug: go-jira returns a nil
	// response there, so the status switch never runs and the old guard fell through
	// to IssueExists=false with no error.
	t.Run("an unreachable Jira is an error, not a missing issue", func(t *testing.T) {
		fake := httpfake.New()
		fake.Close() // nothing is listening on this address any more
		unreachableURL := fake.Server.URL

		jc := NewJiraConfig(unreachableURL, "user@example.com", "token", "")
		result, err := jc.GetJiraIssueInfo(issueID, "")
		require.Error(t, err)
		assert.False(t, result.IssueExists)
		assert.Contains(t, err.Error(), unreachableURL)
	})
}

// TestVerifyCredentials pins the check that tells a stale credential apart from a
// genuinely missing issue. Jira answers a request it cannot authenticate with 404 on
// the issue endpoint - it will not confirm that an issue exists to someone who may not
// see it - so the issue lookup alone cannot distinguish the two, and an expired API
// token reads as "every referenced issue is missing". Asking who we are is the question
// that does come back with a straight answer.
func TestVerifyCredentials(t *testing.T) {
	t.Run("a working credential verifies", func(t *testing.T) {
		fake := httpfake.New()
		defer fake.Close()
		fake.NewHandler().
			Get("/rest/api/2/myself").
			Reply(200).
			BodyString(`{"accountId":"1234","emailAddress":"user@example.com","displayName":"A User"}`)

		jc := NewJiraConfig(fake.Server.URL, "user@example.com", "token", "")
		require.NoError(t, jc.VerifyCredentials())
	})

	t.Run("a rejected credential is named in the error", func(t *testing.T) {
		fake := httpfake.New()
		defer fake.Close()
		fake.NewHandler().
			Get("/rest/api/2/myself").
			Reply(401).
			BodyString(`{"errorMessages":["Client must be authenticated to access this resource."],"errors":{}}`)

		jc := NewJiraConfig(fake.Server.URL, "user@example.com", "token", "")
		err := jc.VerifyCredentials()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "authenticate")
		assert.Contains(t, err.Error(), "user@example.com")
		assert.Contains(t, err.Error(), fake.Server.URL)
	})

	// Jira Cloud answers /myself with 404 rather than 401 for some rejected
	// credentials, so a 404 here cannot be waved through as "no such user".
	t.Run("a 404 on the identity endpoint is still a failure to verify", func(t *testing.T) {
		fake := httpfake.New()
		defer fake.Close()
		fake.NewHandler().
			Get("/rest/api/2/myself").
			Reply(404).
			BodyString(`{"errorMessages":["Not found"],"errors":{}}`)

		jc := NewJiraConfig(fake.Server.URL, "user@example.com", "token", "")
		require.Error(t, jc.VerifyCredentials())
	})

	t.Run("an unreachable Jira is a failure to verify", func(t *testing.T) {
		fake := httpfake.New()
		fake.Close() // nothing is listening on this address any more
		jc := NewJiraConfig(fake.Server.URL, "user@example.com", "token", "")
		err := jc.VerifyCredentials()
		require.Error(t, err)
		assert.Contains(t, err.Error(), fake.Server.URL)
	})
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
