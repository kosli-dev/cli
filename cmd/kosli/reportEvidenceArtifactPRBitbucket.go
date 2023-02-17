package main

import (
	"io"

	bbUtils "github.com/kosli-dev/cli/internal/bitbucket"
	"github.com/spf13/cobra"
)

const pullRequestEvidenceBitbucketShortDesc = `Report a Bitbucket pull request evidence for an artifact in a Kosli flow.`

const pullRequestEvidenceBitbucketLongDesc = pullRequestEvidenceBitbucketShortDesc + `
It checks if a pull request exists for the artifact (based on its git commit) and report the pull-request evidence to the artifact in Kosli. 
` + sha256Desc

const pullRequestEvidenceBitbucketExample = `
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
	--owner yourOrgName \
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
	--owner yourOrgName \
	--api-token yourAPIToken \
	--assert
`

func newPullRequestEvidenceBitbucketCmd(out io.Writer) *cobra.Command {
	config := new(bbUtils.Config)
	config.Logger = logger
	config.KosliClient = kosliClient

	o := new(pullRequestArtifactOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	o.retriever = config

	cmd := &cobra.Command{
		Use:     "bitbucket [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Aliases: []string{"bb"},
		Short:   pullRequestEvidenceBitbucketShortDesc,
		Long:    pullRequestEvidenceBitbucketLongDesc,
		Example: pullRequestEvidenceBitbucketExample,
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
	addBitbucketFlags(cmd, o.getRetriever().(*bbUtils.Config), ci)
	addArtifactPRFlags(cmd, o, ci, true)
	cmd.Flags().BoolVar(&o.getRetriever().(*bbUtils.Config).Assert, "assert", false, assertPREvidenceFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"bitbucket-username", "bitbucket-password",
		"bitbucket-workspace", "commit", "repository", "flow", "name", "build-url"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
