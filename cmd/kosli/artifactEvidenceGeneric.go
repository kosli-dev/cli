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
	sha256             string // This is calculated or provided by the user
	pipelineName       string
	description        string
	isCompliant        bool
	buildUrl           string
	userDataFile       string
	payload            EvidencePayload
}

type EvidencePayload struct {
	EvidenceType string                 `json:"evidence_type"`
	Contents     map[string]interface{} `json:"contents"`
}

const artifactEvidenceGenericExample = `
# report a generic evidence about a pre-built docker image:
kosli pipeline artifact report evidence generic yourDockerImageName \
	--api-token yourAPIToken \
	--artifact-type docker \
	--build-url https://exampleci.com \
	--evidence-type yourEvidenceType \
	--owner yourOrgName \
	--pipeline yourPipelineName 

# report a generic evidence about a directory type artifact:
kosli pipeline artifact report evidence generic /path/to/your/dir \
	--api-token yourAPIToken \
	--artifact-type dir \
	--build-url https://exampleci.com \
	--evidence-type yourEvidenceType \
	--owner yourOrgName	\
	--pipeline yourPipelineName 


# report a generic evidence about an artifact with a provided fingerprint (sha256)
kosli pipeline artifact report evidence generic \
	--api-token yourAPIToken \
	--build-url https://exampleci.com \	
	--evidence-type yourEvidenceType \
	--owner yourOrgName \
	--pipeline yourPipelineName \
	--sha256 yourSha256
`

func newGenericEvidenceCmd(out io.Writer) *cobra.Command {
	o := new(genericEvidenceOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "generic [ARTIFACT-NAME-OR-PATH]",
		Short:   "Report a generic evidence to an artifact in a Kosli pipeline. ",
		Example: artifactEvidenceGenericExample,
		Long:    genericEvidenceDesc(),
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
	cmd.Flags().StringVarP(&o.description, "description", "d", "", evidenceDescriptionFlag)
	cmd.Flags().StringVarP(&o.buildUrl, "build-url", "b", DefaultValue(ci, "build-url"), evidenceBuildUrlFlag)
	cmd.Flags().BoolVarP(&o.isCompliant, "compliant", "C", true, evidenceCompliantFlag)
	cmd.Flags().StringVarP(&o.payload.EvidenceType, "evidence-type", "e", "", evidenceTypeFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", evidenceUserDataFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)

	err := RequireFlags(cmd, []string{"pipeline", "build-url", "evidence-type"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *genericEvidenceOptions) run(args []string) error {
	var err error
	if o.sha256 == "" {
		o.sha256, err = GetSha256Digest(args[0], o.fingerprintOptions)
		if err != nil {
			return err
		}
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s", global.Host, global.Owner, o.pipelineName, o.sha256)
	o.payload.Contents = map[string]interface{}{}
	o.payload.Contents["is_compliant"] = o.isCompliant
	o.payload.Contents["url"] = o.buildUrl
	o.payload.Contents["description"] = o.description
	o.payload.Contents["user_data"], err = LoadUserData(o.userDataFile)
	if err != nil {
		return err
	}

	_, err = requests.SendPayload(o.payload, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodPut)
	return err
}

func genericEvidenceDesc() string {
	return `
   Report a generic evidence to an artifact to a Kosli pipeline. 
   ` + sha256Desc
}
