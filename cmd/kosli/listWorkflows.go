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

const listWorkflowsDesc = `List workflows for an audit trail.`

type listWorkflowsOptions struct {
	listOptions
	auditTrailName string
}

func newListWorkflowsCmd(out io.Writer) *cobra.Command {
	o := new(listWorkflowsOptions)
	cmd := &cobra.Command{
		Use:    "workflows",
		Hidden: true,
		Short:  listWorkflowsDesc,
		Long:   listWorkflowsDesc,
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

func (o *listWorkflowsOptions) run(out io.Writer) error {
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
			"table": printWorkflowsListAsTable,
			"json":  output.PrintJson,
		})
}

func printWorkflowsListAsTable(raw string, out io.Writer, page int) error {
	var workflows []map[string]interface{}
	err := json.Unmarshal([]byte(raw), &workflows)
	if err != nil {
		return err
	}

	if len(workflows) == 0 {
		msg := "No workflows were found"
		if page != 1 {
			msg = fmt.Sprintf("%s at page number %d", msg, page)
		}
		logger.Info(msg + ".")
		return nil
	}

	header := []string{"EXTERNAL_ID", "EVIDENCE", "CREATED_AT", "LAST_MODIFIED_AT"}
	rows := []string{}
	for _, workflow := range workflows {
		externalId := workflow["external_id"].(string)
		evidenceNames := []string{}
		evidence := workflow["evidence"].([]interface{})
		for _, evidenceItem := range evidence {
			evidenceItemMap := evidenceItem.(map[string]interface{})
			evidenceItemStep := evidenceItemMap["step"].(string)
			evidenceNames = append(evidenceNames, evidenceItemStep)
		}
		createdAt, err := formattedTimestamp(workflow["created_at"], true)
		if err != nil {
			return err
		}
		lastModifiedAt, err := formattedTimestamp(workflow["last_modified_at"], true)
		if err != nil {
			return err
		}

		row := fmt.Sprintf("%s\t%s\t%s\t%s", externalId, evidenceNames, createdAt, lastModifiedAt)
		rows = append(rows, row)
	}
	tabFormattedPrint(out, header, rows)

	return nil
}
