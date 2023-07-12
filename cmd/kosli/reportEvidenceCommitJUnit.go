package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type reportEvidenceCommitJunitOptions struct {
	testResultsDir   string
	userDataFilePath string
	payload          EvidenceJUnitPayload
}

const reportEvidenceCommitJunitShortDesc = `Report JUnit test evidence for a commit in Kosli flows.  `

const reportEvidenceCommitJunitLongDesc = reportEvidenceCommitJunitShortDesc + `  
All .xml files from --results-dir are parsed and uploaded to Kosli's evidence vault.  
If there are no failing tests and no errors the evidence is reported as compliant. Otherwise the evidence is reported as non-compliant.
`

const reportEvidenceCommitJunitExample = `
# report JUnit test evidence for a commit related to one Kosli flow:
kosli report evidence commit junit \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--flows yourFlowName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName	\
	--results-dir yourFolderWithJUnitResults

# report JUnit test evidence for a commit related to multiple Kosli flows:
kosli report evidence commit junit \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--flows yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName	\
	--results-dir yourFolderWithJUnitResults
`

func newReportEvidenceCommitJunitCmd(out io.Writer) *cobra.Command {
	o := new(reportEvidenceCommitJunitOptions)
	cmd := &cobra.Command{
		Use:     "junit",
		Short:   reportEvidenceCommitJunitShortDesc,
		Long:    reportEvidenceCommitJunitLongDesc,
		Example: reportEvidenceCommitJunitExample,
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
	cmd.Flags().StringVarP(&o.testResultsDir, "results-dir", "R", ".", resultsDirFlag)
	cmd.Flags().StringVarP(&o.userDataFilePath, "user-data", "u", "", evidenceUserDataFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"commit", "build-url", "name"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *reportEvidenceCommitJunitOptions) run(args []string) error {
	var err error
	url := fmt.Sprintf("%s/api/v2/evidence/%s/commit/junit", global.Host, global.Org)
	o.payload.UserData, err = LoadJsonData(o.userDataFilePath)
	if err != nil {
		return err
	}

	o.payload.JUnitResults, err = ingestJunitDir(o.testResultsDir)
	if err != nil {
		return err
	}

	// prepare the files to upload as evidence. We are only interested in the actual Junit XMl files
	junitFilenames, err := getJunitFilenames(o.testResultsDir)
	if err != nil {
		return err
	}

	form, cleanupNeeded, evidencePath, err := newEvidenceForm(o.payload, junitFilenames)
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
		logger.Info("junit test evidence is reported to commit: %s", o.payload.CommitSHA)
	}
	return err
}
