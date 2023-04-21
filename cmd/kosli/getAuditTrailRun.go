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

const getAuditTrailRunShortDesc = `Get a specific audit trail run for an organization`

const getAuditTrailRunExample = `
# get an audit trail run for an external id
kosli get audit-trail-run externalId \
	--audit-trail auditTrailName \
	--api-token yourAPIToken \
	--org orgName
`

type getAuditTrailRunOptions struct {
	auditTrailName string
	output         string
}

func newGetAuditTrailRunCmd(out io.Writer) *cobra.Command {
	o := new(getAuditTrailRunOptions)
	cmd := &cobra.Command{
		Use:     "audit-trail-run EXTERNAL-ID",
		Short:   getAuditTrailRunShortDesc,
		Long:    getAuditTrailRunShortDesc,
		Example: getAuditTrailRunExample,
		Args:    cobra.ExactArgs(1),
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
	cmd.Flags().StringVar(&o.auditTrailName, "audit-trail", "", auditTrailNameFlag)

	return cmd
}

func (o *getAuditTrailRunOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v2/audit_trails/%s/%s/runs/%s", global.Host, global.Org, o.auditTrailName, args[0])

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
			"table": printAuditTrailRunAsTable,
			"json":  output.PrintJson,
		})
}

func printAuditTrailRunAsTable(raw string, out io.Writer, page int) error {
	var auditTrailRun map[string]interface{}
	err := json.Unmarshal([]byte(raw), &auditTrailRun)
	if err != nil {
		return err
	}

	header := []string{}
	rows := []string{}
	lastModifiedAt, err := formattedTimestamp(auditTrailRun["last_modified_at"], false)
	if err != nil {
		return err
	}
	createdAt, err := formattedTimestamp(auditTrailRun["created_at"], false)
	if err != nil {
		return err
	}
	steps := fmt.Sprintf("%s", auditTrailRun["steps"])
	steps = strings.Replace(steps, " ", ", ", -1)

	evidenceNames := []string{}
	evidence := auditTrailRun["evidence"].([]interface{})
	for _, e := range evidence {
		evidenceNames = append(evidenceNames, fmt.Sprintf("%s", e.(map[string]interface{})["step"]))
	}

	rows = append(rows, fmt.Sprintf("External ID:\t%s", auditTrailRun["external_id"]))
	rows = append(rows, fmt.Sprintf("Audit Trail:\t%s", auditTrailRun["audit_trail_name"]))
	rows = append(rows, fmt.Sprintf("Steps:\t%s", steps))
	rows = append(rows, fmt.Sprintf("Evidence:\t%s", strings.Join(evidenceNames, ", ")))
	rows = append(rows, fmt.Sprintf("Created At:\t%s", createdAt))
	rows = append(rows, fmt.Sprintf("Last Modified At:\t%s", lastModifiedAt))

	tabFormattedPrint(out, header, rows)
	return nil
}
