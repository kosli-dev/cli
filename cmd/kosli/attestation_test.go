package main

import (
	"errors"
	"fmt"
	"testing"

	kosliErrors "github.com/kosli-dev/cli/internal/errors"
	"github.com/kosli-dev/cli/internal/gitview"
	"github.com/stretchr/testify/assert"
)

func TestMergeGitRepoInfo(t *testing.T) {
	tests := []struct {
		name         string
		base         *gitview.GitRepoInfo
		repoID       string
		repoName     string
		repoURL      string
		repoProvider string
		wantNil      bool
		wantID       string
		wantName     string
		wantURL      string
		wantProvider string
	}{
		{
			name:    "nil when both ID and Name are empty",
			wantNil: true,
		},
		{
			name:    "nil when ID is provided but Name is empty",
			repoID:  "repo-id",
			wantNil: true,
		},
		{
			name:     "nil when Name is provided but ID is empty",
			repoName: "repo-name",
			wantNil:  true,
		},
		{
			name:     "nil when ID and Name are provided but URL is empty",
			repoID:   "repo-id",
			repoName: "repo-name",
			wantNil:  true,
		},
		{
			name:     "non-nil when ID, Name, and URL are all provided",
			repoID:   "repo-id",
			repoName: "repo-name",
			repoURL:  "https://github.com/org/repo",
			wantNil:  false,
			wantID:   "repo-id",
			wantName: "repo-name",
			wantURL:  "https://github.com/org/repo",
		},
		{
			name:         "includes URL and Provider when both are provided alongside ID and Name",
			repoID:       "repo-id",
			repoName:     "repo-name",
			repoURL:      "https://github.com/org/repo",
			repoProvider: "github",
			wantNil:      false,
			wantID:       "repo-id",
			wantName:     "repo-name",
			wantURL:      "https://github.com/org/repo",
			wantProvider: "github",
		},
		{
			name:     "flag values override base values",
			base:     &gitview.GitRepoInfo{ID: "base-id", Name: "base-name", URL: "https://base.example.com"},
			repoID:   "override-id",
			repoName: "override-name",
			wantNil:  false,
			wantID:   "override-id",
			wantName: "override-name",
			wantURL:  "https://base.example.com",
		},
		{
			name:    "nil when base has ID but no Name and no flags",
			base:    &gitview.GitRepoInfo{ID: "base-id"},
			wantNil: true,
		},
		{
			name:    "nil when base has Name but no ID and no flags",
			base:    &gitview.GitRepoInfo{Name: "base-name"},
			wantNil: true,
		},
		{
			name:    "nil when base provides Name and flag provides ID but URL is missing",
			base:    &gitview.GitRepoInfo{Name: "base-name"},
			repoID:  "flag-id",
			wantNil: true,
		},
		{
			name:     "non-nil when base provides Name and flags provide ID and URL",
			base:     &gitview.GitRepoInfo{Name: "base-name"},
			repoID:   "flag-id",
			repoURL:  "https://github.com/org/repo",
			wantNil:  false,
			wantID:   "flag-id",
			wantName: "base-name",
			wantURL:  "https://github.com/org/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeGitRepoInfo(tt.base, tt.repoID, tt.repoName, tt.repoURL, tt.repoProvider)
			if tt.wantNil {
				assert.Nil(t, result)
				return
			}
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantID, result.ID)
			assert.Equal(t, tt.wantName, result.Name)
			assert.Equal(t, tt.wantURL, result.URL)
			assert.Equal(t, tt.wantProvider, result.Provider)
		})
	}
}

func TestWrapAttestationError(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		assert.NoError(t, wrapAttestationError(nil))
	})

	t.Run("ErrCompliance is preserved through wrapAttestationError", func(t *testing.T) {
		inner := kosliErrors.NewErrCompliance("assert failed: no pull request found")
		err := wrapAttestationError(inner)
		var ec *kosliErrors.ErrCompliance
		assert.True(t, errors.As(err, &ec), "expected ErrCompliance to be preserved, got %T: %v", err, err)
	})

	t.Run("ErrServer is preserved through wrapAttestationError", func(t *testing.T) {
		inner := kosliErrors.NewErrServer("server error")
		err := wrapAttestationError(inner)
		var es *kosliErrors.ErrServer
		assert.True(t, errors.As(err, &es), "expected ErrServer to be preserved, got %T: %v", err, err)
	})

	t.Run("API message requiring rewrite loses type but message is improved", func(t *testing.T) {
		raw := fmt.Errorf("requires at least one of: artifact_fingerprint or git_commit_info. some detail")
		err := wrapAttestationError(raw)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requires at least one of: specifying the fingerprint")
	})
}
