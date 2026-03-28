package main

import (
	"io"

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

	err := RequireFlags(cmd, []string{"input-file", "policy"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *evaluateInputOptions) run(out io.Writer) error {
	return nil
}
