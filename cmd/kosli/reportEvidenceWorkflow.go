package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type reportEvidenceWorkflowOptions struct {
	auditTrailName   string
	userDataFilePath string
	evidencePaths    []string
	payload          WorkflowEvidencePayload
}

const reportEvidenceWorkflowShortDesc = `Report evidence for a workflow in Kosli.`

const reportEvidenceWorkflowExample = `
# report evidence for a workflow:
kosli report evidence workflow \
	--audit-trail auditTrailName \
	--api-token yourAPIToken \
	--id yourID \
	--step step1 \
	--org yourOrgName

# report evidence with a file for a workflow:
kosli report evidence workflow \
	--audit-trail auditTrailName \
	--api-token yourAPIToken \
	--id yourID \
	--step step1 \
	--org yourOrgName \
	--evidence-paths /path/to/your/file
`

func newReportEvidenceWorkflowCmd(out io.Writer) *cobra.Command {
	o := new(reportEvidenceWorkflowOptions)
	cmd := &cobra.Command{
		Use:         "workflow",
		Short:       reportEvidenceWorkflowShortDesc,
		Long:        reportEvidenceWorkflowShortDesc,
		Example:     reportEvidenceWorkflowExample,
		Annotations: map[string]string{"betaCLI": "true"},
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
	cmd.Flags().StringVar(&o.payload.ExternalId, "id", "", workflowIDFlag)
	cmd.Flags().StringVar(&o.payload.Step, "step", "", stepNameFlag)
	cmd.Flags().StringVar(&o.payload.EvidenceURL, "evidence-url", "", evidenceURLFlag)
	cmd.Flags().StringVar(&o.payload.EvidenceFingerprint, "evidence-fingerprint", "", evidenceFingerprintFlag)
	cmd.Flags().StringVarP(&o.userDataFilePath, "user-data", "u", "", evidenceUserDataFlag)
	cmd.Flags().StringSliceVarP(&o.evidencePaths, "evidence-paths", "e", []string{}, evidencePathsFlag)

	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"audit-trail", "id", "step"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *reportEvidenceWorkflowOptions) run(args []string) error {
	var err error
	url := fmt.Sprintf("%s/api/v2/evidence/%s/workflow/%s/generic", global.Host, global.Org, o.auditTrailName)

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
		logger.Info("evidence '%s' for ID '%s' is reported to audit trail: %s", o.payload.Step, o.payload.ExternalId, o.auditTrailName)
	}
	return err
}
