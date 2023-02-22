package main

import (
	"io"

	"github.com/spf13/cobra"
)

const requestApprovalShortDesc = `Request an approval of a deployment of an artifact in Kosli.`
const requestApprovalLongDesc = requestApprovalShortDesc + `
The request should be reviewed in Kosli UI.` + fingerprintDesc

const requestApprovalExample = `
# Request that a file type artifact needs approval.
# The approval is for the last 5 git commits
kosli pipeline approval request FILE.tgz \
	--api-token yourAPIToken \
	--artifact-type file \
	--description "An optional description for the requested approval" \
	--newest-commit $(git rev-parse HEAD) \
	--oldest-commit $(git rev-parse HEAD~5) \
	--owner yourOrgName \
	--flow yourPipelineName 

# Request and approval for an artifact with a provided fingerprint (sha256).
# The approval is for the last 5 git commits
kosli pipeline approval request \
	--api-token yourAPIToken \
	--description "An optional description for the requested approval" \
	--newest-commit $(git rev-parse HEAD) \
	--oldest-commit $(git rev-parse HEAD~5)	\
	--owner yourOrgName \
	--flow yourPipelineName \
	--fingerprint yourFingerprint 
`

func newRequestApprovalCmd(out io.Writer) *cobra.Command {
	o := new(reportApprovalOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "approval [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   requestApprovalShortDesc,
		Long:    requestApprovalLongDesc,
		Example: requestApprovalExample,
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
			return o.run(args, true)
		},
	}

	cmd.Flags().StringVarP(&o.payload.ArtifactFingerprint, "fingerprint", "F", "", fingerprintFlag)
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", pipelineNameFlag)
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", approvalDescriptionFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", approvalUserDataFlag)
	cmd.Flags().StringVar(&o.oldestSrcCommit, "oldest-commit", "", oldestCommitFlag)
	cmd.Flags().StringVar(&o.newestSrcCommit, "newest-commit", "HEAD", newestCommitFlag)
	cmd.Flags().StringVar(&o.srcRepoRoot, "repo-root", ".", repoRootFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"flow", "oldest-commit"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
