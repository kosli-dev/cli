package main

import (
	"github.com/kosli-dev/cli/internal/aws"
	azUtils "github.com/kosli-dev/cli/internal/azure"
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
	cmd.Flags().StringSliceVarP(&o.excludePaths, "exclude", "x", []string{}, excludePathsFlag)
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

func addGithubFlags(cmd *cobra.Command, githubFlagsValueHolder *ghUtils.GithubFlagsTempValueHolder, ci string) {
	cmd.Flags().StringVar(&githubFlagsValueHolder.Token, "github-token", "", githubTokenFlag)
	cmd.Flags().StringVar(&githubFlagsValueHolder.Org, "github-org", DefaultValue(ci, "org"), githubOrgFlag)
	cmd.Flags().StringVar(&githubFlagsValueHolder.Repository, "repository", DefaultValue(ci, "repository"), repositoryFlag)
	cmd.Flags().StringVar(&githubFlagsValueHolder.BaseURL, "github-base-url", "", githubBaseURLFlag)
}

func addAzureFlags(cmd *cobra.Command, azureFlagsValueHolder *azUtils.AzureFlagsTempValueHolder, ci string) {
	cmd.Flags().StringVar(&azureFlagsValueHolder.Token, "azure-token", "", azureTokenFlag)
	cmd.Flags().StringVar(&azureFlagsValueHolder.OrgUrl, "azure-org-url", DefaultValue(ci, "org-url"), azureOrgUrlFlag)
	cmd.Flags().StringVar(&azureFlagsValueHolder.Project, "project", DefaultValue(ci, "project"), azureProjectFlag)
	cmd.Flags().StringVar(&azureFlagsValueHolder.Repository, "repository", DefaultValue(ci, "repository"), repositoryFlag)
}

func addGitlabFlags(cmd *cobra.Command, gitlabConfig *gitlabUtils.GitlabConfig, ci string) {
	cmd.Flags().StringVar(&gitlabConfig.Token, "gitlab-token", "", gitlabTokenFlag)
	cmd.Flags().StringVar(&gitlabConfig.Org, "gitlab-org", DefaultValue(ci, "namespace"), gitlabOrgFlag)
	cmd.Flags().StringVar(&gitlabConfig.BaseURL, "gitlab-base-url", "", gitlabBaseURLFlag)
	cmd.Flags().StringVar(&gitlabConfig.Repository, "repository", DefaultValue(ci, "repository"), repositoryFlag)
}

func addArtifactPRFlags(cmd *cobra.Command, o *pullRequestArtifactOptions, ci string) {
	addArtifactEvidenceFlags(cmd, &o.payload.TypedEvidencePayload, ci)
	cmd.Flags().StringVarP(&o.userDataFilePath, "user-data", "u", "", evidenceUserDataFlag)
	cmd.Flags().StringVar(&o.commit, "commit", DefaultValue(ci, "git-commit"), commitPREvidenceFlag)
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().BoolVar(&o.assert, "assert", false, assertPREvidenceFlag)
}

func addArtifactEvidenceFlags(cmd *cobra.Command, payload *TypedEvidencePayload, ci string) {
	addEvidenceFlags(cmd, payload, ci)
	cmd.Flags().StringVarP(&payload.ArtifactFingerprint, "fingerprint", "F", "", fingerprintFlag)
}

func addCommitPRFlags(cmd *cobra.Command, o *pullRequestCommitOptions, ci string) {
	addCommitEvidenceFlags(cmd, &o.payload.TypedEvidencePayload, ci)
	cmd.Flags().StringVarP(&o.userDataFilePath, "user-data", "u", "", evidenceUserDataFlag)
	cmd.Flags().BoolVar(&o.assert, "assert", false, assertPREvidenceFlag)
}

func addCommitEvidenceFlags(cmd *cobra.Command, payload *TypedEvidencePayload, ci string) {
	addEvidenceFlags(cmd, payload, ci)
	cmd.Flags().StringVar(&payload.CommitSHA, "commit", DefaultValue(ci, "git-commit"), commitEvidenceFlag)
	cmd.Flags().StringSliceVarP(&payload.Flows, "flows", "f", []string{}, flowNamesFlag)
}

func addEvidenceFlags(cmd *cobra.Command, payload *TypedEvidencePayload, ci string) {
	cmd.Flags().StringVarP(&payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), evidenceBuildUrlFlag)
	cmd.Flags().StringVarP(&payload.EvidenceName, "name", "n", "", evidenceNameFlag)
	cmd.Flags().StringVar(&payload.EvidenceFingerprint, "evidence-fingerprint", "", evidenceFingerprintFlag)
	cmd.Flags().StringVar(&payload.EvidenceURL, "evidence-url", "", evidenceURLFlag)
}

func addListFlags(cmd *cobra.Command, o *listOptions) {
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().IntVar(&o.pageNumber, "page", 1, pageNumberFlag)
	cmd.Flags().IntVarP(&o.pageLimit, "page-limit", "n", 15, pageLimitFlag)
}

func addAttestationFlags(cmd *cobra.Command, o *CommonAttestationOptions, payload *CommonAttestationPayload, ci string) {
	cmd.Flags().StringVarP(&payload.ArtifactFingerprint, "fingerprint", "F", "", attestationFingerprintFlag)
	cmd.Flags().StringVarP(&o.commitSHA, "commit", "g", DefaultValue(ci, "git-commit"), attestationCommitFlag)
	cmd.Flags().StringVarP(&payload.Url, "url", "b", DefaultValue(ci, "build-url"), attestationUrlFlag)
	cmd.Flags().StringVarP(&o.attestationNameTemplate, "name", "n", "", attestationNameFlag)
	cmd.Flags().StringVar(&payload.EvidenceFingerprint, "evidence-fingerprint", "", evidenceFingerprintFlag)
	cmd.Flags().StringVar(&payload.EvidenceURL, "evidence-url", "", evidenceURLFlag)
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.trailName, "trail", "T", "", trailNameFlag)
	cmd.Flags().StringVarP(&o.userDataFilePath, "user-data", "u", "", attestationUserDataFlag)
	cmd.Flags().StringSliceVarP(&o.evidencePaths, "evidence-paths", "e", []string{}, evidencePathsFlag)
	cmd.Flags().StringVar(&o.srcRepoRoot, "repo-root", ".", attestationRepoRootFlag)

	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)
}
