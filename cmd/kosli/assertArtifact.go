package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const assertArtifactShortDesc = `Assert the compliance status of an artifact in Kosli (in its flow or against an environment).  `

const assertArtifactLongDesc = assertArtifactShortDesc + `
Exits with non-zero code if the artifact has a non-compliant status.`

const assertArtifactExample = `
# assert that an artifact meets all compliance requirements for an environment
kosli assert artifact \
	--fingerprint 184c799cd551dd1d8d5c5f9a5d593b2e931f5e36122ee5c793c1d08a19839cc0 \
	--flow yourFlowName \
	--against-env prod \
	--api-token yourAPIToken \
	--org yourOrgName 

# fail if an artifact has a non-compliant status (using the artifact fingerprint)
kosli assert artifact \
	--fingerprint 184c799cd551dd1d8d5c5f9a5d593b2e931f5e36122ee5c793c1d08a19839cc0 \
	--flow yourFlowName \
	--api-token yourAPIToken \
	--org yourOrgName 

# fail if an artifact has a non-compliant status (using the artifact name and type)
kosli assert artifact library/nginx:1.21 \
	--artifact-type docker \
	--flow yourFlowName \
	--api-token yourAPIToken \
	--org yourOrgName 
`

type assertArtifactOptions struct {
	fingerprintOptions *fingerprintOptions
	fingerprint        string // This is calculated or provided by the user
	flowName           string
	envName            string
}

func newAssertArtifactCmd(out io.Writer) *cobra.Command {
	o := &assertArtifactOptions{}
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "artifact [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   assertArtifactShortDesc,
		Long:    assertArtifactLongDesc,
		Example: assertArtifactExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
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
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.fingerprint, "fingerprint", "F", "", fingerprintFlag)
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVar(&o.envName, "environment", "", envNameFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	return cmd
}

func (o *assertArtifactOptions) run(out io.Writer, args []string) error {
	var err error
	if o.fingerprint == "" {
		o.fingerprint, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	}

	baseURL := fmt.Sprintf("%s/api/v2/asserts/%s/fingerprint/%s", global.Host, global.Org, o.fingerprint)
	params := url.Values{}

	if o.flowName != "" {
		params.Add("flow_name", o.flowName)
	}

	if o.envName != "" {
		params.Add("environment_name", o.envName)
	}

	fullURL := baseURL
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    fullURL,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	var evaluationResult map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &evaluationResult)
	if err != nil {
		return err
	}

	scope := evaluationResult["scope"].(string)

	if evaluationResult["compliant"].(bool) {
		logger.Info("COMPLIANT")
		if scope == "flow" {
			logger.Info("See more details at %s", evaluationResult["html_url"].(string))
		}
	} else {
		if scope == "flow" {
			return fmt.Errorf("not compliant\nSee more details at %s", evaluationResult["html_url"].(string))
		} else {
			jsonData, err := json.MarshalIndent(evaluationResult["policy_evaluations"], "", "  ")
			if err != nil {
				return fmt.Errorf("error marshalling evaluation result: %v", err)
			}
			return fmt.Errorf("not compliant for env [%s]: \n %v", o.envName,
				string(jsonData))
		}
	}

	return nil
}
