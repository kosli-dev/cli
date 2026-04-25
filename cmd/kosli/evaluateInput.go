package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type evaluateInputOptions struct {
	commonEvaluateOptions
	inputFile string
}

const evaluateInputShortDesc = `[BETA] Evaluate a local JSON input against a Rego policy.`

const evaluateInputLongDesc = evaluateInputShortDesc + `
Read JSON from a file or stdin and evaluate it against a Rego policy.
The input file should contain the raw JSON object your policy expects —
not the wrapper produced by ` + "`--show-input`" + `. Use ` + "`jq '.input'`" + ` to extract
the policy input from a ` + "`--show-input --output json`" + ` capture.

The policy must use ` + "`package policy`" + ` and define an ` + "`allow`" + ` rule.
An optional ` + "`violations`" + ` rule (a set of strings) can provide human-readable denial reasons.

By default a deny exits with code 1. Pass ` + "`--no-assert`" + ` to print the verdict
and exit 0 even on deny, when this command is feeding another tool as a
policy decision point.

When ` + "`--input-file`" + ` is omitted, JSON is read from stdin.

Use ` + "`--params`" + ` to pass configuration data to the policy as ` + "`data.params`" + `.
This accepts inline JSON or a file reference (` + "`@file.json`" + `).`

const evaluateInputExample = `
# capture trail data for local policy iteration:
kosli evaluate trail TRAIL --flow FLOW \
	--policy allow-all.rego \
	--show-input --output json | jq '.input' > trail-data.json

# then iterate on your policy locally:
kosli evaluate input \
	--input-file trail-data.json \
	--policy policy.rego

# evaluate and show the data passed to the policy:
kosli evaluate input \
	--input-file trail-data.json \
	--policy policy.rego \
	--show-input \
	--output json

# read input from stdin:
cat trail-data.json | kosli evaluate input \
	--policy policy.rego

# evaluate with policy parameters (inline JSON):
kosli evaluate input \
	--input-file trail-data.json \
	--policy policy.rego \
	--params '{"threshold": 3}'

# evaluate with policy parameters from a file:
kosli evaluate input \
	--input-file trail-data.json \
	--policy policy.rego \
	--params @params.json

# evaluate as a decision point (print verdict, never fail the step):
kosli evaluate input \
	--input-file trail-data.json \
	--policy policy.rego \
	--no-assert`

func newEvaluateInputCmd(out io.Writer) *cobra.Command {
	o := new(evaluateInputOptions)
	cmd := &cobra.Command{
		Use:     "input",
		Short:   evaluateInputShortDesc,
		Long:    evaluateInputLongDesc,
		Example: evaluateInputExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, cmd.InOrStdin())
		},
	}

	o.addFlags(cmd, "Path to a Rego policy file to evaluate against the input.")
	cmd.Flags().StringVarP(&o.inputFile, "input-file", "i", "", "[optional] Path to a JSON input file. Reads from stdin if omitted.")

	cmd.Flags().Lookup("flow").Hidden = true
	cmd.Flags().Lookup("attestations").Hidden = true

	err := RequireFlags(cmd, []string{"policy"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *evaluateInputOptions) run(out io.Writer, in io.Reader) error {
	var input map[string]interface{}
	var err error

	if o.inputFile == "" {
		if f, ok := in.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
			return fmt.Errorf("no input provided: use --input-file or pipe JSON to stdin")
		}
		input, err = loadInput(in)
	} else {
		input, err = loadInputFromFile(o.inputFile)
	}
	if err != nil {
		return err
	}

	params, err := parseParams(o.params)
	if err != nil {
		return err
	}

	return evaluateAndPrintResult(out, o.policyFile, input, o.output, o.showInput, params, o.assertOnDeny())
}

func loadInputFromFile(filePath string) (result map[string]interface{}, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	return loadInput(f)
}

func loadInput(r io.Reader) (map[string]interface{}, error) {
	var input map[string]interface{}
	if err := json.NewDecoder(r).Decode(&input); err != nil {
		return nil, fmt.Errorf("failed to parse input: %w", err)
	}
	return input, nil
}
