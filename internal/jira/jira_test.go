package jira

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
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
			want:        `[A-Z][A-Z0-9]{1,9}-[0-9]+`,
			matches: []string{
				"ABC-123",
				"A1-456",
				"XY-789",
			},
			nonMatches: []string{
				"abc-123", // project key should start with uppercase
				"A-123",   // project key too short
				"1A-123",  // project key starts with a number
				"ABC_123", // wrong separator
				"ABC-",    // missing number
				"-123",    // missing project key
			},
		},
		{
			name:        "With project keys",
			projectKeys: []string{"ABC", "XYZ"},
			want:        `(ABC|XYZ)-[0-9]+`, // Currently empty in the function implementation
			matches: []string{
				"ABC-123",
				"XYZ-789",
			},
			nonMatches: []string{
				"xyz-123", // project key should start with uppercase
				"ABC_123", // wrong separator
				"ABC-",    // missing number
				"-123",    // missing project key
				"DEF-123", // wrong project key
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MakeJiraIssueKeyPattern(tt.projectKeys)
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
			name:        "empty text returns nil",
			text:        "",
			projectKeys: []string{},
			want:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindJiraIssueKeys(tt.text, tt.projectKeys)
			assert.Equal(t, tt.want, got)
		})
	}
}
