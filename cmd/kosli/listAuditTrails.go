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

const listAuditTrailsDesc = `List audit trails for an org.`

type listAuditTrailsOptions struct {
	output string
}

func newListAuditTrailsCmd(out io.Writer) *cobra.Command {
	o := new(listAuditTrailsOptions)
	cmd := &cobra.Command{
		Use:         "audit-trails",
		Short:       listAuditTrailsDesc,
		Long:        listAuditTrailsDesc,
		Annotations: map[string]string{"betaCLI": "true"},
		Args:        cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)

	return cmd
}

func (o *listAuditTrailsOptions) run(out io.Writer) error {
	url := fmt.Sprintf("%s/api/v2/audit_trails/%s", global.Host, global.Org)

	reqParams := &requests.RequestParams{
		Method:   http.MethodGet,
		URL:      url,
		Password: global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printAuditTrailsListAsTable,
			"json":  output.PrintJson,
		})
}

func printAuditTrailsListAsTable(raw string, out io.Writer, page int) error {
	var auditTrails []map[string]interface{}
	err := json.Unmarshal([]byte(raw), &auditTrails)
	if err != nil {
		return err
	}

	if len(auditTrails) == 0 {
		logger.Info("No audit trails were found.")
		return nil
	}

	header := []string{"NAME", "DESCRIPTION", "STEPS"}
	rows := []string{}
	for _, auditTrail := range auditTrails {
		row := fmt.Sprintf("%s\t%s\t%s", auditTrail["name"], auditTrail["description"], auditTrail["steps"])
		rows = append(rows, row)
	}
	tabFormattedPrint(out, header, rows)

	return nil
}
