package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type reportEvidenceCommitSnykOptions struct {
	snykJsonFile     string
	userDataFilePath string
	payload          EvidenceSnykPayload
}

const reportEvidenceCommitSnykShortDesc = `Report Snyk vulnerability scan evidence for a commit in Kosli flows.  `

const reportEvidenceCommitSnykLongDesc = reportEvidenceCommitSnykShortDesc + `  
The --scan-results .json file is parsed and uploaded to Kosli's evidence vault.  
If no vulnerabilities are detected the evidence is reported as compliant. Otherwise the evidence is reported as non-compliant.
`

const reportEvidenceCommitSnykExample = `
# report Snyk evidence for a commit related to one Kosli flow:
kosli report evidence commit snyk \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--flows yourFlowName1 \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName	\
	--scan-results yourSnykJSONScanResults

# report Snyk evidence for a commit related to multiple Kosli flows:
kosli report evidence commit snyk \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--flows yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName	\
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
	cmd.Flags().StringVarP(&o.snykJsonFile, "scan-results", "R", "", snykJsonResultsFileFlag)
	cmd.Flags().StringVarP(&o.userDataFilePath, "user-data", "u", "", evidenceUserDataFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"commit", "build-url", "name", "scan-results"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *reportEvidenceCommitSnykOptions) run(args []string) error {
	var err error
	url := fmt.Sprintf("%s/api/v2/evidence/%s/commit/snyk", global.Host, global.Org)
	o.payload.UserData, err = LoadJsonData(o.userDataFilePath)
	if err != nil {
		return err
	}

	o.payload.SnykResults, err = LoadJsonData(o.snykJsonFile)
	if err != nil {
		return err
	}

	form, cleanupNeeded, evidencePath, err := newEvidenceForm(o.payload, []string{o.snykJsonFile})
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
		logger.Info("snyk scan evidence is reported to commit: %s", o.payload.CommitSHA)
	}
	return err
}
