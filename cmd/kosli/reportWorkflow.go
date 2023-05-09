package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type reportWorkflowOptions struct {
	auditTrailName string
	externalId     string
}

const reportWorkflowShortDesc = `Report a workflow creation to a Kosli audit-trail.`

const reportWorkflowLongDesc = reportWorkflowShortDesc

const reportWorkflowExample = `
# Report to a Kosli audit-trail that a workflow has been created
kosli report workflow \
	--audit-trail auditTrailName \
	--api-token yourAPIToken \
	--id yourID \
	--org yourOrgName
`

func newReportWorkflowCmd(out io.Writer) *cobra.Command {
	o := new(reportWorkflowOptions)
	cmd := &cobra.Command{
		Use:     "workflow",
		Short:   reportWorkflowShortDesc,
		Long:    reportWorkflowLongDesc,
		Example: reportWorkflowExample,
		Hidden:  true,
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

	cmd.Flags().StringVar(&o.auditTrailName, "audit-trail", "", auditTrailNameFlag)
	cmd.Flags().StringVar(&o.externalId, "id", "", workflowIDFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"audit-trail", "id"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *reportWorkflowOptions) run(args []string) error {
	url := fmt.Sprintf("%s/api/v2/workflows/%s/%s/%s", global.Host, global.Org, o.auditTrailName, o.externalId)

	reqParams := &requests.RequestParams{
		Method:   http.MethodPost,
		URL:      url,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err := kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("workflow was created in audit-trail '%s' with ID '%s'", o.auditTrailName, o.externalId)
	}
	return err
}
