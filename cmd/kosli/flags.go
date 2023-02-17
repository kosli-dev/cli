package main

import (
	"github.com/kosli-dev/cli/internal/aws"
	bbUtils "github.com/kosli-dev/cli/internal/bitbucket"
	ghUtils "github.com/kosli-dev/cli/internal/github"
	gitlabUtils "github.com/kosli-dev/cli/internal/gitlab"
	"github.com/spf13/cobra"
)

func addFingerprintFlags(cmd *cobra.Command, o *fingerprintOptions) {
	cmd.Flags().StringVarP(&o.artifactType, "artifact-type", "t", "", artifactTypeFlag)
	cmd.Flags().StringVar(&o.registryProvider, "registry-provider", "", registryProviderFlag)
	cmd.Flags().StringVar(&o.registryUsername, "registry-username", "", registryUsernameFlag)
	cmd.Flags().StringVar(&o.registryPassword, "registry-password", "", registryPasswordFlag)
}

func addAWSAuthFlags(cmd *cobra.Command, o *aws.AWSStaticCreds) {
	cmd.Flags().StringVar(&o.AccessKeyID, "aws-key-id", "", awsKeyIdFlag)
	cmd.Flags().StringVar(&o.SecretAccessKey, "aws-secret-key", "", awsSecretKeyFlag)
	cmd.Flags().StringVar(&o.Region, "aws-region", "", awsRegionFlag)
}

func addDryRunFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&global.DryRun, "dry-run", "D", false, dryRunFlag)
}

func addBitbucketFlags(cmd *cobra.Command, bbConfig *bbUtils.Config, ci string) {
	cmd.Flags().StringVar(&bbConfig.Username, "bitbucket-username", "", bbUsernameFlag)
	cmd.Flags().StringVar(&bbConfig.Password, "bitbucket-password", "", bbPasswordFlag)
	cmd.Flags().StringVar(&bbConfig.Workspace, "bitbucket-workspace", DefaultValue(ci, "workspace"), bbWorkspaceFlag)
	cmd.Flags().StringVar(&bbConfig.Repository, "repository", DefaultValue(ci, "repository"), repositoryFlag)
}

func addGithubFlags(cmd *cobra.Command, githubConfig *ghUtils.GithubConfig, ci string) {
	cmd.Flags().StringVar(&githubConfig.Token, "github-token", "", githubTokenFlag)
	cmd.Flags().StringVar(&githubConfig.Org, "github-org", DefaultValue(ci, "owner"), githubOrgFlag)
	cmd.Flags().StringVar(&githubConfig.Repository, "repository", DefaultValue(ci, "repository"), repositoryFlag)
	cmd.Flags().StringVar(&githubConfig.BaseURL, "github-base-url", "", githubBaseURLFlag)
}

func addGitlabFlags(cmd *cobra.Command, gitlabConfig *gitlabUtils.GitlabConfig, ci string) {
	cmd.Flags().StringVar(&gitlabConfig.Token, "gitlab-token", "", gitlabTokenFlag)
	cmd.Flags().StringVar(&gitlabConfig.Org, "gitlab-org", DefaultValue(ci, "namespace"), gitlabOrgFlag)
	cmd.Flags().StringVar(&gitlabConfig.BaseURL, "gitlab-base-url", "", gitlabBaseURLFlag)
	cmd.Flags().StringVar(&gitlabConfig.Repository, "repository", DefaultValue(ci, "repository"), repositoryFlag)
}

func addArtifactPRFlags(cmd *cobra.Command, o *pullRequestArtifactOptions, ci string, deprecatedFlags bool) {
	if deprecatedFlags {
		cmd.Flags().StringVarP(&o.payload.ArtifactFingerprint, "sha256", "s", "", sha256Flag)
		cmd.Flags().StringVarP(&o.description, "description", "d", "", evidenceDescriptionFlag)
		cmd.Flags().StringVarP(&o.payload.EvidenceName, "evidence-type", "e", "", evidenceTypeFlag)
	}
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), evidenceBuildUrlFlag)
	cmd.Flags().StringVarP(&o.payload.EvidenceName, "name", "n", "", evidenceNameFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", evidenceUserDataFlag)
	cmd.Flags().StringVar(&o.commit, "commit", DefaultValue(ci, "git-commit"), commitPREvidenceFlag)
	cmd.Flags().StringVarP(&o.payload.ArtifactFingerprint, "fingerprint", "F", "", sha256Flag)
	cmd.Flags().StringVarP(&o.pipelineName, "flow", "f", "", pipelineNameFlag)
}

func addCommitPRFlags(cmd *cobra.Command, o *pullRequestCommitOptions, ci string) {
	cmd.Flags().StringVar(&o.payload.CommitSHA, "commit", DefaultValue(ci, "git-commit"), commitPREvidenceFlag)
	cmd.Flags().StringSliceVarP(&o.payload.Pipelines, "pipelines", "p", []string{}, pipelinesFlag)
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), evidenceBuildUrlFlag)
	cmd.Flags().StringVarP(&o.payload.EvidenceName, "name", "n", "", evidenceNameFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", evidenceUserDataFlag)
}

type TypedEvidencePayload struct {
	ArtifactFingerprint string      `json:"artifact_fingerprint,omitempty"`
	CommitSHA           string      `json:"commit_sha,omitempty"`
	EvidenceName        string      `json:"name"`
	BuildUrl            string      `json:"build_url"`
	UserData            interface{} `json:"user_data"`
}
