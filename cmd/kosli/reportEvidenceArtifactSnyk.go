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

type EvidenceSnykPayload struct {
	TypedEvidencePayload
	SnykResults      interface{}    `json:"snyk_results,omitempty"`
	SnykSarifResults *snyk.SnykData `json:"processed_snyk_results,omitempty"`
}

type reportEvidenceArtifactSnykOptions struct {
	fingerprintOptions *fingerprintOptions
	flowName           string
	snykJsonFilePath   string
	userDataFilePath   string
	uploadResultsFile  bool
	payload            EvidenceSnykPayload
}

const reportEvidenceArtifactSnykShortDesc = `Report Snyk vulnerability scan evidence for an artifact in a Kosli flow.  `

const reportEvidenceArtifactSnykLongDesc = reportEvidenceArtifactSnykShortDesc + `  
The --scan-results .json file is parsed and uploaded to Kosli's evidence vault.

In CLI <v2.8.2, Snyk results could only be in the Snyk JSON output format. "snyk code test" results were not supported by 
this command and could be reported as generic evidence.

Starting from v2.8.2, the Snyk results can be in Snyk JSON or SARIF output format for "snyk container test". 
"snyk code test" is now supported but only in the SARIF format.

If no vulnerabilities are detected, the evidence is reported as compliant. Otherwise the evidence is reported as non-compliant.

` + fingerprintDesc

const reportEvidenceArtifactSnykExample = `
# report Snyk vulnerability scan evidence about a file artifact:
kosli report evidence artifact snyk FILE.tgz \
	--artifact-type file \
	--name yourEvidenceName \
	--flow yourFlowName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName	\
	--scan-results yourSnykJSONScanResults

# report Snyk vulnerability scan evidence about an artifact using an available Sha256 digest:
kosli report evidence artifact snyk \
	--fingerprint yourSha256 \
	--name yourEvidenceName \
	--flow yourFlowName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName	\
	--scan-results yourSnykJSONScanResults
`

func newReportEvidenceArtifactSnykCmd(out io.Writer) *cobra.Command {
	o := new(reportEvidenceArtifactSnykOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:        "snyk [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:      reportEvidenceArtifactSnykShortDesc,
		Long:       reportEvidenceArtifactSnykLongDesc,
		Example:    reportEvidenceArtifactSnykExample,
		Deprecated: deprecatedKosliReportEvidenceMessage,
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
	cmd.Flags().StringVarP(&o.snykJsonFilePath, "scan-results", "R", "", snykJsonResultsFileFlag)
	cmd.Flags().StringVarP(&o.userDataFilePath, "user-data", "u", "", evidenceUserDataFlag)
	cmd.Flags().BoolVar(&o.uploadResultsFile, "upload-results", true, uploadSnykResultsFlag)

	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"flow", "build-url", "name", "scan-results"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *reportEvidenceArtifactSnykOptions) run(args []string) error {
	var err error
	if o.payload.ArtifactFingerprint == "" {
		o.payload.ArtifactFingerprint, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	}
	url := fmt.Sprintf("%s/api/v2/evidence/%s/artifact/%s/snyk", global.Host, global.Org, o.flowName)
	o.payload.UserData, err = LoadJsonData(o.userDataFilePath)
	if err != nil {
		return err
	}

	o.payload.SnykSarifResults, err = snyk.ProcessSnykResultFile(o.snykJsonFilePath)
	if err != nil {
		sarifErr := err
		o.payload.SnykResults, err = LoadJsonData(o.snykJsonFilePath)
		if err != nil {
			return fmt.Errorf("failed to parse Snyk results file [%s]. Failed to parse as Sarif: %s. Fallen back to parse Snyk Json, but also failed: %s", o.snykJsonFilePath, err, sarifErr)
		}
	}

	attachments := []string{}
	if o.uploadResultsFile {
		attachments = append(attachments, o.snykJsonFilePath)
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
		logger.Info("snyk scan evidence is reported to artifact: %s", o.payload.ArtifactFingerprint)
	}
	return err
}
