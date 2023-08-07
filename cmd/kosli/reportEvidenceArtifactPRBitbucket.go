package main

import (
	"io"

	bbUtils "github.com/kosli-dev/cli/internal/bitbucket"
	"github.com/spf13/cobra"
)

const reportEvidenceArtifactPRBitbucketShortDesc = `Report a Bitbucket pull request evidence for an artifact in a Kosli flow.  `

const reportEvidenceArtifactPRBitbucketLongDesc = reportEvidenceArtifactPRBitbucketShortDesc + `
It checks if a pull request exists for the artifact (based on its git commit) and reports the pull-request evidence to the artifact in Kosli.  
` + fingerprintDesc

const reportEvidenceArtifactPRBitbucketExample = `
# report a pull request evidence to kosli for a docker image
kosli report evidence artifact pullrequest bitbucket yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--flow yourFlowName \
	--bitbucket-username yourBitbucketUsername \
	--bitbucket-password yourBitbucketPassword \
	--bitbucket-workspace yourBitbucketWorkspace \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--org yourOrgName \
	--api-token yourAPIToken
	
# fail if a pull request does not exist for your artifact
kosli report evidence artifact pullrequest bitbucket yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--flow yourFlowName \
	--bitbucket-username yourBitbucketUsername \
	--bitbucket-password yourBitbucketPassword \
	--bitbucket-workspace yourBitbucketWorkspace \
	--commit yourArtifactGitCommit \
	--repository yourBitbucketGitRepository \
	--org yourOrgName \
	--api-token yourAPIToken \
	--assert
`

func newReportEvidenceArtifactPRBitbucketCmd(out io.Writer) *cobra.Command {
	config := new(bbUtils.Config)
	config.Logger = logger
	config.KosliClient = kosliClient

	o := new(pullRequestArtifactOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	o.retriever = config

	cmd := &cobra.Command{
		Use:     "bitbucket [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Aliases: []string{"bb"},
		Short:   reportEvidenceArtifactPRBitbucketShortDesc,
		Long:    reportEvidenceArtifactPRBitbucketLongDesc,
		Example: reportEvidenceArtifactPRBitbucketExample,
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
			o.retriever.(*bbUtils.Config).Assert = o.assert
			return o.run(out, args)
		},
	}

	ci := WhichCI()
	addBitbucketFlags(cmd, o.getRetriever().(*bbUtils.Config), ci)
	addArtifactPRFlags(cmd, o, ci)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{
		"bitbucket-username", "bitbucket-password",
		"bitbucket-workspace", "commit", "repository", "flow", "name", "build-url",
	})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
