package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/kosli-dev/cli/internal/evaluations"
	"github.com/spf13/cobra"
)

const evaluatePolicyShortDesc = `Evaluate a policy against a trail in Kosli.`

const evaluatePolicyLongDesc = evaluatePolicyShortDesc + `
The policy is evaluated where the trail is stored, against the trail as Kosli
recorded it, and the verdict is printed here.

Name what to evaluate with ` + "`--context trail=<flow>/<trail>`" + `, repeated once per
trail. Trails named in one command are all evaluated at the same instant.

Pass ` + "`--control`" + ` to record the outcome as a decision against that control, in the
` + "`--flow`" + ` and ` + "`--trail`" + ` given. The decision is recorded where the policy runs, so
the verdict is never asserted from here. Without ` + "`--control`" + ` nothing is recorded.

Use ` + "`--params`" + ` to pass values the policy reads as ` + "`data.params`" + `.
Pass ` + "`--assert`" + ` to exit with a non-zero status when the policy denies.
Use ` + "`--output json`" + ` for structured output.`

const evaluatePolicyExample = `
# evaluate a policy against a trail:
kosli evaluate policy \
	--context trail=yourFlowName/yourTrailName \
	--policy yourPolicyFile.rego \
	--api-token yourAPIToken \
	--org yourOrgName

# evaluate several trails at one instant:
kosli evaluate policy \
	--context trail=yourFlowName/yourTrailName \
	--context trail=anotherFlowName/anotherTrailName \
	--policy yourPolicyFile.rego \
	--api-token yourAPIToken \
	--org yourOrgName

# evaluate a policy with parameters (inline JSON or @file.json):
kosli evaluate policy \
	--context trail=yourFlowName/yourTrailName \
	--policy yourPolicyFile.rego \
	--params '{"protected_branch": "master"}' \
	--api-token yourAPIToken \
	--org yourOrgName

# evaluate a policy and record the outcome as a decision:
kosli evaluate policy \
	--context trail=yourFlowName/yourTrailName \
	--policy yourPolicyFile.rego \
	--control yourControlIdentifier \
	--flow yourFlowName \
	--trail yourTrailName \
	--fingerprint yourArtifactFingerprint \
	--api-token yourAPIToken \
	--org yourOrgName

# evaluate a policy and fail the step when it denies:
kosli evaluate policy \
	--context trail=yourFlowName/yourTrailName \
	--policy yourPolicyFile.rego \
	--assert \
	--api-token yourAPIToken \
	--org yourOrgName`

type evaluatePolicyOptions struct {
	contexts    []string
	policyRef   string
	params      string
	output      string
	assert      bool
	control     string
	flowName    string
	trailName   string
	name        string
	fingerprint string
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

	cmd.Flags().StringArrayVar(&o.contexts, "context", []string{}, policyContextFlag)
	cmd.Flags().StringVarP(&o.policyRef, "policy", "p", "", "Path or http(s):// URL of a Rego policy to evaluate the trail against.")
	cmd.Flags().StringVar(&o.params, "params", "", policyParamsFlag)
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().BoolVar(&o.assert, "assert", false, policyAssertFlag)
	cmd.Flags().StringVar(&o.control, "control", "", policyControlFlag)
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", policyDecisionFlowFlag)
	cmd.Flags().StringVar(&o.trailName, "trail", "", policyDecisionTrailFlag)
	cmd.Flags().StringVar(&o.name, "name", "", policyDecisionNameFlag)
	cmd.Flags().StringVar(&o.fingerprint, "fingerprint", "", policyDecisionFingerprintFlag)

	err := RequireFlags(cmd, []string{"context", "policy"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *evaluatePolicyOptions) run(out io.Writer) error {
	trails, err := parseTrailContexts(o.contexts)
	if err != nil {
		return err
	}

	decision, err := o.decision()
	if err != nil {
		return err
	}

	return runServerEvaluation(out, serverEvaluation{
		policyRef:    o.policyRef,
		params:       o.params,
		trails:       trails,
		decision:     decision,
		output:       o.output,
		assertOnDeny: o.assert,
	})
}

// decision resolves where the outcome is recorded, or nil where none was
// asked for. The flags are read as resolved values rather than as flags the
// caller typed, so KOSLI_FLOW and KOSLI_TRAIL satisfy the destination too.
func (o *evaluatePolicyOptions) decision() (*evaluations.Decision, error) {
	if o.control == "" {
		// Accepting these silently would look like a decision was recorded.
		if o.name != "" || o.fingerprint != "" {
			return nil, fmt.Errorf(
				"--name and --fingerprint record a decision, so they need --control")
		}
		// --flow and --trail are not refused with them: a pipeline sets those
		// for every command it runs, and without a control they name nothing.
		return nil, nil
	}

	var missing []string
	if o.flowName == "" {
		missing = append(missing, "--flow")
	}
	if o.trailName == "" {
		missing = append(missing, "--trail")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf(
			"a decision is recorded in a trail, so --control needs %s "+
				"(set as flags, or as KOSLI_FLOW and KOSLI_TRAIL)",
			strings.Join(missing, " and "))
	}

	name := o.name
	if name == "" {
		name = o.control + "-decision"
	}

	return &evaluations.Decision{
		Control:     o.control,
		Name:        name,
		Flow:        o.flowName,
		Trail:       o.trailName,
		Fingerprint: o.fingerprint,
	}, nil
}

// parseTrailContexts reads the --context values as trail references, keeping
// the order they were given in. `trail` is the only kind of context there is
// today; the key is spelled out so another kind can be added without a second
// flag.
func parseTrailContexts(values []string) ([]evaluations.TrailRef, error) {
	trails := make([]evaluations.TrailRef, 0, len(values))
	for _, value := range values {
		kind, reference, found := strings.Cut(value, "=")
		if !found || kind != "trail" {
			return nil, fmt.Errorf("--context %q is not a context this command knows; "+
				"expected trail=<flow>/<trail>", value)
		}
		flow, trail, found := strings.Cut(reference, "/")
		if !found || flow == "" || trail == "" || strings.Contains(trail, "/") {
			return nil, fmt.Errorf("--context %q does not name a trail; "+
				"expected trail=<flow>/<trail>", value)
		}
		trails = append(trails, evaluations.TrailRef{Flow: flow, Trail: trail})
	}
	return trails, nil
}
