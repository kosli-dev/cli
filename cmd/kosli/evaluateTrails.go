package main

import (
	"io"

	"github.com/spf13/cobra"
)

const evaluateTrailsShortDesc = `Evaluate multiple trails against a policy.`

const evaluateTrailsLongDesc = evaluateTrailsShortDesc + `
Fetch multiple trails from Kosli and evaluate them together against a Rego policy using OPA.
The trail data is passed to the policy as ` + "`input.trails`" + ` (an array), unlike
` + "`evaluate trail`" + ` which passes ` + "`input.trail`" + ` (a single object).

Use ` + "`--attestations`" + ` to enrich the input with detailed attestation data
(e.g. pull request approvers, scan results). Use ` + "`--show-input`" + ` to inspect the
full data structure available to the policy. Use ` + "`--output json`" + ` for structured output.`

const evaluateTrailsExample = `
# evaluate multiple trails against a policy:
kosli evaluate trails yourTrailName1 yourTrailName2 \
	--policy yourPolicyFile.rego \
	--flow yourFlowName \
	--api-token yourAPIToken \
	--org yourOrgName

# evaluate trails with attestation enrichment:
kosli evaluate trails yourTrailName1 yourTrailName2 \
	--policy yourPolicyFile.rego \
	--flow yourFlowName \
	--attestations pull-request \
	--api-token yourAPIToken \
	--org yourOrgName

# evaluate trails with JSON output and show the policy input:
kosli evaluate trails yourTrailName1 yourTrailName2 \
	--policy yourPolicyFile.rego \
	--flow yourFlowName \
	--show-input \
	--output json \
	--api-token yourAPIToken \
	--org yourOrgName`

type evaluateTrailsOptions struct {
	commonEvaluateOptions
}

func newEvaluateTrailsCmd(out io.Writer) *cobra.Command {
	o := new(evaluateTrailsOptions)
	cmd := &cobra.Command{
		Use:     "trails TRAIL-NAME [TRAIL-NAME...]",
		Short:   evaluateTrailsShortDesc,
		Long:    evaluateTrailsLongDesc,
		Example: evaluateTrailsExample,
		Args:    cobra.MinimumNArgs(1),
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
