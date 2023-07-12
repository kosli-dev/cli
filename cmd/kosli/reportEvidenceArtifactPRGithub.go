package main

import (
	"io"

	ghUtils "github.com/kosli-dev/cli/internal/github"

	"github.com/spf13/cobra"
)

const reportEvidenceArtifactPRGithubShortDesc = `Report a Github pull request evidence for an artifact in a Kosli flow.  `

const reportEvidenceArtifactPRGithubLongDesc = reportEvidenceArtifactPRGithubShortDesc + `
It checks if a pull request exists for the artifact (based on its git commit) and reports the pull-request evidence to the artifact in Kosli.  
` + fingerprintDesc

const reportEvidenceArtifactPRGithubExample = `
# report a pull request evidence to kosli for a docker image
kosli report evidence artifact pullrequest github yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--flow yourFlowName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--org yourOrgName \
	--api-token yourAPIToken
	
# fail if a pull request does not exist for your artifact
kosli report evidence artifact pullrequest github yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--flow yourFlowName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--org yourOrgName \
	--api-token yourAPIToken \
	--assert
`

func newReportEvidenceArtifactPRGithubCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestArtifactOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	githubFlagsValues := new(ghUtils.GithubFlagsTempValueHolder)
	cmd := &cobra.Command{
		Use:     "github [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Aliases: []string{"gh"},
		Short:   reportEvidenceArtifactPRGithubShortDesc,
		Long:    reportEvidenceArtifactPRGithubLongDesc,
		Example: reportEvidenceArtifactPRGithubExample,
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
			o.retriever = ghUtils.NewGithubConfig(githubFlagsValues.Token, githubFlagsValues.BaseURL,
				githubFlagsValues.Org, githubFlagsValues.Repository)
			return o.run(out, args)
		},
	}

	ci := WhichCI()
	addGithubFlags(cmd, githubFlagsValues, ci)
	addArtifactPRFlags(cmd, o, ci)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{
		"github-token", "github-org", "commit",
		"repository", "flow", "build-url", "name",
	})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
