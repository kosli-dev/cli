package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type CommitEvidenceJUnitPayload struct {
	CommitSHA    string          `json:"commit_sha"`
	Flows        []string        `json:"pipelines,omitempty"`
	EvidenceName string          `json:"name"`
	BuildUrl     string          `json:"build_url"`
	JUnitResults []*JUnitResults `json:"junit_results"`
	UserData     interface{}     `json:"user_data"`
}

type reportEvidenceCommitJunitOptions struct {
	testResultsDir string
	userDataFile   string
	payload        CommitEvidenceJUnitPayload
}

const reportEvidenceCommitJunitShortDesc = `Report JUnit test evidence for a commit in a Kosli flow.`

const reportEvidenceCommitJunitLongDesc = reportEvidenceCommitJunitShortDesc

const reportEvidenceCommitJunitExample = `
# report JUnit test evidence for a commit related to one Kosli flow:
kosli report evidence commit junit \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--flow yourFlowName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName	\
	--results-dir yourFolderWithJUnitResults

# report JUnit test evidence for a commit related to multiple Kosli flows:
kosli report evidence commit junit \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--flow yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName	\
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
	cmd.Flags().StringVar(&o.payload.CommitSHA, "commit", DefaultValue(ci, "git-commit"), evidenceCommitFlag)
	cmd.Flags().StringSliceVarP(&o.payload.Flows, "flow", "f", []string{}, pipelinesFlag)
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), evidenceBuildUrlFlag)
	cmd.Flags().StringVarP(&o.testResultsDir, "results-dir", "R", ".", resultsDirFlag)
	cmd.Flags().StringVarP(&o.payload.EvidenceName, "name", "n", "", evidenceNameFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", evidenceUserDataFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"commit", "build-url", "name"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *reportEvidenceCommitJunitOptions) run(args []string) error {
	var err error
	url := fmt.Sprintf("%s/api/v1/projects/%s/commit/evidence/junit", global.Host, global.Owner)
	o.payload.UserData, err = LoadJsonData(o.userDataFile)
	if err != nil {
		return err
	}

	o.payload.JUnitResults, err = ingestJunitDir(o.testResultsDir)
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
		logger.Info("junit test evidence is reported to commit: %s", o.payload.CommitSHA)
	}
	return err
}
