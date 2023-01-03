package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type EvidenceSnykPayload struct {
	// TODO: Put version in payload
	ArtifactFingerprint string      `json:"artifact_fingerprint"`
	EvidenceName        string      `json:"name"`
	BuildUrl            string      `json:"build_url"`
	SnykResults         interface{} `json:"snyk_results"`
	UserData            interface{} `json:"user_data"`
}

type snykEvidenceOptions struct {
	fingerprintOptions *fingerprintOptions
	fingerprint        string // This is calculated or provided by the user
	pipelineName       string
	snykJsonFile       string
	userDataFile       string
	payload            EvidenceSnykPayload
}

const snykEvidenceShortDesc = `Report Snyk vulnerability scan evidence for an artifact in a Kosli pipeline.`

const snykEvidenceLongDesc = testEvidenceShortDesc + `
` + sha256Desc

const snykEvidenceExample = `
# report Snyk vulnerability scan evidence about a file artifact:
kosli pipeline artifact report evidence snyk FILE.tgz \
	--artifact-type file \
	--name yourEvidenceName \
	--pipeline yourPipelineName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName	\
	--scan-results yourSnykJSONScanResults

# report Snyk vulnerability scan evidence about an artifact using an available Sha256 digest:
kosli pipeline artifact report evidence snyk \
	--fingerprint yourSha256 \
	--name yourEvidenceName \
	--pipeline yourPipelineName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName	\
	--scan-results yourSnykJSONScanResults
`

func newSnykEvidenceCmd(out io.Writer) *cobra.Command {
	o := new(snykEvidenceOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "snyk [ARTIFACT-NAME-OR-PATH]",
		Short:   snykEvidenceShortDesc,
		Long:    snykEvidenceLongDesc,
		Example: snykEvidenceExample,
		Hidden:  true,
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
	cmd.Flags().StringVarP(&o.snykJsonFile, "scan-results", "R", "", snykJsonResultsFileFlag)
	cmd.Flags().StringVarP(&o.payload.EvidenceName, "name", "n", "", evidenceNameFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", evidenceUserDataFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"pipeline", "build-url", "name", "scan-results"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *snykEvidenceOptions) run(args []string) error {
	var err error
	if o.fingerprint == "" {
		o.payload.ArtifactFingerprint, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	} else {
		o.payload.ArtifactFingerprint = o.fingerprint
	}
	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/evidence/snyk/", global.Host, global.Owner, o.pipelineName)
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
		logger.Info("snyk scan evidence is reported to artifact: %s", o.fingerprint)
	}
	return err
}
