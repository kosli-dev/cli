package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"

	junit "github.com/joshdk/go-junit"
)

type EvidenceJUnitPayload struct {
	// TODO: Put version in payload
	ArtifactFingerprint string          `json:"artifact_fingerprint"`
	EvidenceName        string          `json:"name"`
	BuildUrl            string          `json:"build_url"`
	JUnitResults        []*JUnitResults `json:"junit_results"`
	UserData            interface{}     `json:"user_data"`
}

type JUnitResults struct {
	Name      string  `json:"name"`
	Failures  int     `json:"failures"`
	Errors    int     `json:"errors"`
	Skipped   int     `json:"skipped"`
	Total     int     `json:"total"`
	Duration  float64 `json:"duration"`
	Timestamp float64 `json:"timestamp,omitempty"`
}

type junitEvidenceOptions struct {
	fingerprintOptions *fingerprintOptions
	fingerprint        string // This is calculated or provided by the user
	pipelineName       string
	testResultsDir     string
	userDataFile       string
	payload            EvidenceJUnitPayload
}

const junitEvidenceShortDesc = `Report JUnit test evidence for an artifact in a Kosli pipeline.`

const junitEvidenceLongDesc = junitEvidenceShortDesc + `
` + fingerprintDesc

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
	--fingerprint yourSha256 \
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
		Use:     "junit [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   junitEvidenceShortDesc,
		Long:    junitEvidenceLongDesc,
		Example: junitEvidenceExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.fingerprint, false)
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
	cmd.Flags().StringVarP(&o.fingerprint, "fingerprint", "f", "", fingerprintFlag)
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", pipelineNameFlag)
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), evidenceBuildUrlFlag)
	cmd.Flags().StringVarP(&o.testResultsDir, "results-dir", "R", ".", resultsDirFlag)
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
	if o.fingerprint == "" {
		o.payload.ArtifactFingerprint, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	} else {
		o.payload.ArtifactFingerprint = o.fingerprint
	}
	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/evidence/junit", global.Host, global.Owner, o.pipelineName)
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
		logger.Info("junit test evidence is reported to artifact: %s", o.payload.ArtifactFingerprint)
	}
	return err
}

func ingestJunitDir(testResultsDir string) ([]*JUnitResults, error) {
	results := []*JUnitResults{}
	suites, err := junit.IngestDir(testResultsDir)
	if err != nil {
		return results, err
	}

	if len(suites) == 0 {
		return results, fmt.Errorf("no tests found in %s directory", testResultsDir)
	}

	for _, suite := range suites {
		errors, err := strconv.Atoi(suite.Properties["errors"])
		if err != nil {
			return results, err
		}
		failures, err := strconv.Atoi(suite.Properties["failures"])
		if err != nil {
			return results, err
		}
		duration, err := strconv.ParseFloat(suite.Properties["time"], 64)
		if err != nil {
			return results, err
		}

		// There is no official schema for the timestamp in the junit xml
		// This one comes from pytest
		createdAt, err := time.Parse("2006-01-02T15:04:05.999999", suite.Properties["timestamp"])
		if err != nil {
			// This one comes from Ruby minitest
			createdAt, err = time.Parse("2006-01-02T15:04:05+00:00", suite.Properties["timestamp"])
		}

		// maven surefire plugin generates Junit xml with no timestamp
		var timestamp float64
		if err == nil {
			timestamp = float64(createdAt.UTC().Unix())
		} else {
			timestamp = 0.0
		}

		suiteResult := &JUnitResults{
			Name:      suite.Name,
			Duration:  duration,
			Total:     suite.Totals.Tests,
			Skipped:   suite.Totals.Skipped,
			Errors:    errors,
			Failures:  failures,
			Timestamp: timestamp,
		}
		logger.Debug("parsed <testsuite> result: %+v", suiteResult)
		results = append(results, suiteResult)
	}

	return results, nil
}
