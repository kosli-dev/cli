package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const listAttestationTypesDesc = `List all Kosli attestation types for an org.`

type listAttestationTypesOptions struct {
	output string
}

func newListAttestationTypesCmd(out io.Writer) *cobra.Command {
	o := new(listAttestationTypesOptions)
	cmd := &cobra.Command{
		Use:   "attestation-types",
		Short: listAttestationTypesDesc,
		Long:  listAttestationTypesDesc,
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)

	return cmd
}

func (o *listAttestationTypesOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v2/custom-attestation-types/%s", global.Host, global.Org)

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    url,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printAttestationTypesListAsTable,
			"json":  output.PrintJson,
		})
}

func printAttestationTypesListAsTable(raw string, out io.Writer, page int) error {
	var attestationTypes []map[string]interface{}
	err := json.Unmarshal([]byte(raw), &attestationTypes)
	if err != nil {
		return err
	}

	if len(attestationTypes) == 0 {
		logger.Info("No attestation types were found.")
		return nil
	}

	header := []string{"NAME", "DESCRIPTION", "LATEST VERSION"}
	rows := []string{}

	for _, attestationType := range attestationTypes {
		description := attestationType["description"]
		if description == nil {
			description = ""
		}
		latestVersion := len(attestationType["versions"].([]interface{}))

		rows = append(rows, fmt.Sprintf("%s\t%s\t%d", attestationType["name"], description, latestVersion))
	}

	tabFormattedPrint(out, header, rows)
	return nil
}
