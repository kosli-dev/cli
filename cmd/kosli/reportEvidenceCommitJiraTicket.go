package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type reportEvidenceCommitJiraTicketOptions struct {
	userDataFilePath string
	evidencePaths    []string
	payload          JiraTicketEvidencePayload
}

const reportEvidenceCommitJiraTicketShortDesc = `Report JiraTicket evidence for a commit in Kosli flows.`

const reportEvidenceCommitJiraTicketLongDesc = reportEvidenceCommitJiraTicketShortDesc

const reportEvidenceCommitJiraTicketExample = `
# report JiraTicket evidence for a commit related to one Kosli flow:
kosli report evidence commit generic \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--description "some description" \
	--compliant \
	--flows yourFlowName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName

# report JiraTicket evidence for a commit related to multiple Kosli flows with user-data:
kosli report evidence commit generic \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--description "some description" \
	--compliant \
	--flows yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName \
	--user-data /path/to/json/file.json
`

func newReportEvidenceCommitJiraTicketCmd(out io.Writer) *cobra.Command {
	o := new(reportEvidenceCommitJiraTicketOptions)
	cmd := &cobra.Command{
		Use:     "jira-ticket",
		Short:   reportEvidenceCommitJiraTicketShortDesc,
		Long:    reportEvidenceCommitJiraTicketLongDesc,
		Example: reportEvidenceCommitJiraTicketExample,
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	ci := WhichCI()
	addCommitEvidenceFlags(cmd, &o.payload.TypedEvidencePayload, ci)
	cmd.Flags().BoolVarP(&o.payload.Compliant, "compliant", "C", false, evidenceCompliantFlag)
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", evidenceDescriptionFlag)
	cmd.Flags().StringVarP(&o.userDataFilePath, "user-data", "u", "", evidenceUserDataFlag)
	cmd.Flags().StringSliceVarP(&o.evidencePaths, "evidence-paths", "e", []string{}, evidencePathsFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"commit", "build-url", "name"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *reportEvidenceCommitJiraTicketOptions) run(args []string) error {
	var err error
	url := fmt.Sprintf("%s/api/v2/evidence/%s/commit/generic", global.Host, global.Org)
	o.payload.UserData, err = LoadJsonData(o.userDataFilePath)
	if err != nil {
		return err
	}

	form, cleanupNeeded, evidencePath, err := newEvidenceForm(o.payload, o.evidencePaths)
	// if we created a tar package, remove it after uploading it
	if cleanupNeeded {
		defer os.Remove(evidencePath)
	}

	if err != nil {
		return err
	}

	reqParams := &requests.RequestParams{
		Method:   http.MethodPost,
		URL:      url,
		Form:     form,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}

	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("generic evidence '%s' is reported to commit: %s", o.payload.EvidenceName, o.payload.CommitSHA)
	}
	return err
}
