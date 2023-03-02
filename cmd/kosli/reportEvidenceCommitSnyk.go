package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type reportEvidenceCommitSnykOptions struct {
	snykJsonFile string
	userDataFile string
	payload      EvidenceSnykPayload
}

const reportEvidenceCommitSnykShortDesc = `Report Snyk evidence for a commit in Kosli flows.`

const reportEvidenceCommitSnykLongDesc = reportEvidenceCommitSnykShortDesc

const reportEvidenceCommitSnykExample = `
# report Snyk evidence for a commit related to one Kosli flow:
kosli report evidence commit snyk \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--flows yourFlowName1 \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName	\
	--scan-results yourSnykJSONScanResults

# report Snyk evidence for a commit related to multiple Kosli flows:
kosli report evidence commit snyk \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--flows yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName	\
	--scan-results yourSnykJSONScanResults
`

func newReportEvidenceCommitSnykCmd(out io.Writer) *cobra.Command {
	o := new(reportEvidenceCommitSnykOptions)
	cmd := &cobra.Command{
		Use:     "snyk",
		Short:   reportEvidenceCommitSnykShortDesc,
		Long:    reportEvidenceCommitSnykLongDesc,
		Example: reportEvidenceCommitSnykExample,
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
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
	cmd.Flags().StringVarP(&o.snykJsonFile, "scan-results", "R", "", snykJsonResultsFileFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", evidenceUserDataFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"commit", "build-url", "name", "scan-results"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *reportEvidenceCommitSnykOptions) run(args []string) error {
	var err error
	url := fmt.Sprintf("%s/api/v1/projects/%s/commit/evidence/snyk", global.Host, global.Owner)
	o.payload.UserData, err = LoadJsonData(o.userDataFile)
	if err != nil {
		return err
	}

	o.payload.SnykResults, err = LoadJsonData(o.snykJsonFile)
	if err != nil {
		return err
	}

	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Payload:  o.payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("snyk scan evidence is reported to commit: %s", o.payload.CommitSHA)
	}
	return err
}
