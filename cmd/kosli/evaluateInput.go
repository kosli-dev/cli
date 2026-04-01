package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

type evaluateInputOptions struct {
	inputFile  string
	policyFile string
	output     string
	showInput  bool
}

const evaluateInputShortDesc = `Evaluate a local JSON input against a Rego policy.`

const evaluateInputLongDesc = evaluateInputShortDesc + `
Read JSON from a file or stdin and evaluate it against a Rego policy.
The input can contain any JSON structure — the shape is defined by your policy.

The policy must use ` + "`package policy`" + ` and define an ` + "`allow`" + ` rule.
An optional ` + "`violations`" + ` rule (a set of strings) can provide human-readable denial reasons.
The command exits with code 0 when allowed and code 1 when denied.

When ` + "`--input-file`" + ` is omitted, JSON is read from stdin.`

const evaluateInputExample = `
# evaluate a local JSON file against a policy:
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
	--policy policy.rego`

func newEvaluateInputCmd(out io.Writer) *cobra.Command {
	o := new(evaluateInputOptions)
	cmd := &cobra.Command{
		Use:     "input",
		Short:   evaluateInputShortDesc,
		Long:    evaluateInputLongDesc,
		Example: evaluateInputExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	cmd.Flags().StringVarP(&o.inputFile, "input-file", "i", "", "[optional] Path to a JSON input file. Reads from stdin if omitted.")
	cmd.Flags().StringVarP(&o.policyFile, "policy", "p", "", "Path to a Rego policy file to evaluate against the input.")
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().BoolVar(&o.showInput, "show-input", false, "[optional] Include the policy input data in the output.")

	err := RequireFlags(cmd, []string{"policy"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *evaluateInputOptions) run(out io.Writer) error {
	var input map[string]interface{}
	var err error

	if o.inputFile == "" {
		input, err = loadInput(os.Stdin)
	} else {
		input, err = loadInputFromFile(o.inputFile)
	}
	if err != nil {
		return err
	}

	return evaluateAndPrintResult(out, o.policyFile, input, o.output, o.showInput)
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
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}
	var input map[string]interface{}
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("failed to parse input: %w", err)
	}
	return input, nil
}
