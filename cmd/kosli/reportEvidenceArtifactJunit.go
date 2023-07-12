package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"

	junit "github.com/joshdk/go-junit"
)

type EvidenceJUnitPayload struct {
	TypedEvidencePayload
	JUnitResults []*JUnitResults `json:"junit_results"`
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

type reportEvidenceArtifactJunitOptions struct {
	fingerprintOptions *fingerprintOptions
	flowName           string
	testResultsDir     string
	userDataFilePath   string
	payload            EvidenceJUnitPayload
}

const reportEvidenceArtifactJunitShortDesc = `Report JUnit test evidence for an artifact in a Kosli flow.  `

const reportEvidenceArtifactJunitLongDesc = reportEvidenceArtifactJunitShortDesc + `  
All .xml files from --results-dir are parsed and uploaded to Kosli's evidence vault.  
If there are no failing tests and no errors the evidence is reported as compliant. Otherwise the evidence is reported as non-compliant.  
` + fingerprintDesc

const reportEvidenceArtifactJunitExample = `
# report JUnit test evidence about a file artifact:
kosli report evidence artifact junit FILE.tgz \
	--artifact-type file \
	--name yourEvidenceName \
	--flow yourFlowName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName	\
	--results-dir yourFolderWithJUnitResults

# report JUnit test evidence about an artifact using an available Sha256 digest:
kosli report evidence artifact junit \
	--fingerprint yourSha256 \
	--name yourEvidenceName \
	--flow yourFlowName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName	\
	--results-dir yourFolderWithJUnitResults
`

func newReportEvidenceArtifactJunitCmd(out io.Writer) *cobra.Command {
	o := new(reportEvidenceArtifactJunitOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "junit [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   reportEvidenceArtifactJunitShortDesc,
		Long:    reportEvidenceArtifactJunitLongDesc,
		Example: reportEvidenceArtifactJunitExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.payload.ArtifactFingerprint, false)
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
	addArtifactEvidenceFlags(cmd, &o.payload.TypedEvidencePayload, ci)
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.testResultsDir, "results-dir", "R", ".", resultsDirFlag)
	cmd.Flags().StringVarP(&o.userDataFilePath, "user-data", "u", "", evidenceUserDataFlag)

	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"flow", "build-url", "name"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *reportEvidenceArtifactJunitOptions) run(args []string) error {
	var err error
	if o.payload.ArtifactFingerprint == "" {
		o.payload.ArtifactFingerprint, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	}
	url := fmt.Sprintf("%s/api/v2/evidence/%s/artifact/%s/junit", global.Host, global.Org, o.flowName)
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
		var timestamp float64
		// There is no official schema for the timestamp in the junit xml
		suite_timestamp := suite.Properties["timestamp"]
		if suite_timestamp != "" {
			// This one comes from pytest
			createdAt, err := time.Parse("2006-01-02T15:04:05.999999", suite_timestamp)
			if err != nil {
				// This one comes from Ruby minitest
				createdAt, err = time.Parse("2006-01-02T15:04:05+00:00", suite_timestamp)
				if err != nil {
					return results, err
				}
			}
			timestamp = float64(createdAt.UTC().Unix())
		} else {
			// maven surefire plugin generates Junit xml with no timestamp
			timestamp = 0.0
		}

		// The values in suite.Totals are based on the results of the tests in the suite and not in the header of the suite.
		suiteResult := &JUnitResults{
			Name:      suite.Name,
			Duration:  suite.Totals.Duration.Seconds(),
			Total:     suite.Totals.Tests,
			Skipped:   suite.Totals.Skipped,
			Errors:    suite.Totals.Error,
			Failures:  suite.Totals.Failed,
			Timestamp: timestamp,
		}
		logger.Debug("parsed <testsuite> result: %+v", suiteResult)
		results = append(results, suiteResult)
	}

	return results, nil
}

func getJunitFilenames(directory string) ([]string, error) {
	var filenames []string

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Add all regular files that end with ".xml"
		if info.Mode().IsRegular() && strings.HasSuffix(info.Name(), ".xml") {
			suites, err := junit.IngestFile(path)
			if err != nil {
				return err
			}
			if len(suites) > 0 {
				filenames = append(filenames, path)
			}
		}

		return nil
	})

	if err != nil {
		return filenames, err
	}

	return filenames, nil
}
