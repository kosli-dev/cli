package main

import (
	"io"

	"github.com/spf13/cobra"
)

const evaluateTrailShortDesc = `Evaluate a trail against a policy.`

const evaluateTrailLongDesc = evaluateTrailShortDesc + `
Fetch a single trail from Kosli and evaluate it against a Rego policy using OPA.
The trail data is passed to the policy as ` + "`input.trail`" + `.

Use ` + "`--attestations`" + ` to enrich the input with detailed attestation data
(e.g. pull request approvers, scan results). Use ` + "`--show-input`" + ` to inspect the
full data structure available to the policy. Use ` + "`--output json`" + ` for structured output.`

const evaluateTrailExample = `
# evaluate a trail against a policy:
kosli evaluate trail yourTrailName \
	--policy yourPolicyFile.rego \
	--flow yourFlowName \
	--api-token yourAPIToken \
	--org yourOrgName

# evaluate a trail with attestation enrichment:
kosli evaluate trail yourTrailName \
	--policy yourPolicyFile.rego \
	--flow yourFlowName \
	--attestations pull-request \
	--api-token yourAPIToken \
	--org yourOrgName

# evaluate a trail and show the policy input data:
kosli evaluate trail yourTrailName \
	--policy yourPolicyFile.rego \
	--flow yourFlowName \
	--show-input \
	--output json \
	--api-token yourAPIToken \
	--org yourOrgName`

type evaluateTrailOptions struct {
	commonEvaluateOptions
}

func newEvaluateTrailCmd(out io.Writer) *cobra.Command {
	o := new(evaluateTrailOptions)
	cmd := &cobra.Command{
		Use:     "trail TRAIL-NAME",
		Short:   evaluateTrailShortDesc,
		Long:    evaluateTrailLongDesc,
		Example: evaluateTrailExample,
		Args:    cobra.ExactArgs(1),
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
