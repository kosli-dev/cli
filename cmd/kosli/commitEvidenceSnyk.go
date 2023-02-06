package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type CommitEvidenceSnykPayload struct {
	CommitSHA    string      `json:"commit_sha"`
	Pipelines    []string    `json:"pipelines,omitempty"`
	EvidenceName string      `json:"name"`
	BuildUrl     string      `json:"build_url"`
	SnykResults  interface{} `json:"snyk_results"`
	UserData     interface{} `json:"user_data"`
}

type snykCommitEvidenceOptions struct {
	snykJsonFile string
	userDataFile string
	payload      CommitEvidenceSnykPayload
}

const snykCommitEvidenceShortDesc = `Report Snyk evidence for a commit in a Kosli pipeline.`

const snykCommitEvidenceLongDesc = snykEvidenceShortDesc

const snykCommitEvidenceExample = `
# report Snyk evidence for a commit related to one Kosli pipeline:
kosli commit report evidence snyk \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--pipelines yourPipelineName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName	\
	--scan-results yourSnykJSONScanResults

# report Snyk evidence for a commit related to multiple Kosli pipelines:
kosli commit report evidence snyk \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--pipelines yourFirstPipelineName,yourSecondPipelineName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName	\
	--scan-results yourSnykJSONScanResults
`

func newSnykCommitEvidenceCmd(out io.Writer) *cobra.Command {
	o := new(snykCommitEvidenceOptions)
	cmd := &cobra.Command{
		Use:     "snyk",
		Hidden:  true,
		Short:   snykCommitEvidenceShortDesc,
		Long:    snykCommitEvidenceLongDesc,
		Example: snykCommitEvidenceExample,
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
	cmd.Flags().StringVar(&o.payload.CommitSHA, "commit", "", evidenceCommit)
	cmd.Flags().StringSliceVarP(&o.payload.Pipelines, "pipelines", "p", []string{}, pipelinesFlag)
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), evidenceBuildUrlFlag)
	cmd.Flags().StringVarP(&o.snykJsonFile, "scan-results", "R", ".", snykJsonResultsFileFlag)
	cmd.Flags().StringVarP(&o.payload.EvidenceName, "name", "n", "", evidenceNameFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", evidenceUserDataFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"commit", "build-url", "name"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *snykCommitEvidenceOptions) run(args []string) error {
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
