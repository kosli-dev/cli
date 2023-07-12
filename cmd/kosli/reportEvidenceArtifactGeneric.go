package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type reportEvidenceArtifactGenericOptions struct {
	fingerprintOptions *fingerprintOptions
	flowName           string
	userDataFilePath   string
	evidencePaths      []string
	payload            GenericEvidencePayload
}

const reportEvidenceArtifactGenericShortDesc = `Report generic evidence to an artifact in a Kosli flow.  `

const reportEvidenceArtifactGenericLongDesc = reportEvidenceArtifactGenericShortDesc + `
` + fingerprintDesc

const reportEvidenceArtifactGenericExample = `
# report a generic evidence about a pre-built docker image:
kosli report evidence artifact generic yourDockerImageName \
	--api-token yourAPIToken \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--org yourOrgName \
	--flow yourFlowName 

# report a generic evidence about a directory type artifact:
kosli report evidence artifact generic /path/to/your/dir \
	--api-token yourAPIToken \
	--artifact-type dir \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--org yourOrgName	\
	--flow yourFlowName 

# report a generic evidence about an artifact with a provided fingerprint (sha256)
kosli report evidence artifact generic \
	--api-token yourAPIToken \
	--build-url https://exampleci.com \	
	--name yourEvidenceName \
	--org yourOrgName \
	--flow yourFlowName \
	--fingerprint yourArtifactFingerprint

# report a generic evidence about an artifact with evidence file upload
kosli report evidence artifact generic \
	--api-token yourAPIToken \
	--build-url https://exampleci.com \	
	--name yourEvidenceName \
	--org yourOrgName \
	--flow yourFlowName \
	--fingerprint yourArtifactFingerprint \
	--evidence-paths=yourEvidencePathName

# report a generic evidence about an artifact with evidence file upload via API
curl -X 'POST' \
	'https://app.kosli.com/api/v2/evidence/yourOrgName/artifact/yourFlowName/generic' \
	-H 'accept: application/json' \
	-H 'Content-Type: multipart/form-data' \
	-F 'evidence_json={
  	  "artifact_fingerprint": "yourArtifactFingerprint",
	  "name": "yourEvidenceName",
      "build_url": "https://exampleci.com",
      "is_compliant": true
    }' \
	-F 'evidence_file=@yourEvidencePathName'
`

func newReportEvidenceArtifactGenericCmd(out io.Writer) *cobra.Command {
	o := new(reportEvidenceArtifactGenericOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "generic [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   reportEvidenceArtifactGenericShortDesc,
		Long:    reportEvidenceArtifactGenericLongDesc,
		Example: reportEvidenceArtifactGenericExample,
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
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", evidenceDescriptionFlag)
	cmd.Flags().BoolVarP(&o.payload.Compliant, "compliant", "C", true, evidenceCompliantFlag)
	cmd.Flags().StringVarP(&o.userDataFilePath, "user-data", "u", "", evidenceUserDataFlag)
	cmd.Flags().StringSliceVarP(&o.evidencePaths, "evidence-paths", "e", []string{}, evidencePathsFlag)

	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"flow", "build-url", "name"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *reportEvidenceArtifactGenericOptions) run(args []string) error {
	var err error
	if o.payload.ArtifactFingerprint == "" {
		o.payload.ArtifactFingerprint, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	}

	url := fmt.Sprintf("%s/api/v2/evidence/%s/artifact/%s/generic", global.Host, global.Org, o.flowName)

	o.payload.UserData, err = LoadJsonData(o.userDataFilePath)
	if err != nil {
		return err
	}

	form, cleanupNeeded, evidencePath, err := newEvidenceForm(o.payload, o.evidencePaths)
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
		logger.Info("generic evidence '%s' is reported to artifact: %s", o.payload.EvidenceName, o.payload.ArtifactFingerprint)
	}
	return err
}
