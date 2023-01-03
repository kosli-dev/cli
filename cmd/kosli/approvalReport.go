package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/gitview"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const approvalReportShortDesc = `Report an approval of deploying an artifact to Kosli.`
const approvalReportLongDesc = approvalReportShortDesc + `
` + sha256Desc

const approvalReportExample = `
# Report that a file type artifact has been approved for deployment.
# The approval is for the last 5 git commits
kosli pipeline approval report FILE.tgz \
	--api-token yourAPIToken \
	--artifact-type file \
	--description "An optional description for the approval" \
	--newest-commit $(git rev-parse HEAD) \
	--oldest-commit $(git rev-parse HEAD~5) \
	--owner yourOrgName \
	--pipeline yourPipelineName 

# Report that an artifact with a provided fingerprint (sha256) has been approved for deployment.
# The approval is for the last 5 git commits
kosli pipeline approval report \
	--api-token yourAPIToken \
	--description "An optional description for the approval" \
	--newest-commit $(git rev-parse HEAD) \
	--oldest-commit $(git rev-parse HEAD~5) \
	--owner yourOrgName \
	--pipeline yourPipelineName \
	--sha256 yourSha256
`

type approvalReportOptions struct {
	fingerprintOptions *fingerprintOptions
	pipelineName       string
	oldestSrcCommit    string
	newestSrcCommit    string
	srcRepoRoot        string
	userDataFile       string
	payload            ApprovalPayload
}

type ApprovalPayload struct {
	ArtifactSha256 string              `json:"artifact_sha256"`
	Description    string              `json:"description"`
	CommitList     []string            `json:"src_commit_list"`
	Reviews        []map[string]string `json:"approvals"`
	UserData       interface{}         `json:"user_data"`
}

//goland:noinspection GoUnusedParameter
func newApprovalReportCmd(out io.Writer) *cobra.Command {
	o := new(approvalReportOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "report [ARTIFACT-NAME-OR-PATH]",
		Short:   approvalReportShortDesc,
		Long:    approvalReportLongDesc,
		Example: approvalReportExample,
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
			return o.run(args, false)
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

func (o *approvalReportOptions) run(args []string, request bool) error {
	var err error
	o.payload.ArtifactSha256, err = o.payloadArtifactSHA256(args)
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

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/approvals/", global.Host, global.Owner, o.pipelineName)

	reqParams := &requests.RequestParams{
		Method:   http.MethodPost,
		URL:      url,
		Payload:  o.payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("approval created for artifact: %s", o.payload.ArtifactSha256)
	}
	return err
}

func (o *approvalReportOptions) payloadArtifactSHA256(args []string) (string, error) {
	if o.payload.ArtifactSha256 == "" {
		sha256, err := GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return sha256, err
		}
		return sha256, nil
	}
	return o.payload.ArtifactSha256, nil
}

func (o *approvalReportOptions) payloadReviews(request bool) []map[string]string {
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

func (o *approvalReportOptions) payloadCommitList() ([]string, error) {
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

func (o *approvalReportOptions) commitsHistory() ([]*gitview.ArtifactCommit, error) {
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
