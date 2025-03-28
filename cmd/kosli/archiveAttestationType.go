package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const archiveAttestationTypeShortDesc = `Archive a custom Kosli attestation type.`

const archiveAttestationTypeLongDesc = archiveAttestationTypeShortDesc + `
New custom attestations using this type cannot be made, but existing attestations will still be visible.
`

const archiveAttestationTypeExample = `
# archive a Kosli custom attestation type:
kosli archive attestation-type yourAttestationTypeName \
	--api-token yourAPIToken \
	--org yourOrgName 
`

func newArchiveAttestationTypeCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "attestation-type TYPE-NAME",
		Short:   archiveAttestationTypeShortDesc,
		Long:    archiveAttestationTypeLongDesc,
		Example: archiveAttestationTypeExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			url := fmt.Sprintf("%s/api/v2/custom-attestation-types/%s/%s/archive", global.Host, global.Org, args[0])

			reqParams := &requests.RequestParams{
				Method: http.MethodPut,
				URL:    url,
				DryRun: global.DryRun,
				Token:  global.ApiToken,
			}
			_, err := kosliClient.Do(reqParams)
			if err == nil && !global.DryRun {
				logger.Info("Custom attestation type %s was archived", args[0])
			}
			return err
		},
	}
	addDryRunFlag(cmd)
	return cmd
}
