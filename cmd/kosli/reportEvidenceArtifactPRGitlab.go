package main

import (
	"io"

	gitlabUtils "github.com/kosli-dev/cli/internal/gitlab"

	"github.com/spf13/cobra"
)

const reportEvidenceArtifactPRGitlabShortDesc = `Report a Gitlab merge request evidence for an artifact in a Kosli flow.  `

const reportEvidenceArtifactPRGitlabLongDesc = reportEvidenceArtifactPRGitlabShortDesc + `
It checks if a merge request exists for the artifact (based on its git commit) and reports the merge request evidence to the artifact in Kosli.  
` + fingerprintDesc

const reportEvidenceArtifactPRGitlabExample = `
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
	--org yourOrgName \
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
	--org yourOrgName \
	--api-token yourAPIToken
	
# fail if a merge request does not exist for your artifact
kosli report evidence artifact mergerequest gitlab yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--flow yourFlowName \
	--gitlab-token yourGitlabToken \
	--gitlab-org yourGitlabOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--org yourOrgName \
	--api-token yourAPIToken \
	--assert
`

func newReportEvidenceArtifactPRGitlabCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestArtifactOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	o.retriever = new(gitlabUtils.GitlabConfig)
	cmd := &cobra.Command{
		Use:     "gitlab [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Aliases: []string{"gl"},
		Short:   reportEvidenceArtifactPRGitlabShortDesc,
		Long:    reportEvidenceArtifactPRGitlabLongDesc,
		Example: reportEvidenceArtifactPRGitlabExample,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
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
	addArtifactPRFlags(cmd, o, ci)
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
