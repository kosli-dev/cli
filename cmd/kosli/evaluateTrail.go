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

	o.addFlags(cmd, "Path to a Rego policy file to evaluate against the trail.")

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
