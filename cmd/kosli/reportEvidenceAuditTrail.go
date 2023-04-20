package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type reportEvidenceAuditTrailOptions struct {
	auditTrailName   string
	userDataFilePath string
	evidencePaths    []string
	payload          AuditTrailEvidencePayload
}

const reportEvidenceAuditTrailShortDesc = `Report an evidence to an audit trail in Kosli.`

const reportEvidenceAuditTrailExample = `
# report an audit trail evidence:
kosli report evidence audit-trail \
	--audit-trail auditTrailName \
	--api-token yourAPIToken \
	--external-id externalID \
	--step step1 \
	--org yourOrgName

# report an audit trail evidence with a file:
kosli report evidence audit-trail \
	--audit-trail auditTrailName \
	--api-token yourAPIToken \
	--external-id externalID \
	--step step1 \
	--org yourOrgName \
	--evidence-paths /path/to/your/file
`

func newReportEvidenceAuditTrailCmd(out io.Writer) *cobra.Command {
	o := new(reportEvidenceAuditTrailOptions)
	cmd := &cobra.Command{
		Use:     "audit-trail",
		Short:   reportEvidenceAuditTrailShortDesc,
		Long:    reportEvidenceAuditTrailShortDesc,
		Example: reportEvidenceAuditTrailExample,
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
	cmd.Flags().StringVar(&o.payload.ExternalId, "external-id", "", externalIdFlag)
	cmd.Flags().StringVar(&o.payload.Step, "step", "", stepNameFlag)
	cmd.Flags().StringVarP(&o.userDataFilePath, "user-data", "u", "", evidenceUserDataFlag)
	cmd.Flags().StringSliceVarP(&o.evidencePaths, "evidence-paths", "e", []string{}, evidencePathsFlag)

	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"audit-trail", "external-id", "step"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *reportEvidenceAuditTrailOptions) run(args []string) error {
	var err error
	// /audit_trails/{org}/{audit_trail_name}/evidence
	url := fmt.Sprintf("%s/api/v2/audit_trails/%s/%s/evidence", global.Host, global.Org, o.auditTrailName)

	o.payload.UserData, err = LoadJsonData(o.userDataFilePath)
	if err != nil {
		return err
	}

	form, cleanupNeeded, evidencePath, err := newEvidenceForm(o.payload, o.evidencePaths)
	// if we created a tar package, remove it after uploading it
	if cleanupNeeded {
		defer os.Remove(evidencePath)
	}

	if err != nil {
		return err
	}

	reqParams := &requests.RequestParams{
		Method:   http.MethodPost,
		URL:      url,
		Form:     form,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("evidence '%s' is reported to audit trail: %s", o.payload.Step, o.auditTrailName)
	}
	return err
}
