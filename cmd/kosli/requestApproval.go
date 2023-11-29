package main

import (
	"io"

	"github.com/spf13/cobra"
)

const (
	requestApprovalShortDesc = `Request an approval of a deployment of an artifact to an environment in Kosli.  `
	requestApprovalLongDesc  = requestApprovalShortDesc + `
The request should be reviewed in the Kosli UI.  
` + fingerprintDesc
)

const requestApprovalExample = `
# Request an approval for an artifact with a provided fingerprint (sha256)
# for deployment to environment <yourEnvironmentName>.
# The approval is for all git commits since the last approval to this environment.
kosli request approval \
	--api-token yourAPIToken \
	--description "An optional description for the approval" \
	--environment yourEnvironmentName \
	--org yourOrgName \
	--flow yourFlowName \
	--fingerprint yourArtifactFingerprint

# Request that a file type artifact needs approval for deployment to environment <yourEnvironmentName>.
# The approval is for all git commits since the last approval to this environment.
kosli request approval FILE.tgz \
	--api-token yourAPIToken \
	--artifact-type file \
	--description "An optional description for the requested approval" \
	--environment yourEnvironmentName \
	--newest-commit HEAD \
	--org yourOrgName \
	--flow yourFlowName 

# Request an approval for an artifact with a provided fingerprint (sha256).
# The approval is for all environments.
# The approval is for all commits since the git commit of origin/production branch.
kosli request approval \
	--api-token yourAPIToken \
	--description "An optional description for the requested approval" \
	--newest-commit HEAD \
	--oldest-commit origin/production \
	--org yourOrgName \
	--flow yourFlowName \
	--fingerprint yourArtifactFingerprint
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
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = RequireAtLeastOneOfFlags(cmd, []string{"environment", "oldest-commit"})
			if err != nil {
				return err
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
	cmd.Flags().StringVarP(&o.payload.Environment, "environment", "e", "", approvalEnvironmentNameFlag)
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", approvalDescriptionFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", approvalUserDataFlag)
	cmd.Flags().StringVar(&o.oldestSrcCommit, "oldest-commit", "", oldestCommitFlag)
	cmd.Flags().StringVar(&o.newestSrcCommit, "newest-commit", "HEAD", newestCommitFlag)
	cmd.Flags().StringVar(&o.srcRepoRoot, "repo-root", ".", repoRootFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"flow"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}
