package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

const evaluateTrailsDesc = `Evaluate multiple trails against a policy.`

type evaluateTrailsOptions struct {
	flowName     string
	policyFile   string
	output       string
	showInput    bool
	attestations []string
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

	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.policyFile, "policy", "p", "", "Path to a Rego policy file to evaluate against the trails.")
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().BoolVar(&o.showInput, "show-input", false, "[optional] Include the policy input data in the output.")
	cmd.Flags().StringSliceVar(&o.attestations, "attestations", nil, "[optional] Limit which attestations are included. Plain name for trail-level, dot-qualified (artifact.name) for artifact-level.")

	err := RequireFlags(cmd, []string{"flow", "policy"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *evaluateTrailsOptions) run(out io.Writer, args []string) error {
	if o.output != "table" && o.output != "json" {
		return fmt.Errorf("invalid --output value %q: must be one of [table, json]", o.output)
	}

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
