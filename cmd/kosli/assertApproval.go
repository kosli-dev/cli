package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const assertApprovalShortDesc = `Assert if an artifact in Kosli has been approved for deployment.`

const assertApprovalLongDesc = assertApprovalShortDesc + `
Exits with non-zero code if artifact has not been approved.
` + fingerprintDesc

const assertApprovalExample = `
# Assert that a file type artifact has been approved
kosli assert approval FILE.tgz \
	--api-token yourAPIToken \
	--artifact-type file \
	--owner yourOrgName \
	--flow yourPipelineName 


# Assert that an artifact with a provided fingerprint (sha256) has been approved
kosli asser approval \
	--api-token yourAPIToken \
	--owner yourOrgName \
	--flow yourPipelineName \
	--fingerprint yourFingerprint
`

type assertApprovalOptions struct {
	fingerprintOptions *fingerprintOptions
	sha256             string
	pipelineName       string
}

func newAssertApprovalCmd(out io.Writer) *cobra.Command {
	o := new(assertApprovalOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "approval [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   assertApprovalShortDesc,
		Long:    assertApprovalLongDesc,
		Example: assertApprovalExample,
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

	cmd.Flags().StringVarP(&o.sha256, "fingerprint", "F", "", fingerprintFlag)
	cmd.Flags().StringVarP(&o.pipelineName, "flow", "f", "", pipelineNameFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"flow"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *assertApprovalOptions) run(args []string) error {
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
		return fmt.Errorf("artifact with fingerprint %s has no approvals created", o.sha256)
	}

	state, ok := approvals[len(approvals)-1]["state"].(string)
	if ok && state == "APPROVED" {
		approvalNumber := approvals[len(approvals)-1]["release_number"]
		logger.Info("artifact with fingerprint %s is approved (approval no. [%v])", o.sha256, approvalNumber)
		return nil
	} else {
		return fmt.Errorf("artifact with fingerprint %s is not approved", o.sha256)
	}
}
