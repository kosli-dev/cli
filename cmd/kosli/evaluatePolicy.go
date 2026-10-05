package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/kosli-dev/cli/internal/digest"
	"github.com/kosli-dev/cli/internal/evaluations"
	"github.com/spf13/cobra"
)

const evaluatePolicyShortDesc = `Evaluate a policy against a trail in Kosli.`

const evaluatePolicyLongDesc = evaluatePolicyShortDesc + `
The policy is evaluated where the trail is stored, against the trail as Kosli
recorded it, and the verdict is printed here.

Name what to evaluate with ` + "`--context trail=<flow>/<trail>`" + `, repeated once per
trail. Trails named in one command are all evaluated at the same instant.

` + "`--policy`" + ` takes a Rego file or a directory, and can be given more than once,
so a policy and a shared library are sent together from wherever each is kept.
Every ` + "`--policy`" + ` joins one bundle. A file is keyed by its own name, and a file
in a directory by its path relative to that directory. Files in dot-directories,
such as ` + "`.git`" + `, are left out; every other file travels, so notes such as
` + "`README.md`" + ` stay with the policy.

Of what travels, the evaluator loads only:
  - Rego modules: ` + "`*.rego`" + `, other than tests (` + "`*_test.rego`" + `). Modules import
    each other by package, so where each file sits does not matter.
  - Data files: ` + "`data.json`" + `, ` + "`data.yaml`" + `, ` + "`data.yml`" + `, ` + "`*.ergo.yaml`" + ` and
    ` + "`*.ergo.yml`" + `. Each is read into ` + "`data`" + ` at its directory's path in the
    bundle, so ` + "`controls/x/data.yaml`" + ` becomes ` + "`data.controls.x`" + `. A data file
    cannot write ` + "`data.params`" + `, which belongs to ` + "`--params`" + `.
Every other file is carried but not read.

Two loaded files under one name are refused, naming both. For any other file the
first ` + "`--policy`" + ` given wins, and the copy left behind is named on stderr.

The policy package is the one that declares ` + "`allow`" + `; every other package is a
library. Where several packages declare ` + "`allow`" + `, annotate an entrypoint in exactly
one of them to say which is the policy. In the policy package:
  - ` + "`allow`" + ` is always evaluated, whether annotated or not. It must be a boolean
    for every input, so give it a default (` + "`default allow := false`" + `): an undefined
    ` + "`allow`" + ` fails the evaluation rather than denying.
  - Every rule annotated with ` + "`# METADATA`" + ` and ` + "`# entrypoint: true`" + ` is evaluated
    too, and printed under its rule name, such as ` + "`violations`" + ` or ` + "`report`" + `. A
    rule undefined for this input is left out. An output named ` + "`input`" + `, ` + "`params`" + `
    or ` + "`decision_attestation_id`" + ` is not printed, as those names are taken.
  - Every other rule is evaluated only as far as those rules need it.
An entrypoint annotated outside the policy package, or on a function, is refused.

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

# evaluate a policy directory with a shared library kept elsewhere:
kosli evaluate policy \
	--context trail=yourFlowName/yourTrailName \
	--policy yourPolicyDirectory \
	--policy path/to/ergo/ergo.rego \
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
	policyRefs  []string
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
		// Hidden while the command is proved out against a server that can run
		// it. Unhiding it is this line, and it also restores its docs page.
		Hidden: true,
	}

	cmd.Flags().StringArrayVar(&o.contexts, "context", []string{}, policyContextFlag)
	cmd.Flags().StringArrayVarP(&o.policyRefs, "policy", "p", []string{}, "Path of a Rego policy file, or of a directory sent as one bundle. Repeat it to send a policy with libraries kept elsewhere.")
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
	// Fetching a policy from a URL is on its way out, so this command does not
	// offer it, though the older evaluate commands still do.
	for _, ref := range o.policyRefs {
		if isRemotePolicyRef(ref) {
			return fmt.Errorf("--policy takes a file or a directory on this machine, not a URL")
		}
	}

	// Refused before the request: a format refused where it is printed would
	// leave a decision recorded, and a rerun would record a second.
	if _, known := evaluatePolicyOutputs[o.output]; !known {
		return fmt.Errorf("unsupported output format: %s. Valid formats are: [table, json]", o.output)
	}

	trails, err := parseTrailContexts(o.contexts)
	if err != nil {
		return err
	}

	decision, err := o.decision()
	if err != nil {
		return err
	}

	return runServerEvaluation(out, serverEvaluation{
		policyRefs:   o.policyRefs,
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
		if given := namedDecisionFlags(o.name, o.fingerprint); given != "" {
			return nil, fmt.Errorf("%s records a decision, so it needs --control", given)
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

	// A malformed fingerprint is ours to catch; one that names no artifact is
	// the API's.
	if o.fingerprint != "" {
		if err := digest.ValidateDigest(o.fingerprint); err != nil {
			return nil, err
		}
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

var evaluatePolicyOutputs = map[string]bool{"table": true, "json": true}

// namedDecisionFlags names the decision flags that were given, so a refusal
// speaks of what was typed.
func namedDecisionFlags(name, fingerprint string) string {
	var given []string
	if name != "" {
		given = append(given, "--name")
	}
	if fingerprint != "" {
		given = append(given, "--fingerprint")
	}
	return strings.Join(given, " and ")
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
