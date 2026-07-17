package main

import (
	"testing"

	"github.com/kosli-dev/cli/internal/gitview"
	"github.com/stretchr/testify/assert"
)

func TestMergeGitRepoInfo(t *testing.T) {
	tests := []struct {
		name             string
		base             *gitview.GitRepoInfo
		repoID           string
		repoName         string
		repoURL          string
		repoProvider     string
		repoNameExplicit bool
		wantNil          bool
		wantID           string
		wantName         string
		wantURL          string
		wantProvider     string
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
			name:             "explicit flag values override base values",
			base:             &gitview.GitRepoInfo{ID: "base-id", Name: "base-name", URL: "https://base.example.com"},
			repoID:           "override-id",
			repoName:         "override-name",
			repoNameExplicit: true,
			wantNil:          false,
			wantID:           "override-id",
			wantName:         "override-name",
			wantURL:          "https://base.example.com",
		},
		{
			name:         "fully-populated base is returned unchanged when no flags are passed",
			base:         &gitview.GitRepoInfo{ID: "53419335", Name: "cyber-dojo/creator", URL: "https://gitlab.com/cyber-dojo/creator", Provider: "gitlab"},
			wantNil:      false,
			wantID:       "53419335",
			wantName:     "cyber-dojo/creator",
			wantURL:      "https://gitlab.com/cyber-dojo/creator",
			wantProvider: "gitlab",
		},
		{
			name:             "CI-detected full-path name is preserved when --repository is not set explicitly",
			base:             &gitview.GitRepoInfo{ID: "53419335", Name: "cyber-dojo/creator", URL: "https://gitlab.com/cyber-dojo/creator"},
			repoID:           "53419335",
			repoName:         "creator", // short CI default from --repository (e.g. GitLab's CI_PROJECT_NAME)
			repoURL:          "https://gitlab.com/cyber-dojo/creator",
			repoNameExplicit: false,
			wantNil:          false,
			wantID:           "53419335",
			wantName:         "cyber-dojo/creator", // full CI_PROJECT_PATH preserved
			wantURL:          "https://gitlab.com/cyber-dojo/creator",
		},
		{
			name:             "explicit --repository overrides CI-detected full-path name",
			base:             &gitview.GitRepoInfo{ID: "53419335", Name: "cyber-dojo/creator", URL: "https://gitlab.com/cyber-dojo/creator"},
			repoID:           "53419335",
			repoName:         "my/custom-name",
			repoURL:          "https://gitlab.com/cyber-dojo/creator",
			repoNameExplicit: true,
			wantNil:          false,
			wantID:           "53419335",
			wantName:         "my/custom-name",
			wantURL:          "https://gitlab.com/cyber-dojo/creator",
		},
		{
			name:             "flag name applied when base has no name even if not explicit",
			repoID:           "flag-id",
			repoName:         "flag-name",
			repoURL:          "https://github.com/org/repo",
			repoNameExplicit: false,
			wantNil:          false,
			wantID:           "flag-id",
			wantName:         "flag-name",
			wantURL:          "https://github.com/org/repo",
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
			result := mergeGitRepoInfo(tt.base, tt.repoID, tt.repoName, tt.repoURL, tt.repoProvider, tt.repoNameExplicit)
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

func TestGetGitRepoInfoFromAzureDevops(t *testing.T) {
	tests := []struct {
		name                    string
		systemCollectionURI     string
		systemTeamProject       string
		buildRepositoryName     string
		buildRepositoryProvider string
		wantName                string
		wantProvider            string
		wantNamespacePath       []string
	}{
		{
			name:                "Azure DevOps Services composes Org/Project/repo",
			systemCollectionURI: "https://dev.azure.com/MyOrg/",
			systemTeamProject:   "Payment",
			buildRepositoryName: "my-repo",
			wantName:            "MyOrg/Payment/my-repo",
			wantProvider:        "azure_devops_services",
			wantNamespacePath:   []string{"MyOrg", "Payment"},
		},
		{
			name:                    "TfsGit provider composes full path and refines provider (regression guard)",
			systemCollectionURI:     "https://dev.azure.com/MyOrg/",
			systemTeamProject:       "Payment",
			buildRepositoryName:     "my-repo",
			buildRepositoryProvider: "TfsGit",
			wantName:                "MyOrg/Payment/my-repo",
			wantProvider:            "azure_devops_services",
			wantNamespacePath:       []string{"MyOrg", "Payment"},
		},
		{
			name:                    "empty provider defaults to TfsGit behaviour (back-compat)",
			systemCollectionURI:     "https://dev.azure.com/MyOrg/",
			systemTeamProject:       "Payment",
			buildRepositoryName:     "my-repo",
			buildRepositoryProvider: "",
			wantName:                "MyOrg/Payment/my-repo",
			wantProvider:            "azure_devops_services",
			wantNamespacePath:       []string{"MyOrg", "Payment"},
		},
		{
			name:                    "GitHub-hosted repo built on Azure Pipelines: bare name, relabeled to github, no namespace",
			systemCollectionURI:     "https://dev.azure.com/MyOrg/",
			systemTeamProject:       "Payment",
			buildRepositoryName:     "my-org/my-repo",
			buildRepositoryProvider: "GitHub",
			wantName:                "my-org/my-repo",
			wantProvider:            "github",
			wantNamespacePath:       nil,
		},
		{
			name:                    "GitHub Enterprise-hosted repo built on Azure Pipelines: bare name, relabeled to github, no namespace",
			systemCollectionURI:     "https://dev.azure.com/MyOrg/",
			systemTeamProject:       "Payment",
			buildRepositoryName:     "my-org/my-repo",
			buildRepositoryProvider: "GitHubEnterprise",
			wantName:                "my-org/my-repo",
			wantProvider:            "github",
			wantNamespacePath:       nil,
		},
		{
			name:                    "Bitbucket-hosted repo built on Azure Pipelines: bare name, relabeled to bitbucket_cloud, no namespace",
			systemCollectionURI:     "https://dev.azure.com/MyOrg/",
			systemTeamProject:       "Payment",
			buildRepositoryName:     "my-repo",
			buildRepositoryProvider: "Bitbucket",
			wantName:                "my-repo",
			wantProvider:            "bitbucket_cloud",
			wantNamespacePath:       nil,
		},
		{
			name:                    "generic external Git repo built on Azure Pipelines: bare name, relabeled to git, no namespace",
			systemCollectionURI:     "https://dev.azure.com/MyOrg/",
			systemTeamProject:       "Payment",
			buildRepositoryName:     "my-repo",
			buildRepositoryProvider: "Git",
			wantName:                "my-repo",
			wantProvider:            "git",
			wantNamespacePath:       nil,
		},
		{
			name:                    "Subversion repo built on Azure Pipelines: bare name, relabeled to subversion, no namespace",
			systemCollectionURI:     "https://dev.azure.com/MyOrg/",
			systemTeamProject:       "Payment",
			buildRepositoryName:     "my-repo",
			buildRepositoryProvider: "Svn",
			wantName:                "my-repo",
			wantProvider:            "subversion",
			wantNamespacePath:       nil,
		},
		{
			name:                    "unmapped provider is treated as an Azure DevOps-hosted repo (refined), no path composition",
			systemCollectionURI:     "https://dev.azure.com/MyOrg/",
			systemTeamProject:       "Payment",
			buildRepositoryName:     "my-repo",
			buildRepositoryProvider: "SomeFutureProvider",
			wantName:                "my-repo",
			wantProvider:            "azure_devops_services",
			wantNamespacePath:       nil,
		},
		{
			name:                    "TFVC repo (not git-based): bare name, refined Azure provider, no namespace",
			systemCollectionURI:     "https://dev.azure.com/MyOrg/",
			systemTeamProject:       "Payment",
			buildRepositoryName:     "my-repo",
			buildRepositoryProvider: "TfsVersionControl",
			wantName:                "my-repo",
			wantProvider:            "azure_devops_services",
			wantNamespacePath:       nil,
		},
		{
			name:                "Azure DevOps Services on a *.visualstudio.com host",
			systemCollectionURI: "https://fabrikam.visualstudio.com/",
			systemTeamProject:   "Payment",
			buildRepositoryName: "my-repo",
			wantName:            "fabrikam/Payment/my-repo",
			wantProvider:        "azure_devops_services",
			wantNamespacePath:   []string{"fabrikam", "Payment"},
		},
		{
			name:                "Azure DevOps Services on a legacy *.vsrm.visualstudio.com release-pipeline host extracts the org, not org.vsrm",
			systemCollectionURI: "https://fabrikam.vsrm.visualstudio.com/",
			systemTeamProject:   "Payment",
			buildRepositoryName: "my-repo",
			wantName:            "fabrikam/Payment/my-repo",
			wantProvider:        "azure_devops_services",
			wantNamespacePath:   []string{"fabrikam", "Payment"},
		},
		{
			name:                "Azure DevOps Services on a vsrm.dev.azure.com release-pipeline host still composes and refines to services",
			systemCollectionURI: "https://vsrm.dev.azure.com/MyOrg/",
			systemTeamProject:   "Payment",
			buildRepositoryName: "my-repo",
			wantName:            "MyOrg/Payment/my-repo",
			wantProvider:        "azure_devops_services",
			wantNamespacePath:   []string{"MyOrg", "Payment"},
		},
		{
			name:                "Azure DevOps Server (on-prem) composes Collection/Project/repo",
			systemCollectionURI: "https://tfs.corp.local/tfs/PRDCollection/",
			systemTeamProject:   "Payment",
			buildRepositoryName: "my-repo",
			wantName:            "PRDCollection/Payment/my-repo",
			wantProvider:        "azure_devops_server",
			wantNamespacePath:   []string{"PRDCollection", "Payment"},
		},
		{
			name:                "collection URI without trailing slash composes the same",
			systemCollectionURI: "https://dev.azure.com/MyOrg",
			systemTeamProject:   "Payment",
			buildRepositoryName: "my-repo",
			wantName:            "MyOrg/Payment/my-repo",
			wantProvider:        "azure_devops_services",
			wantNamespacePath:   []string{"MyOrg", "Payment"},
		},
		{
			name:                "missing SYSTEM_TEAMPROJECT falls back to bare repository name but still refines provider",
			systemCollectionURI: "https://dev.azure.com/MyOrg/",
			systemTeamProject:   "",
			buildRepositoryName: "my-repo",
			wantName:            "my-repo",
			wantProvider:        "azure_devops_services",
			wantNamespacePath:   nil,
		},
		{
			name:                "missing SYSTEM_COLLECTIONURI falls back to bare repository name and coarse provider",
			systemCollectionURI: "",
			systemTeamProject:   "Payment",
			buildRepositoryName: "my-repo",
			wantName:            "my-repo",
			wantProvider:        "azure-devops",
			wantNamespacePath:   nil,
		},
		{
			name:                "unparseable SYSTEM_COLLECTIONURI (no path segment) falls back to bare name but still refines provider",
			systemCollectionURI: "https://dev.azure.com/",
			systemTeamProject:   "Payment",
			buildRepositoryName: "my-repo",
			wantName:            "my-repo",
			wantProvider:        "azure_devops_services",
			wantNamespacePath:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SYSTEM_COLLECTIONURI", tt.systemCollectionURI)
			t.Setenv("SYSTEM_TEAMPROJECT", tt.systemTeamProject)
			t.Setenv("BUILD_REPOSITORY_NAME", tt.buildRepositoryName)
			t.Setenv("BUILD_REPOSITORY_PROVIDER", tt.buildRepositoryProvider)
			t.Setenv("BUILD_REPOSITORY_URI", "https://dev.azure.com/MyOrg/Payment/_git/my-repo")
			t.Setenv("BUILD_REPOSITORY_ID", "repo-id")

			result := getGitRepoInfoFromAzureDevops()

			assert.Equal(t, tt.wantName, result.Name)
			assert.Equal(t, tt.wantProvider, result.Provider)
			assert.Equal(t, tt.wantNamespacePath, result.NamespacePath)
		})
	}
}

// TestAzureRepoProvider guards the helper shared by getGitRepoInfoFromAzureDevops
// and the --repo-provider flag default (DefaultValue): both resolve the provider
// the same way, so the flag default can't clobber the CI-detected value in
// mergeGitRepoInfo (the bug where an Azure-built GitHub repo was mislabelled
// azure-devops).
func TestAzureRepoProvider(t *testing.T) {
	tests := []struct {
		name                    string
		buildRepositoryProvider string
		systemCollectionURI     string
		want                    string
	}{
		{name: "GitHub source maps to github", buildRepositoryProvider: "GitHub", systemCollectionURI: "https://dev.azure.com/MyOrg/", want: "github"},
		{name: "GitHub Enterprise source maps to github", buildRepositoryProvider: "GitHubEnterprise", systemCollectionURI: "https://dev.azure.com/MyOrg/", want: "github"},
		{name: "Bitbucket source maps to bitbucket_cloud", buildRepositoryProvider: "Bitbucket", systemCollectionURI: "https://dev.azure.com/MyOrg/", want: "bitbucket_cloud"},
		{name: "generic Git source maps to git", buildRepositoryProvider: "Git", systemCollectionURI: "https://dev.azure.com/MyOrg/", want: "git"},
		{name: "SVN source maps to subversion", buildRepositoryProvider: "Svn", systemCollectionURI: "https://dev.azure.com/MyOrg/", want: "subversion"},
		{name: "TfsGit refines to services on dev.azure.com", buildRepositoryProvider: "TfsGit", systemCollectionURI: "https://dev.azure.com/MyOrg/", want: "azure_devops_services"},
		{name: "empty provider refines to services on dev.azure.com", buildRepositoryProvider: "", systemCollectionURI: "https://dev.azure.com/MyOrg/", want: "azure_devops_services"},
		{name: "TFVC is Azure-hosted and refines to services", buildRepositoryProvider: "TfsVersionControl", systemCollectionURI: "https://dev.azure.com/MyOrg/", want: "azure_devops_services"},
		{name: "unmapped provider is Azure-hosted and refines to on-prem server", buildRepositoryProvider: "SomeFutureProvider", systemCollectionURI: "https://tfs.corp.local/tfs/PRDCollection/", want: "azure_devops_server"},
		{name: "Azure-hosted degrades to coarse azure-devops without a collection URI", buildRepositoryProvider: "TfsGit", systemCollectionURI: "", want: "azure-devops"},
		{name: "vsrm.dev.azure.com (classic release pipelines) is still Services", buildRepositoryProvider: "TfsGit", systemCollectionURI: "https://vsrm.dev.azure.com/MyOrg/", want: "azure_devops_services"},
		{name: "legacy org.vsrm.visualstudio.com (classic release pipelines) is still Services", buildRepositoryProvider: "TfsGit", systemCollectionURI: "https://fabrikam.vsrm.visualstudio.com/", want: "azure_devops_services"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("BUILD_REPOSITORY_PROVIDER", tt.buildRepositoryProvider)
			t.Setenv("SYSTEM_COLLECTIONURI", tt.systemCollectionURI)
			assert.Equal(t, tt.want, azureRepoProvider(tt.buildRepositoryProvider, parseAzureCollectionURI()))
		})
	}
}

func TestGetGitRepoInfoFromBitbucket(t *testing.T) {
	tests := []struct {
		name                string
		bitbucketProjectKey string
		wantAdditionalInfo  map[string]interface{}
	}{
		{
			name:                "with project key",
			bitbucketProjectKey: "PROJ",
			wantAdditionalInfo:  map[string]interface{}{"project_key": "PROJ"},
		},
		{
			name:                "without project key",
			bitbucketProjectKey: "",
			wantAdditionalInfo:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("BITBUCKET_GIT_HTTP_ORIGIN", "https://bitbucket.org/myteam/my-repo.git")
			t.Setenv("BITBUCKET_REPO_FULL_NAME", "myteam/my-repo")
			t.Setenv("BITBUCKET_REPO_UUID", "repo-uuid")
			t.Setenv("BITBUCKET_PROJECT_KEY", tt.bitbucketProjectKey)

			result := getGitRepoInfoFromBitbucket()

			assert.Equal(t, "bitbucket_cloud", result.Provider)
			assert.Equal(t, []string{"myteam"}, result.NamespacePath)
			assert.Equal(t, tt.wantAdditionalInfo, result.AdditionalInfo)
		})
	}
}

func TestGetGitRepoInfoFromGitLab(t *testing.T) {
	tests := []struct {
		name               string
		ciProjectNamespace string
		wantNamespacePath  []string
	}{
		{
			name:               "top-level group",
			ciProjectNamespace: "my-group",
			wantNamespacePath:  []string{"my-group"},
		},
		{
			name:               "nested subgroup",
			ciProjectNamespace: "my-group/my-subgroup",
			wantNamespacePath:  []string{"my-group", "my-subgroup"},
		},
		{
			name:               "missing namespace",
			ciProjectNamespace: "",
			wantNamespacePath:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CI_PROJECT_URL", "https://gitlab.com/my-group/my-repo")
			t.Setenv("CI_PROJECT_PATH", "my-group/my-repo")
			t.Setenv("CI_PROJECT_ID", "project-id")
			t.Setenv("CI_PROJECT_NAMESPACE", tt.ciProjectNamespace)

			result := getGitRepoInfoFromGitLab()

			assert.Equal(t, "gitlab", result.Provider)
			assert.Equal(t, tt.wantNamespacePath, result.NamespacePath)
		})
	}
}

func TestValidateRepoFlags(t *testing.T) {
	tests := []struct {
		name         string
		repoProvider string
		wantError    bool
	}{
		{name: "empty provider is allowed", repoProvider: ""},
		{name: "github is allowed", repoProvider: "github"},
		{name: "gitlab is allowed", repoProvider: "gitlab"},
		{name: "coarse bitbucket is allowed", repoProvider: "bitbucket"},
		{name: "bitbucket_cloud is allowed", repoProvider: "bitbucket_cloud"},
		{name: "bitbucket_dc is allowed", repoProvider: "bitbucket_dc"},
		{name: "coarse azure-devops is allowed", repoProvider: "azure-devops"},
		{name: "azure_devops_services is allowed", repoProvider: "azure_devops_services"},
		{name: "azure_devops_server is allowed", repoProvider: "azure_devops_server"},
		{name: "git is allowed", repoProvider: "git"},
		{name: "subversion is allowed", repoProvider: "subversion"},
		{name: "unrecognised provider is rejected", repoProvider: "made-up-provider", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRepoFlags("", tt.repoProvider, false)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseAttestationNameTemplate(t *testing.T) {
	tests := []struct {
		name      string
		template  string
		wantP1    string
		wantP2    string
		wantError bool
	}{
		{
			name:     "no dot returns whole string as p1",
			template: "myattestation",
			wantP1:   "myattestation",
			wantP2:   "",
		},
		{
			name:     "dot separates flow and attestation name",
			template: "myflow.myattestation",
			wantP1:   "myflow",
			wantP2:   "myattestation",
		},
		{
			name:      "leading dot is invalid",
			template:  ".myattestation",
			wantError: true,
		},
		{
			name:      "trailing dot is invalid",
			template:  "myflow.",
			wantError: true,
		},
		{
			name:      "multiple dots are invalid",
			template:  "myflow.myattestation.extra",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p1, p2, err := parseAttestationNameTemplate(tt.template)
			if tt.wantError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantP1, p1)
			assert.Equal(t, tt.wantP2, p2)
		})
	}
}
