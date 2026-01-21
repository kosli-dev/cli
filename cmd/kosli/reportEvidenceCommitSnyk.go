package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/snyk"
	"github.com/spf13/cobra"
)

type reportEvidenceCommitSnykOptions struct {
	snykJsonFile      string
	userDataFilePath  string
	uploadResultsFile bool
	payload           EvidenceSnykPayload
}

const reportEvidenceCommitSnykShortDesc = `Report Snyk vulnerability scan evidence for a commit in Kosli flows.  `

const reportEvidenceCommitSnykLongDesc = reportEvidenceCommitSnykShortDesc + `  
The --scan-results .json file is parsed and uploaded to Kosli's evidence vault.

In CLI <v2.8.2, Snyk results could only be in the Snyk JSON output format. "snyk code test" results were not supported by 
this command and could be reported as generic evidence.

Starting from v2.8.2, the Snyk results can be in Snyk JSON or SARIF output format for "snyk container test". 
"snyk code test" is now supported but only in the SARIF format.

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
		Use:        "snyk",
		Short:      reportEvidenceCommitSnykShortDesc,
		Long:       reportEvidenceCommitSnykLongDesc,
		Example:    reportEvidenceCommitSnykExample,
		Deprecated: deprecatedKosliReportEvidenceMessage,
		Args:       cobra.NoArgs,
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
	cmd.Flags().BoolVar(&o.uploadResultsFile, "upload-results", true, uploadSnykResultsFlag)

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

	o.payload.SnykSarifResults, err = snyk.ProcessSnykResultFile(o.snykJsonFile)
	if err != nil {
		sarifErr := err
		o.payload.SnykResults, err = LoadJsonData(o.snykJsonFile)
		if err != nil {
			return fmt.Errorf("failed to parse Snyk results file [%s]. Failed to parse as Sarif: %s. Fallen back to parse Snyk Json, but also failed: %s", o.snykJsonFile, err, sarifErr)
		}
	}

	attachments := []string{}
	if o.uploadResultsFile {
		attachments = append(attachments, o.snykJsonFile)
	}

	form, cleanupNeeded, evidencePath, err := newEvidenceForm(o.payload, attachments)
	// if we created a tar package, remove it after uploading it
	if cleanupNeeded {
		defer func() {
			if err := os.Remove(evidencePath); err != nil {
				logger.Warn("failed to remove evidence file %s: %v", evidencePath, err)
			}
		}()
	}

	if err != nil {
		return err
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodPost,
		URL:    url,
		Form:   form,
		DryRun: global.DryRun,
		Token:  global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("snyk scan evidence is reported to commit: %s", o.payload.CommitSHA)
	}
	return err
}
