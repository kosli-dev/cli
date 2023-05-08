package azure

import (
	"context"
	"os"
	"testing"

	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AzureTestSuite struct {
	suite.Suite
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *AzureTestSuite) TestNewAzureClientFromToken() {
	client, err := NewAzureClientFromToken(context.Background(), "some_fake_token", "https://dev.azure.com/kosli_xxxxx")
	require.Error(suite.T(), err)
	require.Nil(suite.T(), client)
}

func (suite *AzureTestSuite) TestPREvidenceForCommit() {
	type result struct {
		wantError   bool
		numberOfPRs int
	}
	for _, t := range []struct {
		name           string
		config         *AzureConfig
		commit         string
		requireEnvVars bool // indicates that a test case needs real credentials from env vars
		result         result
	}{
		{
			name:   "invalid token causes an error",
			commit: "5f61be8f00a01c84e491922a630c9a418c684c7a",
			config: &AzureConfig{
				Token:      "some_fake_token",
				OrgURL:     "https://dev.azure.com/kosli",
				Repository: "cli",
				Project:    "kosli-azure",
			},
			result: result{
				wantError: true,
			},
		},
		{
			name: "can list pull requests for a commit.",
			config: &AzureConfig{
				OrgURL:     "https://dev.azure.com/kosli",
				Repository: "cli",
				Project:    "kosli-azure",
			},
			commit:         "5f61be8f00a01c84e491922a630c9a418c684c7a",
			requireEnvVars: true,
			result: result{
				numberOfPRs: 1,
			},
		},
	} {
		suite.Run(t.name, func() {
			if t.requireEnvVars {
				testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_AZURE_TOKEN"})
				t.config.Token = os.Getenv("KOSLI_AZURE_TOKEN")
			}
			prs, err := t.config.PREvidenceForCommit(t.commit)
			if t.result.wantError {
				require.Errorf(suite.T(), err, "expected an error but got: %s", err)
			} else {
				require.NoErrorf(suite.T(), err, "was NOT expecting error but got: %s", err)
				require.Len(suite.T(), prs, t.result.numberOfPRs)
			}
		})
	}
}

func (suite *AzureTestSuite) TestPullRequestsForCommit() {
	type result struct {
		wantError   bool
		numberOfPRs int
	}
	for _, t := range []struct {
		name       string
		azOrgURL   string
		repository string
		project    string
		commit     string
		result     result
	}{
		{
			name:       "can list pull requests for a commit.",
			azOrgURL:   "https://dev.azure.com/kosli",
			repository: "cli",
			project:    "kosli-azure",
			commit:     "5f61be8f00a01c84e491922a630c9a418c684c7a",
			result: result{
				wantError:   false,
				numberOfPRs: 1,
			},
		},
	} {
		suite.Run(t.name, func() {
			testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_AZURE_TOKEN"})
			token := os.Getenv("KOSLI_AZURE_TOKEN")
			c := &AzureConfig{
				Token:      token,
				OrgURL:     t.azOrgURL,
				Repository: t.repository,
				Project:    t.project,
			}

			prs, err := c.PullRequestsForCommit(t.commit)
			if t.result.wantError {
				require.Errorf(suite.T(), err, "expected an error but got: %s", err)
			} else {
				require.NoErrorf(suite.T(), err, "was NOT expecting error but got: %s", err)
				require.Lenf(suite.T(), prs, t.result.numberOfPRs, "expected %d PRs but got %d", t.result.numberOfPRs, len(prs))
			}
		})
	}
}

func (suite *AzureTestSuite) TestGetPullRequestApprovers() {
	type result struct {
		wantError bool
		approvers []string
	}
	for _, t := range []struct {
		name       string
		azOrgURL   string
		repository string
		project    string
		number     int
		result     result
	}{
		{
			name:       "get an empty list for a PR without approvers",
			azOrgURL:   "https://dev.azure.com/kosli",
			repository: "cli",
			project:    "kosli-azure",
			number:     2,
			result: result{
				approvers: []string{},
			},
		},
		{
			name:       "get the list of approvers for an approved PR",
			azOrgURL:   "https://dev.azure.com/kosli",
			repository: "cli",
			project:    "kosli-azure",
			number:     1,
			result: result{
				approvers: []string{"Ewelina"},
			},
		},
		{
			name:       "non-existing PR causes an error",
			azOrgURL:   "https://dev.azure.com/kosli",
			repository: "cli",
			project:    "kosli-azure",
			number:     666,
			result: result{
				wantError: true,
			},
		},
	} {
		suite.Run(t.name, func() {
			testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_AZURE_TOKEN"})
			token := os.Getenv("KOSLI_AZURE_TOKEN")
			c := &AzureConfig{
				Token:      token,
				OrgURL:     t.azOrgURL,
				Repository: t.repository,
				Project:    t.project,
			}
			approvers, err := c.GetPullRequestApprovers(t.number)
			if t.result.wantError {
				require.Errorf(suite.T(), err, "expected an error but got: %s", err)
			} else {
				require.NoErrorf(suite.T(), err, "was NOT expecting error but got: %s", err)
				require.ElementsMatchf(suite.T(), t.result.approvers, approvers, "want approvers: %v, got approvers: %v",
					t.result.approvers, approvers)
			}
		})
	}
}

func (suite *AzureTestSuite) TestExtractRepoName() {
	for _, t := range []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "full repo name (including org) is separated",
			input: "kosli-dev/cli",
			want:  "cli",
		},
		{
			name:  "short repo name is returned as is",
			input: "cli",
			want:  "cli",
		},
	} {
		suite.Run(t.name, func() {
			repo := extractRepoName(t.input)
			require.Equalf(suite.T(), t.want, repo, "expected %s but got %s", t.want, repo)
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAzureTestSuite(t *testing.T) {
	suite.Run(t, new(AzureTestSuite))
}
