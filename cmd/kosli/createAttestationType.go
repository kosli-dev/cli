package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const createAttestationTypeShortDesc = `Create or update a Kosli attestation type.`

const createAttestationTypeLongDesc = createAttestationTypeShortDesc + ``

const createAttestationTypeExample = ` `

type createAttestationTypeOptions struct {
	payload CreateAttestationTypePayload
}

type CreateAttestationTypePayload struct {
	TypeName string
}

func newCreateAttestationTypeCmd(out io.Writer) *cobra.Command {
	o := new(createAttestationTypeOptions)
	cmd := &cobra.Command{
		Use:     "attestation-type TYPE-NAME",
		Short:   createAttestationTypeShortDesc,
		Long:    createAttestationTypeLongDesc,
		Example: createAttestationTypeExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	//cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", envDescriptionFlag)

	addDryRunFlag(cmd)
	return cmd
}

func (o *createAttestationTypeOptions) run(args []string) error {
	o.payload.TypeName = args[0]
	url := fmt.Sprintf("%s/api/v2/custom-attestation-types/%s", global.Host, global.Org)

	reqParams := &requests.RequestParams{
		Method:  http.MethodPost,
		URL:     url,
		Payload: o.payload,
		DryRun:  global.DryRun,
		Token:   global.ApiToken,
	}
	_, err := kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("foo bar fix me %s", o.payload.TypeName)
	}
	return err
}
