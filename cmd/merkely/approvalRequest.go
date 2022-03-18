package main

import (
	"io"

	"github.com/spf13/cobra"
)

const approvalRequestDesc = `
Request an approval of a deployment of an artifact in Merkely. The request should be reviewed in Merkely UI. 
` + sha256Desc

const approvalRequestExample = `
# Request that a file artifact needs approval.
# The approval is for the last 5 git commits
merkely pipeline approval request FILE.tgz \
	--api-token yourAPIToken \
	--owner yourOrgName \
	--pipeline yourPipelineName \
	--artifact-type file \
	--description "An optional description for the requested approval" \
	--newest-commit $(git rev-parse HEAD) \
	--oldest-commit $(git rev-parse HEAD~5)

# Request that an artifact with a sha256 needs approval.
# The approval is for the last 5 git commits
merkely pipeline approval request \
	--api-token yourAPIToken \
	--owner yourOrgName \
	--pipeline yourPipelineName \
	--sha256 yourSha256 \
	--description "An optional description for the requested approval" \
	--newest-commit $(git rev-parse HEAD) \
	--oldest-commit $(git rev-parse HEAD~5)	
`

func newApprovalRequestCmd(out io.Writer) *cobra.Command {
	o := new(approvalReportOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "request [ARTIFACT-NAME-OR-PATH]",
		Short:   "Request an approval for deploying an artifact in Merkely. ",
		Long:    approvalRequestDesc,
		Example: approvalRequestExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorAfterPrintingHelp(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.payload.ArtifactSha256, false)
			if err != nil {
				return ErrorAfterPrintingHelp(cmd, err.Error())
			}
			return ValidateRegisteryFlags(cmd, o.fingerprintOptions)
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

	err := RequireFlags(cmd, []string{"pipeline", "oldest-commit"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}
