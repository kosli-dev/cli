package gitlab

import (
	"os"
	"testing"

	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type GitlabTestSuite struct {
	suite.Suite
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *GitlabTestSuite) TestNewGitlabClientFromToken() {
	for _, t := range []struct {
		name         string
		gitlabConfig *GitlabConfig
	}{
		{
			name: "when provided a token, a client is created.",
			gitlabConfig: &GitlabConfig{
				Token: "some_fake_token",
			},
		},
	} {
		suite.Run(t.name, func() {
			client, err := t.gitlabConfig.NewGitlabClientFromToken()
			require.NoError(suite.T(), err)
			require.NotNilf(suite.T(), client, "client should not be nil")
		})
	}
}

func (suite *GitlabTestSuite) TestGetClientOptFns() {
	for _, t := range []struct {
		name         string
		gitlabConfig *GitlabConfig
		expectedURL  string
	}{
		{
			name: "when base URL is provided, it is present in the client",
			gitlabConfig: &GitlabConfig{
				Token:   "some_fake_token",
				BaseURL: "https://gitlab.example.com",
			},
			expectedURL: "https://gitlab.example.com/api/v4/",
		},
		{
			name: "when base URL is not provided, the default is the public SaaS gitlab",
			gitlabConfig: &GitlabConfig{
				Token: "some_fake_token",
			},
			expectedURL: "https://gitlab.com/api/v4/",
		},
	} {
		suite.Run(t.name, func() {
			client, err := t.gitlabConfig.NewGitlabClientFromToken()
			require.NoError(suite.T(), err)
			require.NotNilf(suite.T(), client, "client should not be nil")
			require.Equal(suite.T(), t.expectedURL, client.BaseURL().String())
		})
	}
}

func (suite *GitlabTestSuite) TestProjectID() {
	gitlabConfig := &GitlabConfig{
		Org:        "my_org",
		Repository: "test",
	}
	projectID := gitlabConfig.ProjectID()
	require.Equal(suite.T(), "my_org/test", projectID)
}

func (suite *GitlabTestSuite) TestMergeRequestsForCommit() {
	type result struct {
		wantError   bool
		numberOfPRs int
	}
	for _, t := range []struct {
		name           string
		commit         string
		gitlabConfig   *GitlabConfig
		requireEnvVars bool // indicates that a test case needs real credentials from env vars
		result         result
	}{
		{
			name:   "invalid token causes an error",
			commit: "ab4979c426d2d8e77586cfaaf32a7d50a1439bfa",
			gitlabConfig: &GitlabConfig{
				Token: "some_fake_token",
			},
			result: result{
				wantError: true,
			},
		},
		{
			name:   "valid token and commit with an MR find one MR",
			commit: "e6510880aecdc05d79104d937e1adb572bd91911",
			gitlabConfig: &GitlabConfig{
				Org:        "ewelinawilkosz",
				Repository: "merkely-gitlab-demo",
			},
			requireEnvVars: true,
			result: result{
				numberOfPRs: 1,
			},
		},
		{
			name:   "valid token and commit with no MRs find no MRs",
			commit: "2ec23dda01fc85e3f94a2b5ea8cb8cf7e79c4ed6",
			gitlabConfig: &GitlabConfig{
				Org:        "ewelinawilkosz",
				Repository: "merkely-gitlab-demo",
			},
			requireEnvVars: true,
			result: result{
				numberOfPRs: 0,
			},
		},
		{
			name:   "valid token and wrong commit causes an error",
			commit: "ab4979c426d2d8e77586cfaaf32a7d50a1439bfa",
			gitlabConfig: &GitlabConfig{
				Org:        "ewelinawilkosz",
				Repository: "merkely-gitlab-demo",
			},
			requireEnvVars: true,
			result: result{
				wantError: true,
			},
		},
	} {
		suite.Run(t.name, func() {
			if t.requireEnvVars {
				testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_GITLAB_TOKEN"})
				t.gitlabConfig.Token = os.Getenv("KOSLI_GITLAB_TOKEN")
			}
			prs, err := t.gitlabConfig.MergeRequestsForCommit(t.commit)
			if t.result.wantError {
				require.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)
				require.Len(suite.T(), prs, t.result.numberOfPRs)
			}
		})
	}
}

func (suite *GitlabTestSuite) TestPREvidenceForCommit() {
	type result struct {
		wantError   bool
		numberOfPRs int
	}
	for _, t := range []struct {
		name           string
		commit         string
		gitlabConfig   *GitlabConfig
		requireEnvVars bool // indicates that a test case needs real credentials from env vars
		result         result
	}{
		{
			name:   "invalid token causes an error",
			commit: "ab4979c426d2d8e77586cfaaf32a7d50a1439bfa",
			gitlabConfig: &GitlabConfig{
				Token: "some_fake_token",
			},
			result: result{
				wantError: true,
			},
		},
		{
			name:   "valid token and commit with an MR find one MR",
			commit: "e6510880aecdc05d79104d937e1adb572bd91911",
			gitlabConfig: &GitlabConfig{
				Org:        "ewelinawilkosz",
				Repository: "merkely-gitlab-demo",
			},
			requireEnvVars: true,
			result: result{
				numberOfPRs: 1,
			},
		},
		{
			name:   "valid token and commit with no MRs find no MRs",
			commit: "2ec23dda01fc85e3f94a2b5ea8cb8cf7e79c4ed6",
			gitlabConfig: &GitlabConfig{
				Org:        "ewelinawilkosz",
				Repository: "merkely-gitlab-demo",
			},
			requireEnvVars: true,
			result: result{
				numberOfPRs: 0,
			},
		},
		{
			name:   "valid token and wrong commit causes an error",
			commit: "ab4979c426d2d8e77586cfaaf32a7d50a1439bfa",
			gitlabConfig: &GitlabConfig{
				Org:        "ewelinawilkosz",
				Repository: "merkely-gitlab-demo",
			},
			requireEnvVars: true,
			result: result{
				wantError: true,
			},
		},
	} {
		suite.Run(t.name, func() {
			if t.requireEnvVars {
				testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_GITLAB_TOKEN"})
				t.gitlabConfig.Token = os.Getenv("KOSLI_GITLAB_TOKEN")
			}
			prs, err := t.gitlabConfig.PREvidenceForCommit(t.commit)
			if t.result.wantError {
				require.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)
				require.Len(suite.T(), prs, t.result.numberOfPRs)
			}
		})
	}
}

func (suite *GitlabTestSuite) TestGetMergeRequestApprovers() {
	type result struct {
		wantError bool
		approvers []string
	}
	for _, t := range []struct {
		name           string
		mrIID          int
		gitlabConfig   *GitlabConfig
		requireEnvVars bool // indicates that a test case needs real credentials from env vars
		result         result
	}{
		{
			name: "invalid token causes an error",
			gitlabConfig: &GitlabConfig{
				Token: "some_fake_token",
			},
			result: result{
				wantError: true,
			},
		},
		{
			name:  "valid token and mrIID get the correct approvers",
			mrIID: 2,
			gitlabConfig: &GitlabConfig{
				Org:        "ewelinawilkosz",
				Repository: "merkely-gitlab-demo",
			},
			requireEnvVars: true,
			result: result{
				approvers: []string{"Sami Alajrami (@sami.alajrami)"},
			},
		},
		{
			name:  "valid token and mrIID with no approvals returns empty list",
			mrIID: 1,
			gitlabConfig: &GitlabConfig{
				Org:        "ewelinawilkosz",
				Repository: "merkely-gitlab-demo",
			},
			requireEnvVars: true,
			result: result{
				approvers: []string{},
			},
		},
		{
			name:  "valid token and non-existing mrIID causes an error",
			mrIID: 200,
			gitlabConfig: &GitlabConfig{
				Org:        "ewelinawilkosz",
				Repository: "merkely-gitlab-demo",
			},
			requireEnvVars: true,
			result: result{
				wantError: true,
			},
		},
	} {
		suite.Run(t.name, func() {
			if t.requireEnvVars {
				testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_GITLAB_TOKEN"})
				t.gitlabConfig.Token = os.Getenv("KOSLI_GITLAB_TOKEN")
			}
			approvers, err := t.gitlabConfig.GetMergeRequestApprovers(t.mrIID)
			if t.result.wantError {
				require.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)
				require.ElementsMatch(suite.T(), t.result.approvers, approvers)
			}
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGitlabTestSuite(t *testing.T) {
	suite.Run(t, new(GitlabTestSuite))
}
