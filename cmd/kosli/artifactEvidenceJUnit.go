package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"

	junit "github.com/joshdk/go-junit"
)

type EvidenceJUnitPayload struct {
    EvidenceName string        `json:"name"`
    BuildUrl     string        `json:"build_url"`
    JUnitResults JUnitResults  `json:"junit_results"`
    UserData     interface{}   `json:"user_data"`
}

type JUnitResults struct {
    Name      string  `json:"name"`
    Failures  int     `json:"failures"`
    Errors    int     `json:"errors"`
    Skipped   int     `json:"skipped"`
    Total     int     `json:"total"`
    Duration  float32 `json:"duration"`
    Timestamp int64   `json:"timestamp"`
}

type junitEvidenceOptions struct {
	fingerprintOptions *fingerprintOptions
	sha256             string // This is calculated or provided by the user
	pipelineName       string
	description        string
	testResultsDir     string
	buildUrl           string
	userDataFile       string
	payload            EvidenceJUnitPayload
}

const junitEvidenceShortDesc = `Report JUnit test evidence for an artifact in a Kosli pipeline.`

const junitEvidenceLongDesc = testEvidenceShortDesc + `
` + sha256Desc

const junitEvidenceExample = `
# report JUnit test evidence about a file artifact:
kosli pipeline artifact report evidence junit FILE.tgz \
	--artifact-type file \
	--name yourEvidenceName \
	--pipeline yourPipelineName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName	\
	--results-dir yourFolderWithJUnitResults

# report JUnit test evidence about an artifact using an available Sha256 digest:
kosli pipeline artifact report evidence junit \
	--sha256 yourSha256 \
	--name yourEvidenceName \
	--pipeline yourPipelineName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName	\
	--results-dir yourFolderWithJUnitResults
`

func newJUnitEvidenceCmd(out io.Writer) *cobra.Command {
	o := new(junitEvidenceOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "junit [ARTIFACT-NAME-OR-PATH]",
		Short:   junitEvidenceShortDesc,
		Long:    junitEvidenceLongDesc,
		Example: junitEvidenceExample,
		Hidden: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.sha256, false)
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return ValidateRegistryFlags(cmd, o.fingerprintOptions)

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVarP(&o.sha256, "sha256", "s", "", sha256Flag)
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", pipelineNameFlag)
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), evidenceBuildUrlFlag)
	cmd.Flags().StringVarP(&o.testResultsDir, "results-dir", "R", "/data/junit/", resultsDirFlag)
	cmd.Flags().StringVarP(&o.payload.EvidenceName, "name", "n", "", evidenceNameFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", evidenceUserDataFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"pipeline", "build-url", "name"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *junitEvidenceOptions) run(args []string) error {
	var err error
	if o.sha256 == "" {
		o.sha256, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s", global.Host, global.Owner, o.pipelineName, o.sha256)
	o.payload.Contents = map[string]interface{}{}

	// o.payload.Contents["is_compliant"], err = isCompliantTestsDir(o.testResultsDir)
	// if err != nil {
	//	return err
	//}

	o.payload.Contents["url"] = o.buildUrl
	o.payload.Contents["description"] = o.description
	//o.payload.Contents["user_data"], err = LoadUserData(o.userDataFile)

	o.payload.UserData, err = LoadUserData(o.userDataFile)
	if err != nil {
		return err
	}

	// _, err = requests.SendPayload(o.payload, url, "", global.ApiToken,
	// 	global.MaxAPIRetries, global.DryRun, http.MethodPut)
	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Payload:  o.payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("test evidence is reported to artifact: %s", o.sha256)
	}
	return err
}

func isCompliantTestsDir(testResultsDir string) (bool, error) {
	suites, err := junit.IngestDir(testResultsDir)
	if err != nil {
		return false, err
	}

	if len(suites) == 0 {
		return false, fmt.Errorf("no tests found in %s directory", testResultsDir)
	}

	for _, suite := range suites {
		for _, test := range suite.Tests {
			if test.Status == junit.StatusFailed || test.Status == junit.StatusError {
				return false, nil
			}
		}
	}

	return true, nil
}
