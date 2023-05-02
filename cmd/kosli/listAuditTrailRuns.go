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

const listAuditTrailRunsDesc = `List audit trail runs for an audit trail.`

type listAuditTrailRunsOptions struct {
	listOptions
	auditTrailName string
}

func newListAuditTrailRunsCmd(out io.Writer) *cobra.Command {
	o := new(listAuditTrailRunsOptions)
	cmd := &cobra.Command{
		Use:    "audit-trail-runs",
		Hidden: true,
		Short:  listAuditTrailRunsDesc,
		Long:   listAuditTrailRunsDesc,
		Args:   cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return o.validate(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	cmd.Flags().StringVar(&o.auditTrailName, "audit-trail", "", auditTrailNameFlag)
	addListFlags(cmd, &o.listOptions)

	err := RequireFlags(cmd, []string{"audit-trail"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *listAuditTrailRunsOptions) run(out io.Writer) error {
	url := fmt.Sprintf("%s/api/v2/workflows/%s/%s/workflows?page=%d&per_page=%d",
		global.Host, global.Org, o.auditTrailName, o.pageNumber, o.pageLimit)

	reqParams := &requests.RequestParams{
		Method:   http.MethodGet,
		URL:      url,
		Password: global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, o.pageNumber,
		map[string]output.FormatOutputFunc{
			"table": printAuditTrailRunsListAsTable,
			"json":  output.PrintJson,
		})
}

func printAuditTrailRunsListAsTable(raw string, out io.Writer, page int) error {
	var auditTrailRuns []map[string]interface{}
	err := json.Unmarshal([]byte(raw), &auditTrailRuns)
	if err != nil {
		return err
	}

	if len(auditTrailRuns) == 0 {
		msg := "No audit trial runs were found"
		if page != 1 {
			msg = fmt.Sprintf("%s at page number %d", msg, page)
		}
		logger.Info(msg + ".")
		return nil
	}

	header := []string{"EXTERNAL_ID", "EVIDENCE", "CREATED_AT", "LAST_MODIFIED_AT"}
	rows := []string{}
	for _, auditTrailRun := range auditTrailRuns {
		externalId := auditTrailRun["external_id"].(string)
		evidenceNames := []string{}
		evidence := auditTrailRun["evidence"].([]interface{})
		for _, evidenceItem := range evidence {
			evidenceItemMap := evidenceItem.(map[string]interface{})
			evidenceItemStep := evidenceItemMap["step"].(string)
			evidenceNames = append(evidenceNames, evidenceItemStep)
		}
		createdAt, err := formattedTimestamp(auditTrailRun["created_at"], true)
		if err != nil {
			return err
		}
		lastModifiedAt, err := formattedTimestamp(auditTrailRun["last_modified_at"], true)
		if err != nil {
			return err
		}

		row := fmt.Sprintf("%s\t%s\t%s\t%s", externalId, evidenceNames, createdAt, lastModifiedAt)
		rows = append(rows, row)
	}
	tabFormattedPrint(out, header, rows)

	return nil
}
