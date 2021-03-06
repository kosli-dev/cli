package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"

	junit "github.com/joshdk/go-junit"
)

type testEvidenceOptions struct {
	fingerprintOptions *fingerprintOptions
	sha256             string // This is calculated or provided by the user
	pipelineName       string
	description        string
	testResultsDir     string
	buildUrl           string
	userDataFile       string
	payload            EvidencePayload
}

const testEvidenceExample = `
# report a JUnit test evidence about a file artifact:
kosli pipeline artifact report evidence test FILE.tgz \
	--artifact-type file \
	--evidence-type yourEvidenceType \
	--pipeline yourPipelineName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName	\
	--results-dir yourFolderWithJUnitResults

# report a JUnit test evidence about an artifact using an available Sha256 digest:
kosli pipeline artifact report evidence test \
	--sha256 yourSha256 \
	--evidence-type yourEvidenceType \
	--pipeline yourPipelineName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName	\
	--results-dir yourFolderWithJUnitResults
`

func newTestEvidenceCmd(out io.Writer) *cobra.Command {
	o := new(testEvidenceOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "test [ARTIFACT-NAME-OR-PATH]",
		Short:   "Report a JUnit test evidence to an artifact in a Kosli pipeline. ",
		Long:    testEvidenceDesc(),
		Example: testEvidenceExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorAfterPrintingHelp(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.sha256, false)
			if err != nil {
				return ErrorAfterPrintingHelp(cmd, err.Error())
			}
			return ValidateRegisteryFlags(cmd, o.fingerprintOptions)

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVarP(&o.sha256, "sha256", "s", "", sha256Flag)
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", pipelineNameFlag)
	cmd.Flags().StringVarP(&o.description, "description", "d", "", evidenceDescriptionFlag)
	cmd.Flags().StringVarP(&o.buildUrl, "build-url", "b", DefaultValue(ci, "build-url"), evidenceBuildUrlFlag)
	cmd.Flags().StringVarP(&o.testResultsDir, "results-dir", "R", "/data/junit/", resultsDirFlag)
	cmd.Flags().StringVarP(&o.payload.EvidenceType, "evidence-type", "e", "", evidenceTypeFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", evidenceUserDataFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)

	err := RequireFlags(cmd, []string{"pipeline", "build-url", "evidence-type"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func testEvidenceDesc() string {
	return `
   Report a JUnit test evidence to an artifact in a Kosli pipeline. 
   The artifact SHA256 fingerprint is calculated or alternatively it can be provided directly. 
   `
}

func (o *testEvidenceOptions) run(args []string) error {
	var err error
	if o.sha256 == "" {
		o.sha256, err = GetSha256Digest(args[0], o.fingerprintOptions)
		if err != nil {
			return err
		}
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s", global.Host, global.Owner, o.pipelineName, o.sha256)
	o.payload.Contents = map[string]interface{}{}
	o.payload.Contents["is_compliant"], err = isCompliantTestsDir(o.testResultsDir)
	if err != nil {
		return err
	}
	o.payload.Contents["url"] = o.buildUrl
	o.payload.Contents["description"] = o.description
	o.payload.Contents["user_data"], err = LoadUserData(o.userDataFile)
	if err != nil {
		return err
	}

	_, err = requests.SendPayload(o.payload, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
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
