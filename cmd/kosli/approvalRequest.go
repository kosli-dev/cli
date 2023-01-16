package main

import (
	"io"

	"github.com/spf13/cobra"
)

const approvalRequestShortDesc = `Request an approval of a deployment of an artifact in Kosli.`
const approvalRequestLongDesc = approvalRequestShortDesc + `
The request should be reviewed in Kosli UI.` + sha256Desc

const approvalRequestExample = `
# Request that a file type artifact needs approval.
# The approval is for the last 5 git commits
kosli pipeline approval request FILE.tgz \
	--api-token yourAPIToken \
	--artifact-type file \
	--description "An optional description for the requested approval" \
	--newest-commit $(git rev-parse HEAD) \
	--oldest-commit $(git rev-parse HEAD~5) \
	--owner yourOrgName \
	--pipeline yourPipelineName 

# Request and approval for an artifact with a provided fingerprint (sha256).
# The approval is for the last 5 git commits
kosli pipeline approval request \
	--api-token yourAPIToken \
	--description "An optional description for the requested approval" \
	--newest-commit $(git rev-parse HEAD) \
	--oldest-commit $(git rev-parse HEAD~5)	\
	--owner yourOrgName \
	--pipeline yourPipelineName \
	--sha256 yourSha256 
`

func newApprovalRequestCmd(out io.Writer) *cobra.Command {
	o := new(approvalReportOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "request [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   approvalRequestShortDesc,
		Long:    approvalRequestLongDesc,
		Example: approvalRequestExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.payload.ArtifactSha256, false)
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return ValidateRegistryFlags(cmd, o.fingerprintOptions)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args, true)
		},
	}

	cmd.Flags().StringVarP(&o.payload.ArtifactSha256, "sha256", "s", "", sha256Flag)
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", pipelineNameFlag)
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", approvalDescriptionFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", approvalUserDataFlag)
	cmd.Flags().StringVar(&o.oldestSrcCommit, "oldest-commit", "", oldestCommitFlag)
	cmd.Flags().StringVar(&o.newestSrcCommit, "newest-commit", "HEAD", newestCommitFlag)
	cmd.Flags().StringVar(&o.srcRepoRoot, "repo-root", ".", repoRootFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"pipeline", "oldest-commit"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
