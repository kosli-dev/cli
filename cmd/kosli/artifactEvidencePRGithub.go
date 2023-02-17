package main

import (
	"io"

	ghUtils "github.com/kosli-dev/cli/internal/github"

	"github.com/spf13/cobra"
)

const pullRequestEvidenceGithubShortDesc = `Report a Github pull request evidence for an artifact in a Kosli pipeline.`

const pullRequestEvidenceGithubLongDesc = pullRequestEvidenceGithubShortDesc + `
It checks if a pull request exists for the artifact (based on its git commit) and report the pull-request evidence to the artifact in Kosli. 
` + sha256Desc

const pullRequestEvidenceGithubExample = `
# report a pull request evidence to kosli for a docker image
kosli pipeline artifact report evidence github-pullrequest yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--pipeline yourPipelineName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--owner yourOrgName \
	--api-token yourAPIToken
	
# fail if a pull request does not exist for your artifact
kosli pipeline artifact report evidence github-pullrequest yourDockerImageName \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--pipeline yourPipelineName \
	--github-token yourGithubToken \
	--github-org yourGithubOrg \
	--commit yourArtifactGitCommit \
	--repository yourGithubGitRepository \
	--owner yourOrgName \
	--api-token yourAPIToken \
	--assert
`

func newPullRequestEvidenceGithubCmd(out io.Writer) *cobra.Command {
	o := new(pullRequestArtifactOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	o.retriever = new(ghUtils.GithubConfig)
	cmd := &cobra.Command{
		Use:     "github-pullrequest [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Aliases: []string{"gh-pr", "github-pr"},
		Short:   pullRequestEvidenceGithubShortDesc,
		Long:    pullRequestEvidenceGithubLongDesc,
		Example: pullRequestEvidenceGithubExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.payload.ArtifactFingerprint, false)
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			err = MuXRequiredFlags(cmd, []string{"name", "evidence-type"}, true)
			if err != nil {
				return err
			}

			err = MuXRequiredFlags(cmd, []string{"sha256", "fingerprint"}, false)
			if err != nil {
				return err
			}
			return ValidateRegistryFlags(cmd, o.fingerprintOptions)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	ci := WhichCI()
	addGithubFlags(cmd, o.getRetriever().(*ghUtils.GithubConfig), ci)
	addArtifactPRFlags(cmd, o, ci, true)
	cmd.Flags().BoolVar(&o.assert, "assert", false, assertPREvidenceFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := DeprecateFlags(cmd, map[string]string{
		"evidence-type": "use --name instead",
		"description":   "description is no longer used",
		"sha256":        "use --fingerprint instead",
	})
	if err != nil {
		logger.Error("failed to configure deprecated flags: %v", err)
	}

	err = RequireFlags(cmd, []string{
		"github-token", "github-org", "commit",
		"repository", "pipeline", "build-url",
	})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
