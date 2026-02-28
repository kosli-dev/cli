package main

import (
	"io"

	"github.com/spf13/cobra"
)

const evaluateTrailDesc = `Evaluate a trail against a policy.`

type evaluateTrailOptions struct {
	commonEvaluateOptions
}

func newEvaluateTrailCmd(out io.Writer) *cobra.Command {
	o := new(evaluateTrailOptions)
	cmd := &cobra.Command{
		Use:   "trail TRAIL-NAME",
		Short: evaluateTrailDesc,
		Long:  evaluateTrailDesc,
		Args:  cobra.ExactArgs(1),
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

	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.policyFile, "policy", "p", "", "Path to a Rego policy file to evaluate against the trail.")
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().BoolVar(&o.showInput, "show-input", false, "[optional] Include the policy input data in the output.")
	cmd.Flags().StringSliceVar(&o.attestations, "attestations", nil, "[optional] Limit which attestations are included. Plain name for trail-level, dot-qualified (artifact.name) for artifact-level.")

	err := RequireFlags(cmd, []string{"flow", "policy"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *evaluateTrailOptions) run(out io.Writer, args []string) error {
	trailData, err := fetchAndEnrichTrail(o.flowName, args[0], o.attestations)
	if err != nil {
		return err
	}

	input := map[string]interface{}{
		"trail": trailData,
	}

	return evaluateAndPrintResult(out, o.policyFile, input, o.output, o.showInput)
}
