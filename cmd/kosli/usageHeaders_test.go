package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/kosli-dev/cli/internal/version"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

var usageHeaderNames = []string{
	"X-Kosli-Client",
	"X-Kosli-Operation",
	"X-Kosli-Invocation-Id",
	"X-Kosli-Cli-Flags",
	"X-Kosli-Ci",
}

var ciEnvVars = []string{
	"BITBUCKET_BUILD_NUMBER", "GITHUB_RUN_NUMBER", "TEAMCITY_VERSION", "GITLAB_CI",
	"TF_BUILD", "CIRCLECI", "CODEBUILD_CI", "JENKINS_URL", "CI",
}

// clearCIEnv unsets the variables WhichCI and ciVendor read, so a test run in
// CI sees the environment it declares.
func clearCIEnv(t *testing.T) {
	t.Helper()
	for _, name := range ciEnvVars {
		t.Setenv(name, "")
		require.NoError(t, os.Unsetenv(name))
	}
}

func TestUsageHeadersDescribeTheCommandOnlyToTheKosliHost(t *testing.T) {
	clearCIEnv(t)
	const (
		apiToken  = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY"
		jiraToken = "jira-secret-7f3a91"
		orgName   = "usage-headers-org"
		flowName  = "flow-from-environment"
	)
	t.Setenv("KOSLI_FLOW", flowName)

	var jiraHeaders, kosliHeaders []http.Header
	jiraServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jiraHeaders = append(jiraHeaders, r.Header.Clone())
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errorMessages":["Issue does not exist"]}`))
	}))
	defer jiraServer.Close()
	kosliServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		kosliHeaders = append(kosliHeaders, r.Header.Clone())
		_, _ = w.Write([]byte(`{}`))
	}))
	defer kosliServer.Close()

	tmpDir := t.TempDir()
	_, workTree, fs, err := testHelpers.InitializeGitRepo(tmpDir)
	require.NoError(t, err)
	commitSha, err := testHelpers.CommitToRepo(workTree, fs, "EX-1 fix the thing")
	require.NoError(t, err)

	_, _, _, _, err = executeCommandC(fmt.Sprintf(`attest jira --name foo --trail trail
		--org %s --host %s --api-token %s --repo-root %s --commit %s
		--jira-base-url %s --jira-username user@example.com --jira-api-token %s`,
		orgName, kosliServer.URL, apiToken, tmpDir, commitSha, jiraServer.URL, jiraToken))
	require.NoError(t, err)

	require.NotEmpty(t, kosliHeaders)
	for _, headers := range kosliHeaders {
		require.Equal(t, "cli/"+version.GetVersion(), headers.Get("X-Kosli-Client"))
		require.Equal(t, "attest jira", headers.Get("X-Kosli-Operation"))
		require.Equal(t, invocationID, headers.Get("X-Kosli-Invocation-Id"))
		require.Equal(t, "commit,flow,jira-api-token,jira-base-url,jira-username,name,repo-root,trail",
			headers.Get("X-Kosli-Cli-Flags"))
		require.Equal(t, "none", headers.Get("X-Kosli-Ci"))
		require.Equal(t, "Kosli/"+version.GetVersion(), headers.Get("User-Agent"))

		for _, name := range usageHeaderNames {
			for _, value := range []string{apiToken, jiraToken, orgName, flowName, kosliServer.URL, jiraServer.URL, commitSha, tmpDir} {
				require.NotContains(t, headers.Get(name), value, name)
			}
		}
	}
	_, err = uuid.Parse(invocationID)
	require.NoError(t, err)

	require.NotEmpty(t, jiraHeaders, "attest jira should have called the Jira server")
	for _, headers := range jiraHeaders {
		for _, name := range usageHeaderNames {
			require.Empty(t, headers.Values(name), name)
		}
	}
}

func TestCIVendor(t *testing.T) {
	for _, tt := range []struct {
		env  map[string]string
		want string
	}{
		{map[string]string{"GITHUB_RUN_NUMBER": "12"}, "github"},
		{map[string]string{"GITLAB_CI": "true"}, "gitlab"},
		{map[string]string{"BITBUCKET_BUILD_NUMBER": "3"}, "bitbucket"},
		{map[string]string{"TF_BUILD": "True"}, "azure-devops"},
		{map[string]string{"CIRCLECI": "true"}, "circleci"},
		{map[string]string{"TEAMCITY_VERSION": "2024.1"}, "teamcity"},
		{map[string]string{"CODEBUILD_CI": "true"}, "codebuild"},
		{map[string]string{"JENKINS_URL": "https://jenkins.example.com/"}, "jenkins"},
		{map[string]string{"GITHUB_RUN_NUMBER": "12", "CI": "true"}, "github"},
		{map[string]string{"CI": "true"}, "other-ci"},
		{map[string]string{"CI": "false"}, "none"},
		{map[string]string{}, "none"},
	} {
		t.Run(fmt.Sprintf("%v", tt.env), func(t *testing.T) {
			clearCIEnv(t)
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			require.Equal(t, tt.want, ciVendor())
		})
	}
}

func TestFlagNamesInEffectExcludeConnectionFlagsAndValues(t *testing.T) {
	cmd := &cobra.Command{Use: "probe"}
	for _, name := range []string{"api-token", "host", "org", "config-file", "http-proxy", "github-token", "debug"} {
		cmd.Flags().String(name, "", "")
	}
	cmd.Flags().String("unset", "", "")
	for _, name := range []string{"api-token", "host", "org", "config-file", "http-proxy", "github-token", "debug"} {
		require.NoError(t, cmd.Flags().Set(name, "value-of-"+name))
	}

	require.Equal(t, "debug,github-token", flagNamesInEffect(cmd))
}

func TestFlagNamesInEffectIsCutAtAFlagBoundary(t *testing.T) {
	cmd := &cobra.Command{Use: "probe"}
	names := map[string]bool{}
	for i := range 60 {
		name := fmt.Sprintf("a-rather-long-flag-name-number-%02d", i)
		names[name] = true
		cmd.Flags().String(name, "", "")
		require.NoError(t, cmd.Flags().Set(name, "x"))
	}

	value := flagNamesInEffect(cmd)

	require.LessOrEqual(t, len(value), maxCliFlagsHeaderBytes)
	require.Greater(t, len(value), maxCliFlagsHeaderBytes-len("a-rather-long-flag-name-number-00,"))
	for _, name := range strings.Split(value, ",") {
		require.True(t, names[name], "%q is not a whole flag name", name)
	}
}
