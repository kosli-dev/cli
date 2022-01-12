package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

type approvalAssertOptions struct {
	fingerprintOptions *fingerprintOptions
	sha256             string
	pipelineName       string
}

func newApprovalAssertCmd(out io.Writer) *cobra.Command {
	o := new(approvalAssertOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:   "assert ARTIFACT-NAME-OR-PATH",
		Short: "Assert if an artifact in Merkely has been approved for deployment.",
		Long:  approvalAssertDesc(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return err
			}

			return ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.sha256)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVarP(&o.sha256, "sha256", "s", "", "The SHA256 fingerprint for the artifact to be approved. Only required if you don't specify --type.")
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", "The Merkely pipeline name.")
	addFingerprintFlags(cmd, o.fingerprintOptions)

	err := RequireFlags(cmd, []string{"pipeline"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *approvalAssertOptions) run(args []string) error {
	var err error
	if o.sha256 == "" {
		o.sha256, err = GetSha256Digest(args[0], o.fingerprintOptions)
		if err != nil {
			return err
		}
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s/approvals/", global.Host, global.Owner, o.pipelineName, o.sha256)

	response, err := requests.SendPayload([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodGet, log)

	if err != nil && !global.DryRun {
		return err
	}

	var approvals []map[string]interface{}
	if !global.DryRun {
		err = json.Unmarshal([]byte(response.Body), &approvals)
		if err != nil {
			return err
		}
		if len(approvals) == 0 {
			return fmt.Errorf("artifact with sha256 %s has no approvals created", o.sha256)
		}

		state, ok := approvals[len(approvals)-1]["state"].(string)
		if ok && state == "APPROVED" {
			log.Infof("artifact with sha256 %s is approved in approval no. [%d]", o.sha256, len(approvals))
			return nil
		} else {
			return fmt.Errorf("artifact with sha256 %s is not approved", o.sha256)
		}
	} else {
		return nil
	}
}

func approvalAssertDesc() string {
	return `Assert if an artifact in Merkely has been approved for deployment.
   The artifact SHA256 fingerprint is calculated or alternatively it can be provided directly. 
   `
}
