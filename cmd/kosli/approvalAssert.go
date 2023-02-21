package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const approvalAssertShortDesc = `Assert if an artifact in Kosli has been approved for deployment.`

const approvalAssertLongDesc = approvalAssertShortDesc + `
Exits with non-zero code if artifact has not been approved.
` + fingerprintDesc

const approvalAssertExample = `
# Assert that a file type artifact has been approved
kosli pipeline approval assert FILE.tgz \
	--api-token yourAPIToken \
	--artifact-type file \
	--owner yourOrgName \
	--pipeline yourPipelineName 


# Assert that an artifact with a provided fingerprint (sha256) has been approved
kosli pipeline approval assert \
	--api-token yourAPIToken \
	--owner yourOrgName \
	--pipeline yourPipelineName \
	--sha256 yourSha256
`

type approvalAssertOptions struct {
	fingerprintOptions *fingerprintOptions
	sha256             string
	pipelineName       string
}

func newApprovalAssertCmd(out io.Writer) *cobra.Command {
	o := new(approvalAssertOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "assert [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   approvalAssertShortDesc,
		Long:    approvalAssertLongDesc,
		Example: approvalAssertExample,
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

	cmd.Flags().StringVarP(&o.sha256, "sha256", "s", "", fingerprintFlag)
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", pipelineNameFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"pipeline"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *approvalAssertOptions) run(args []string) error {
	var err error
	if o.sha256 == "" {
		o.sha256, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s/approvals/", global.Host, global.Owner, o.pipelineName, o.sha256)

	reqParams := &requests.RequestParams{
		Method:   http.MethodGet,
		URL:      url,
		Password: global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	var approvals []map[string]interface{}

	err = json.Unmarshal([]byte(response.Body), &approvals)
	if err != nil {
		return err
	}
	if len(approvals) == 0 {
		return fmt.Errorf("artifact with sha256 %s has no approvals created", o.sha256)
	}

	state, ok := approvals[len(approvals)-1]["state"].(string)
	if ok && state == "APPROVED" {
		approvalNumber := approvals[len(approvals)-1]["release_number"]
		logger.Info("artifact with sha256 %s is approved (approval no. [%v])", o.sha256, approvalNumber)
		return nil
	} else {
		return fmt.Errorf("artifact with sha256 %s is not approved", o.sha256)
	}
}
