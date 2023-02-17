package main

import (
	"io"

	gitlabUtils "github.com/kosli-dev/cli/internal/gitlab"

	"github.com/spf13/cobra"
)

const pullRequestEvidenceGitlabShortDesc = `Report a Gitlab merge request evidence for an artifact in a Kosli flow.`

const pullRequestEvidenceGitlabLongDesc = pullRequestEvidenceGitlabShortDesc + `
It checks if a merge request exists for the artifact (based on its git commit) and report the merge request evidence to the artifact in Kosli. 
` + sha256Desc

const pullRequestEvidenceGitlabExample = `
# report a merge request evidence to kosli for a docker image
kosli report evidence artifact mergerequest gitlab yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--flow yourFlowName \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--owner yourOrgName \
	--api-token yourAPIToken

# report a merge request evidence (from an on-prem Gitlab) to kosli for a docker image 
kosli report evidence artifact mergerequest gitlab yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--flow yourFlowName \
	--gitlab-base-url https://gitlab.example.org \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--owner yourOrgName \
	--api-token yourAPIToken
	
# fail if a merge request does not exist for your artifact
kosli report evidence artifact mergerequest gitlab yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--flow yourFlowName \
	--pipeline yourPipelineName \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--owner yourOrgName \
	--api-token yourAPIToken \
	--assert
`

func newPullRequestEvidenceGitlabCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestArtifactOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	o.retriever = new(gitlabUtils.GitlabConfig)
	cmd := &cobra.Command{
		Use:     "gitlab [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Aliases: []string{"gl"},
		Short:   pullRequestEvidenceGitlabShortDesc,
		Long:    pullRequestEvidenceGitlabLongDesc,
		Example: pullRequestEvidenceGitlabExample,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.payload.ArtifactFingerprint, false)
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return ValidateRegistryFlags(cmd, o.fingerprintOptions)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	ci := WhichCI()
	addGitlabFlags(cmd, o.getRetriever().(*gitlabUtils.GitlabConfig), ci)
	addArtifactPRFlags(cmd, o, ci, false)
	cmd.Flags().BoolVar(&o.assert, "assert", false, assertPREvidenceFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{
		"gitlab-token", "gitlab-org", "commit", "name",
		"repository", "flow", "build-url",
	})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
