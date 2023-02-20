package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type genericEvidenceOptions struct {
	fingerprintOptions *fingerprintOptions
	pipelineName       string
	userDataFile       string
	payload            GenericEvidencePayload
}

type EvidencePayload struct {
	EvidenceType string                 `json:"evidence_type"`
	Contents     map[string]interface{} `json:"contents"`
}

type GenericEvidencePayload struct {
	TypedEvidencePayload
	Description string `json:"description"`
	Compliant   bool   `json:"is_compliant"`
}

const artifactEvidenceGenericShortDesc = `Report a generic evidence to an artifact in a Kosli flow.`

const artifactEvidenceGenericLongDesc = artifactEvidenceGenericShortDesc + `
` + sha256Desc

const artifactEvidenceGenericExample = `
# report a generic evidence about a pre-built docker image:
kosli report evidence artifact generic yourDockerImageName \
	--api-token yourAPIToken \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--owner yourOrgName \
	--flow yourFlowName 

# report a generic evidence about a directory type artifact:
kosli report evidence artifact generic /path/to/your/dir \
	--api-token yourAPIToken \
	--artifact-type dir \
	--build-url https://exampleci.com \
	--name yourEvidenceName \
	--owner yourOrgName	\
	--flow yourFlowName 

# report a generic evidence about an artifact with a provided fingerprint (sha256)
kosli report evidence artifact generic \
	--api-token yourAPIToken \
	--build-url https://exampleci.com \	
	--name yourEvidenceName \
	--owner yourOrgName \
	--flow yourFlowName \
	--fingerprint yourFingerprint
`

func newGenericEvidenceCmd(out io.Writer) *cobra.Command {
	o := new(genericEvidenceOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "generic [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   artifactEvidenceGenericShortDesc,
		Long:    artifactEvidenceGenericLongDesc,
		Example: artifactEvidenceGenericExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
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
	cmd.Flags().StringVarP(&o.payload.ArtifactFingerprint, "fingerprint", "f", "", sha256Flag)
	cmd.Flags().StringVarP(&o.pipelineName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", evidenceDescriptionFlag)
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), evidenceBuildUrlFlag)
	cmd.Flags().BoolVarP(&o.payload.Compliant, "compliant", "C", true, evidenceCompliantFlag)
	cmd.Flags().StringVarP(&o.payload.EvidenceName, "name", "n", "", evidenceNameFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", evidenceUserDataFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"flow", "build-url"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *genericEvidenceOptions) run(args []string) error {
	var err error
	if o.payload.ArtifactFingerprint == "" {
		o.payload.ArtifactFingerprint, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/evidence/generic", global.Host, global.Owner, o.pipelineName)

	o.payload.UserData, err = LoadJsonData(o.userDataFile)
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
		logger.Info("generic evidence '%s' is reported to artifact: %s", o.payload.EvidenceName, o.payload.ArtifactFingerprint)
	}
	return err
}
