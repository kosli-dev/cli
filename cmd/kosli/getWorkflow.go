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

const getWorkflowShortDesc = `Get a specific workflow for an organization`

const getWorkflowExample = `
# get workflow for an ID
kosli get workflow yourID \
	--audit-trail auditTrailName \
	--api-token yourAPIToken \
	--org orgName
`

type getWorkflowOptions struct {
	auditTrailName string
	output         string
}

func newGetWorkflowCmd(out io.Writer) *cobra.Command {
	o := new(getWorkflowOptions)
	cmd := &cobra.Command{
		Use:         "workflow ID",
		Short:       getWorkflowShortDesc,
		Long:        getWorkflowShortDesc,
		Example:     getWorkflowExample,
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
	cmd.Flags().StringVar(&o.auditTrailName, "audit-trail", "", auditTrailNameFlag)

	return cmd
}

func (o *getWorkflowOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v2/workflows/%s/%s/%s", global.Host, global.Org, o.auditTrailName, args[0])

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
			"table": printWorkflowAsTable,
			"json":  output.PrintJson,
		})
}

func printWorkflowAsTable(raw string, out io.Writer, page int) error {
	var workflow map[string]interface{}
	err := json.Unmarshal([]byte(raw), &workflow)
	if err != nil {
		return err
	}

	header := []string{}
	rows := []string{}
	lastModifiedAt, err := formattedTimestamp(workflow["last_modified_at"], false)
	if err != nil {
		return err
	}
	createdAt, err := formattedTimestamp(workflow["created_at"], false)
	if err != nil {
		return err
	}
	steps := fmt.Sprintf("%s", workflow["steps"])
	steps = strings.Replace(steps, " ", ", ", -1)

	evidence := workflow["evidence"].(map[string]interface{})
	evidenceNames := make([]string, 0, len(evidence))
	for name := range evidence {
		evidenceNames = append(evidenceNames, name)
	}

	rows = append(rows, fmt.Sprintf("ID:\t%s", workflow["id"]))
	rows = append(rows, fmt.Sprintf("Audit Trail:\t%s", workflow["audit_trail_name"]))
	rows = append(rows, fmt.Sprintf("Steps:\t%s", steps))
	rows = append(rows, fmt.Sprintf("Evidence:\t%s", strings.Join(evidenceNames, ", ")))
	rows = append(rows, fmt.Sprintf("Created At:\t%s", createdAt))
	rows = append(rows, fmt.Sprintf("Last Modified At:\t%s", lastModifiedAt))

	tabFormattedPrint(out, header, rows)
	return nil
}
