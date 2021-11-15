package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/merkely-development/reporter/internal/digest"
	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"

	junit "github.com/joshdk/go-junit"
)

type testEvidenceOptions struct {
	artifactType   string
	inputSha256    string
	sha256         string // This is calculated or provided by the user
	pipelineName   string
	description    string
	testResultsDir string
	buildUrl       string
	userDataFile   string
	payload        EvidencePayload
}

func newTestEvidenceCmd(out io.Writer) *cobra.Command {
	o := new(testEvidenceOptions)
	cmd := &cobra.Command{
		Use:   "test ARTIFACT-NAME-OR-PATH",
		Short: "Report/Log a JUnit test evidence to an artifact in Merkely. ",
		Long:  testEvidenceDesc(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("only one argument (docker image name or file/dir path) is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return fmt.Errorf("docker image name or file/dir path is required")
			}

			if o.artifactType == "" && o.inputSha256 == "" {
				return fmt.Errorf("either --type or --sha256 must be specified")
			}

			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return err
			}

			if o.inputSha256 != "" {
				if err := digest.ValidateDigest(o.inputSha256); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if o.inputSha256 != "" {
				o.sha256 = o.inputSha256
			} else {
				o.sha256, err = GetSha256Digest(o.artifactType, args[0])
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

			_, err = requests.SendPayload(o.payload, url, global.ApiToken,
				global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
			return err
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVarP(&o.artifactType, "artifact-type", "t", "", "The type of the artifact related to the evidence. Options are [dir, file, docker].")
	cmd.Flags().StringVarP(&o.inputSha256, "sha256", "s", "", "The SHA256 fingerprint for the artifact. Only required if you don't specify --type.")
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", "The Merkely pipeline name.")
	cmd.Flags().StringVarP(&o.description, "description", "d", "", "[optional] The evidence description.")
	cmd.Flags().StringVarP(&o.buildUrl, "build-url", "b", DefaultValue(ci, "build-url"), "The url of CI pipeline that generated the evidence.")
	cmd.Flags().StringVarP(&o.testResultsDir, "results-dir", "R", "/data/junit/", "The folder with JUnit test results.")
	cmd.Flags().StringVarP(&o.payload.EvidenceType, "evidence-type", "e", "", "The type of evidence being reported.")
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", "[optional] The path to a JSON file containing additional data you would like to attach to this evidence.")

	err := RequireFlags(cmd, []string{"pipeline", "build-url", "evidence-type"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func testEvidenceDesc() string {
	return `
   Report a JUnit test evidence to an artifact in Merkely. 
   The artifact SHA256 fingerprint is calculated or alternatively it can be provided directly. 
   ` + GetCIDefaultsTemplates(supportedCIs, []string{"build-url"})
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
