package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type EvidenceSnykPayload struct {
	TypedEvidencePayload
	SnykResults interface{} `json:"snyk_results"`
}

type reportEvidenceArtifactSnykOptions struct {
	fingerprintOptions *fingerprintOptions
	flowName           string
	snykJsonFilePath   string
	userDataFilePath   string
	payload            EvidenceSnykPayload
}

const reportEvidenceArtifactSnykShortDesc = `Report Snyk vulnerability scan evidence for an artifact in a Kosli flow.  `

const reportEvidenceArtifactSnykLongDesc = reportEvidenceArtifactSnykShortDesc + `  
The --scan-results .json file is parsed and uploaded to Kosli's evidence vault.  
If no vulnerabilities are detected the evidence is reported as compliant. Otherwise the evidence is reported as non-compliant.  
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
		Use:     "snyk [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   reportEvidenceArtifactSnykShortDesc,
		Long:    reportEvidenceArtifactSnykLongDesc,
		Example: reportEvidenceArtifactSnykExample,
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

	o.payload.SnykResults, err = LoadJsonData(o.snykJsonFilePath)
	if err != nil {
		return err
	}

	form, cleanupNeeded, evidencePath, err := newEvidenceForm(o.payload, []string{o.snykJsonFilePath})
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
		logger.Info("snyk scan evidence is reported to artifact: %s", o.payload.ArtifactFingerprint)
	}
	return err
}
