package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/kosli"
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
			return kosliAPI().Send(func(h kosli.Host) error {
				url, err := url.JoinPath(h.URL, "api/v2/custom-attestation-types", global.Org, args[0], "archive")
				if err != nil {
					return err
				}

				reqParams := &requests.RequestParams{
					Method: http.MethodPut,
					URL:    url,
					DryRun: h.DryRun,
					Token:  h.Token,
				}
				_, err = h.Do(reqParams)
				if err == nil && !h.DryRun {
					h.Info("Custom attestation type %s was archived", args[0])
				}
				return err
			})
		},
	}
	addDryRunFlag(cmd)
	return cmd
}
