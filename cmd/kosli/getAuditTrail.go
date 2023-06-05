package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const getAuditTrailDesc = `Get the metadata of a specific audit trail.`

type getAuditTrailsOptions struct {
	output string
}

func newGetAuditTrailCmd(out io.Writer) *cobra.Command {
	o := new(getAuditTrailsOptions)
	cmd := &cobra.Command{
		Use:         "audit-trail AUDIT-TRAIL-NAME",
		Short:       getAuditTrailDesc,
		Long:        getAuditTrailDesc,
		Annotations: map[string]string{"betaCLI": "true"},
		Args:        cobra.ExactArgs(1),
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

func (o *getAuditTrailsOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v2/audit_trails/%s/%s", global.Host, global.Org, args[0])

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
			"table": printAuditTrailAsTable,
			"json":  output.PrintJson,
		})
}

func printAuditTrailAsTable(raw string, out io.Writer, page int) error {
	var auditTrail map[string]interface{}
	err := json.Unmarshal([]byte(raw), &auditTrail)
	if err != nil {
		return err
	}

	header := []string{}
	rows := []string{}

	lastModifiedAt, err := formattedTimestamp(auditTrail["last_modified_at"], false)
	if err != nil {
		return err
	}
	createdAt, err := formattedTimestamp(auditTrail["created_at"], false)
	if err != nil {
		return err
	}
	steps := fmt.Sprintf("%s", auditTrail["steps"])
	steps = strings.Replace(steps, " ", ", ", -1)

	rows = append(rows, fmt.Sprintf("Name:\t%s", auditTrail["name"]))
	rows = append(rows, fmt.Sprintf("Description:\t%s", auditTrail["description"]))
	rows = append(rows, fmt.Sprintf("Steps:\t%s", steps))
	rows = append(rows, fmt.Sprintf("Last Modified At:\t%s", lastModifiedAt))
	rows = append(rows, fmt.Sprintf("Created At:\t%s", createdAt))

	tabFormattedPrint(out, header, rows)
	return nil
}
