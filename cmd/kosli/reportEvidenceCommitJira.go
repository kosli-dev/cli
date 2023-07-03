package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kosli-dev/cli/internal/jira"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type reportEvidenceCommitJiraOptions struct {
	userDataFilePath string
	evidencePaths    []string
	baseURL          string
	payload          JiraEvidencePayload
}

const reportEvidenceCommitJiraShortDesc = `Report Jira evidence for a commit in Kosli flows.`

const reportEvidenceCommitJiraLongDesc = reportEvidenceCommitJiraShortDesc + `
Parses the current commit message for a Jira  reference of the 
form: 'one or more capital letters followed by dash and one or more digits'.
If found and the Jira  exists a compliance status of True is reported.
Otherwise a compliance status of False is reported.
`

const reportEvidenceCommitJiraExample = `
# report Jira  evidence for a commit related to one Kosli flow:
kosli report evidence commit jira \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--jira-base-url https://kosli.atlassian.net/browse/ \
	--flows yourFlowName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName

# report Jira  evidence for a commit related to multiple Kosli flows with user-data:
kosli report evidence commit jira \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--jira-base-url https://kosli.atlassian.net/browse/ \
	--flows yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName \
	--user-data /path/to/json/file.json
`

func newReportEvidenceCommitJiraCmd(out io.Writer) *cobra.Command {
	o := new(reportEvidenceCommitJiraOptions)
	cmd := &cobra.Command{
		Use:     "jira",
		Short:   reportEvidenceCommitJiraShortDesc,
		Long:    reportEvidenceCommitJiraLongDesc,
		Example: reportEvidenceCommitJiraExample,
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
	cmd.Flags().StringVarP(&o.baseURL, "jira-base-url", "j", "", jiraBaseUrlFlag)
	cmd.Flags().StringVarP(&o.userDataFilePath, "user-data", "u", "", evidenceUserDataFlag)
	cmd.Flags().StringSliceVarP(&o.evidencePaths, "evidence-paths", "e", []string{}, evidencePathsFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"commit", "build-url", "name", "jira-base-url"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *reportEvidenceCommitJiraOptions) run(args []string) error {
	var err error
	url := fmt.Sprintf("%s/api/v2/evidence/%s/commit/jira", global.Host, global.Org)
	o.payload.UserData, err = LoadJsonData(o.userDataFilePath)
	if err != nil {
		return err
	}

	result, err := jira.GetJiraIssue(o.baseURL, "EX-1")
	if err != nil {
		return err
	}
	o.payload.JiraResults = append(o.payload.JiraResults, result)

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
		logger.Info("jira evidence '%s' is reported to commit: %s", o.payload.EvidenceName, o.payload.CommitSHA)
	}
	return err
}
