package main

import (
	"io"

	"github.com/kosli-dev/cli/internal/evaluations"
	"github.com/spf13/cobra"
)

const evaluatePolicyShortDesc = `Evaluate a policy against a trail in Kosli.`

const evaluatePolicyLongDesc = evaluatePolicyShortDesc + `
The policy is evaluated where the trail is stored, against the trail as Kosli
recorded it, and the verdict is printed here.

Use ` + "`--params`" + ` to pass values the policy reads as ` + "`data.params`" + `.
Use ` + "`--output json`" + ` for structured output.`

const evaluatePolicyExample = `
# evaluate a policy against a trail:
kosli evaluate policy \
	--flow yourFlowName \
	--trail yourTrailName \
	--policy yourPolicyFile.rego \
	--api-token yourAPIToken \
	--org yourOrgName

# evaluate a policy with parameters (inline JSON or @file.json):
kosli evaluate policy \
	--flow yourFlowName \
	--trail yourTrailName \
	--policy yourPolicyFile.rego \
	--params '{"protected_branch": "master"}' \
	--api-token yourAPIToken \
	--org yourOrgName`

type evaluatePolicyOptions struct {
	flowName  string
	trailName string
	policyRef string
	params    string
	output    string
}

func newEvaluatePolicyCmd(out io.Writer) *cobra.Command {
	o := new(evaluatePolicyOptions)
	cmd := &cobra.Command{
		Use:     "policy",
		Short:   evaluatePolicyShortDesc,
		Long:    evaluatePolicyLongDesc,
		Example: evaluatePolicyExample,
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVar(&o.trailName, "trail", "", trailNameFlag)
	cmd.Flags().StringVarP(&o.policyRef, "policy", "p", "", "Path or http(s):// URL of a Rego policy to evaluate the trail against.")
	cmd.Flags().StringVar(&o.params, "params", "", policyParamsFlag)
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)

	err := RequireFlags(cmd, []string{"flow", "trail", "policy"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *evaluatePolicyOptions) run(out io.Writer) error {
	return runServerEvaluation(out, serverEvaluation{
		policyRef: o.policyRef,
		params:    o.params,
		trails:    []evaluations.TrailRef{{Flow: o.flowName, Trail: o.trailName}},
		output:    o.output,
	})
}
