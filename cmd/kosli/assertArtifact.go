package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const assertArtifactDesc = `Assert the compliance status of an artifact in Kosli. Exits with non-zero code if the artifact has a non-compliant status.`

type assertArtifactOptions struct {
	fingerprintOptions *fingerprintOptions
	sha256             string // This is calculated or provided by the user
	pipelineName       string
}

func newAssertArtifactCmd(out io.Writer) *cobra.Command {
	o := &assertArtifactOptions{}
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:   "artifact",
		Short: assertArtifactDesc,
		Long:  assertArtifactDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorAfterPrintingHelp(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.sha256, false)
			if err != nil {
				return ErrorAfterPrintingHelp(cmd, err.Error())
			}
			return ValidateRegisteryFlags(cmd, o.fingerprintOptions)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.sha256, "sha256", "s", "", sha256Flag)
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", pipelineNameFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)

	err := RequireFlags(cmd, []string{"pipeline"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *assertArtifactOptions) run(out io.Writer, args []string) error {
	var err error
	if o.sha256 == "" {
		o.sha256, err = GetSha256Digest(args[0], o.fingerprintOptions)
		if err != nil {
			return err
		}
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s", global.Host, global.Owner, o.pipelineName, o.sha256)
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken, global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())
	if err != nil {
		return err
	}

	var artifactData map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &artifactData)
	if err != nil {
		return err
	}

	if artifactData["state"].(string) == "COMPLIANT" {
		_, outErr := out.Write([]byte("artifact is COMPLIANT\n"))
		if outErr != nil {
			return outErr
		}
	} else {
		return fmt.Errorf("artifact is %s", artifactData["state"].(string))
	}

	return nil
}
