package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/gitview"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const reportApprovalShortDesc = `Report an approval of deploying an artifact to Kosli.`
const reportApprovalLongDesc = reportApprovalShortDesc + `
` + fingerprintDesc

const reportApprovalExample = `
# Report that a file type artifact has been approved for deployment.
# The approval is for the last 5 git commits
kosli report approval FILE.tgz \
	--api-token yourAPIToken \
	--artifact-type file \
	--description "An optional description for the approval" \
	--newest-commit $(git rev-parse HEAD) \
	--oldest-commit $(git rev-parse HEAD~5) \
	--org yourOrgName \
	--flow yourFlowName 

# Report that an artifact with a provided fingerprint (sha256) has been approved for deployment.
# The approval is for the last 5 git commits
kosli report approval \
	--api-token yourAPIToken \
	--description "An optional description for the approval" \
	--newest-commit $(git rev-parse HEAD) \
	--oldest-commit $(git rev-parse HEAD~5) \
	--org yourOrgName \
	--flow yourFlowName \
	--fingerprint yourArtifactFingerprint
`

type reportApprovalOptions struct {
	fingerprintOptions *fingerprintOptions
	flowName           string
	oldestSrcCommit    string
	newestSrcCommit    string
	srcRepoRoot        string
	userDataFile       string
	payload            ApprovalPayload
}

type ApprovalPayload struct {
	ArtifactFingerprint string              `json:"artifact_sha256"`
	Description         string              `json:"description"`
	CommitList          []string            `json:"src_commit_list"`
	Reviews             []map[string]string `json:"approvals"`
	UserData            interface{}         `json:"user_data"`
}

func newReportApprovalCmd(out io.Writer) *cobra.Command {
	o := new(reportApprovalOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "approval [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   reportApprovalShortDesc,
		Long:    reportApprovalLongDesc,
		Example: reportApprovalExample,
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
			return o.run(args, false)
		},
	}

	cmd.Flags().StringVarP(&o.payload.ArtifactFingerprint, "fingerprint", "F", "", fingerprintFlag)
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
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

func (o *reportApprovalOptions) run(args []string, request bool) error {
	var err error
	o.payload.ArtifactFingerprint, err = o.payloadArtifactSHA256(args)
	if err != nil {
		return err
	}

	o.payload.Reviews = o.payloadReviews(request)

	o.payload.UserData, err = LoadJsonData(o.userDataFile)
	if err != nil {
		return err
	}

	o.payload.CommitList, err = o.payloadCommitList()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/approvals/", global.Host, global.Org, o.flowName)

	reqParams := &requests.RequestParams{
		Method:   http.MethodPost,
		URL:      url,
		Payload:  o.payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("approval created for artifact: %s", o.payload.ArtifactFingerprint)
	}
	return err
}

func (o *reportApprovalOptions) payloadArtifactSHA256(args []string) (string, error) {
	if o.payload.ArtifactFingerprint == "" {
		sha256, err := GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return sha256, err
		}
		return sha256, nil
	}
	return o.payload.ArtifactFingerprint, nil
}

func (o *reportApprovalOptions) payloadReviews(request bool) []map[string]string {
	if !request {
		return []map[string]string{
			{
				"state":        "APPROVED",
				"comment":      o.payload.Description,
				"approved_by":  "External",
				"approval_url": "undefined",
			},
		}
	} else {
		return []map[string]string{}
	}
}

func (o *reportApprovalOptions) payloadCommitList() ([]string, error) {
	commits, err := o.commitsHistory()
	if err != nil {
		return nil, err
	}

	// Need this line to make sure an empty list is converted to [] and not null in SendPayload
	commitList := make([]string, 0)
	for _, commit := range commits {
		commitList = append(commitList, commit.Sha1)
	}
	return commitList, nil
}

func (o *reportApprovalOptions) commitsHistory() ([]*gitview.CommitInfo, error) {
	gitView, err := gitview.New(o.srcRepoRoot)
	if err != nil {
		return nil, err
	}

	commits, err := gitView.CommitsBetween(o.oldestSrcCommit, o.newestSrcCommit, logger)
	if err != nil {
		return nil, err
	}
	return commits, nil
}
