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

func newEvaluateInputCmd(out io.Writer) *cobra.Command {
	o := new(evaluateInputOptions)
	cmd := &cobra.Command{
		Use:   "input",
		Short: "Evaluate a local input file against a policy.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	cmd.Flags().StringVarP(&o.inputFile, "input-file", "i", "", "Path to a JSON input file.")
	cmd.Flags().StringVarP(&o.policyFile, "policy", "p", "", "Path to a Rego policy file.")
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

func loadInputFromFile(filePath string) (map[string]interface{}, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}
	defer f.Close()
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
