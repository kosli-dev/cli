package main

import (
	"log"

	"github.com/kosli-dev/cli/internal/aws"
	azUtils "github.com/kosli-dev/cli/internal/azure"
	bbUtils "github.com/kosli-dev/cli/internal/bitbucket"
	ghUtils "github.com/kosli-dev/cli/internal/github"
	gitlabUtils "github.com/kosli-dev/cli/internal/gitlab"
	"github.com/spf13/cobra"
)

// allowed commit redaction values
var allowedCommitRedactionValues = map[string]struct{}{
	"author":  {},
	"message": {},
	"branch":  {},
}

func addFingerprintFlags(cmd *cobra.Command, o *fingerprintOptions) {
	cmd.Flags().StringVarP(&o.artifactType, "artifact-type", "t", "", artifactTypeFlag)
	cmd.Flags().StringVar(&o.registryProvider, "registry-provider", "", registryProviderFlag)
	cmd.Flags().StringVar(&o.registryUsername, "registry-username", "", registryUsernameFlag)
	cmd.Flags().StringVar(&o.registryPassword, "registry-password", "", registryPasswordFlag)
	cmd.Flags().StringSliceVarP(&o.excludePaths, "exclude", "x", []string{}, excludePathsFlag)

	err := DeprecateFlags(cmd, map[string]string{
		"registry-provider": "no longer used",
	})

	if err != nil {
		log.Fatalf("failed to configure deprecated flags: %v", err)
	}
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
	cmd.Flags().StringVar(&bbConfig.AccessToken, "bitbucket-access-token", "", bbAccessTokenFlag)
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

func addListFlags(cmd *cobra.Command, o *listOptions, customPageLimit ...int) {
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().IntVar(&o.pageNumber, "page", 1, pageNumberFlag)

	// Use customPageLimit if provided, otherwise default to 15
	pageLimit := 15
	if len(customPageLimit) > 0 {
		pageLimit = customPageLimit[0]
	}

	cmd.Flags().IntVarP(&o.pageLimit, "page-limit", "n", pageLimit, pageLimitFlag)
}

func addAttestationFlags(cmd *cobra.Command, o *CommonAttestationOptions, payload *CommonAttestationPayload, ci string) {
	commitFlagDesc := attestationCommitFlag
	if _, ok := cmd.Annotations["pr"]; ok {
		commitFlagDesc = "the git merge commit to be checked for associated pull requests."
	}
	cmd.Flags().StringVarP(&payload.ArtifactFingerprint, "fingerprint", "F", "", attestationFingerprintFlag)
	cmd.Flags().StringVarP(&o.commitSHA, "commit", "g", DefaultValueForCommit(ci, false), commitFlagDesc)
	cmd.Flags().StringSliceVar(&o.redactedCommitInfo, "redact-commit-info", []string{}, attestationRedactCommitInfoFlag)
	cmd.Flags().StringVarP(&payload.OriginURL, "origin-url", "o", DefaultValue(ci, "build-url"), attestationOriginUrlFlag)
	cmd.Flags().StringVarP(&o.attestationNameTemplate, "name", "n", "", attestationNameFlag)
	cmd.Flags().StringToStringVar(&o.externalFingerprints, "external-fingerprint", map[string]string{}, externalFingerprintFlag)
	cmd.Flags().StringToStringVar(&o.externalURLs, "external-url", map[string]string{}, externalURLFlag)
	cmd.Flags().StringToStringVar(&o.annotations, "annotate", map[string]string{}, annotationFlag)
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.trailName, "trail", "T", "", trailNameFlag)
	cmd.Flags().StringVarP(&o.userDataFilePath, "user-data", "u", "", attestationUserDataFlag)
	cmd.Flags().StringSliceVar(&o.attachments, "attachments", []string{}, attachmentsFlag)
	cmd.Flags().StringVar(&o.srcRepoRoot, "repo-root", ".", attestationRepoRootFlag)
	cmd.Flags().StringVar(&payload.Description, "description", "", attestationDescription)

	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)
}
