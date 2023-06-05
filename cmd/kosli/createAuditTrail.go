package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const createAuditTrailShortDesc = `Create or update a Kosli audit trail.`

const createAuditTrailLongDesc = createAuditTrailShortDesc + `
You can specify audit trail parameters in flags.`

const createAuditTrailExample = `
# create/update a Kosli audit trail:
kosli create audit-trail yourAuditTrailName \
	--description yourAuditTrailDescription \
	--steps step1,step2 \
	--api-token yourAPIToken \
	--org yourOrgName
`

type createAuditTrailOptions struct {
	payload AuditTrailPayload
}

type AuditTrailPayload struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Steps       []string `json:"steps"`
}

func newCreateAuditTrailCmd(out io.Writer) *cobra.Command {
	o := new(createAuditTrailOptions)
	cmd := &cobra.Command{
		Use:         "audit-trail AUDIT-TRAIL-NAME",
		Short:       createAuditTrailShortDesc,
		Long:        createAuditTrailLongDesc,
		Example:     createAuditTrailExample,
		Annotations: map[string]string{"betaCLI": "true"},
		Args:        cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = RequireFlags(cmd, []string{"steps"})
			if err != nil {
				logger.Error("failed to configure required flags: %v", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVar(&o.payload.Description, "description", "", flowDescriptionFlag)
	cmd.Flags().StringSliceVarP(&o.payload.Steps, "steps", "s", []string{""}, stepsFlag)
	addDryRunFlag(cmd)

	return cmd
}

func (o *createAuditTrailOptions) run(args []string) error {
	var err error
	url := fmt.Sprintf("%s/api/v2/audit_trails/%s", global.Host, global.Org)

	if o.payload.Name == "" {
		if len(args) == 0 {
			return fmt.Errorf("audit trail name must be provided as an argument")
		}
		o.payload.Name = args[0]
	}

	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Payload:  o.payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("audit trail '%s' was created", o.payload.Name)
	}
	return err
}
