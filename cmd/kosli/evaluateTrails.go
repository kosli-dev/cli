package main

import (
	"io"

	"github.com/spf13/cobra"
)

const evaluateTrailsDesc = `Evaluate multiple trails against a policy.`

type evaluateTrailsOptions struct {
	commonEvaluateOptions
}

func newEvaluateTrailsCmd(out io.Writer) *cobra.Command {
	o := new(evaluateTrailsOptions)
	cmd := &cobra.Command{
		Use:   "trails TRAIL-NAME [TRAIL-NAME...]",
		Short: evaluateTrailsDesc,
		Long:  evaluateTrailsDesc,
		Args:  cobra.MinimumNArgs(1),
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

	o.addFlags(cmd, "Path to a Rego policy file to evaluate against the trails.")

	return cmd
}

func (o *evaluateTrailsOptions) run(out io.Writer, args []string) error {
	var trails []interface{}
	for _, trailName := range args {
		trailData, err := fetchAndEnrichTrail(o.flowName, trailName, o.attestations)
		if err != nil {
			return err
		}
		trails = append(trails, trailData)
	}

	input := map[string]interface{}{
		"trails": trails,
	}

	return evaluateAndPrintResult(out, o.policyFile, input, o.output, o.showInput)
}
