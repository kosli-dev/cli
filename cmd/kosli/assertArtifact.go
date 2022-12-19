package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const assertArtifactShortDesc = `Assert the compliance status of an artifact in Kosli.`

const assertArtifactLongDesc = assertArtifactShortDesc + `
Exits with non-zero code if the artifact has a non-compliant status.`

const assertArtifactExample = `
# fail if an artifact has a non-compliant status (using the artifact fingerprint)
kosli assert artifact \
	--sha256 184c799cd551dd1d8d5c5f9a5d593b2e931f5e36122ee5c793c1d08a19839cc0 \
	--pipeline yourPipelineName \
	--api-token yourAPIToken \
	--owner yourOrgName 

# fail if an artifact has a non-compliant status (using the artifact name and type)
kosli assert artifact library/nginx:1.21 \
	--artifact-type docker \
	--pipeline yourPipelineName \
	--api-token yourAPIToken \
	--owner yourOrgName 
`

type assertArtifactOptions struct {
	fingerprintOptions *fingerprintOptions
	sha256             string // This is calculated or provided by the user
	pipelineName       string
}

func newAssertArtifactCmd(out io.Writer) *cobra.Command {
	o := &assertArtifactOptions{}
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "artifact [ARTIFACT-NAME-OR-PATH]",
		Short:   assertArtifactShortDesc,
		Long:    assertArtifactLongDesc,
		Example: assertArtifactExample,
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
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.sha256, "sha256", "s", "", sha256Flag)
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", pipelineNameFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"pipeline"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *assertArtifactOptions) run(out io.Writer, args []string) error {
	var err error
	if o.sha256 == "" {
		o.sha256, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/artifact/?snappish=%s@%s", global.Host, global.Owner, o.pipelineName, o.sha256)

	reqParams := &requests.RequestParams{
		Method:   http.MethodGet,
		URL:      url,
		Password: global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	var artifactData map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &artifactData)
	if err != nil {
		return err
	}

	if artifactData["state"].(string) == "COMPLIANT" {
		logger.Info("COMPLIANT")
	} else {
		fmt.Fprintf(out, "%s: %s\n", artifactData["state"].(string), artifactData["state_info"].(string))
		os.Exit(1)
	}

	return nil
}
