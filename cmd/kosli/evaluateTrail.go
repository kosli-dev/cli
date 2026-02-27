package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kosli-dev/cli/internal/evaluate"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const evaluateTrailDesc = `Evaluate a trail against a policy.`

type evaluateTrailOptions struct {
	flowName   string
	policyFile string
	format     string
	showInput  bool
}

func newEvaluateTrailCmd(out io.Writer) *cobra.Command {
	o := new(evaluateTrailOptions)
	cmd := &cobra.Command{
		Use:   "trail TRAIL-NAME",
		Short: evaluateTrailDesc,
		Long:  evaluateTrailDesc,
		Args:  cobra.ExactArgs(1),
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
	cmd.Flags().StringVarP(&o.policyFile, "policy", "p", "", "[optional] Path to a Rego policy file to evaluate against the trail.")
	cmd.Flags().StringVar(&o.format, "format", "text", "[defaulted] The format of the policy evaluation output. Valid formats are: [text, json].")
	cmd.Flags().BoolVar(&o.showInput, "show-input", false, "[optional] Include the policy input data in the output.")

	err := RequireFlags(cmd, []string{"flow"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *evaluateTrailOptions) run(out io.Writer, args []string) error {
	if o.format != "text" && o.format != "json" {
		return fmt.Errorf("invalid --format value %q: must be one of [text, json]", o.format)
	}

	url := fmt.Sprintf("%s/api/v2/trails/%s/%s/%s", global.Host, global.Org, o.flowName, args[0])

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    url,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	var trailData interface{}
	err = json.Unmarshal([]byte(response.Body), &trailData)
	if err != nil {
		return fmt.Errorf("failed to parse trail response: %v", err)
	}

	input := map[string]interface{}{
		"trail": trailData,
	}

	if o.policyFile == "" {
		output, err := json.MarshalIndent(input, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal output: %v", err)
		}
		_, err = fmt.Fprintln(out, string(output))
		return err
	}

	policySource, err := os.ReadFile(o.policyFile)
	if err != nil {
		return fmt.Errorf("failed to read policy file: %w", err)
	}

	result, err := evaluate.Evaluate(string(policySource), input)
	if err != nil {
		return err
	}

	if o.format == "json" {
		auditResult := map[string]interface{}{
			"allow":      result.Allow,
			"violations": result.Violations,
		}
		if o.showInput {
			auditResult["input"] = input
		}
		output, err := json.MarshalIndent(auditResult, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal output: %v", err)
		}
		_, err = fmt.Fprintln(out, string(output))
		if err != nil {
			return err
		}
		if !result.Allow {
			return fmt.Errorf("policy denied")
		}
		return nil
	}

	if result.Allow {
		_, err = fmt.Fprintln(out, "Policy evaluation: ALLOWED")
		if err != nil {
			return err
		}
		if o.showInput {
			if err := o.printInput(out, input); err != nil {
				return err
			}
		}
		return nil
	}

	_, err = fmt.Fprintln(out, "Policy evaluation: DENIED")
	if err != nil {
		return err
	}
	if len(result.Violations) > 0 {
		_, err = fmt.Fprintln(out, "Violations:")
		if err != nil {
			return err
		}
		for _, v := range result.Violations {
			_, err = fmt.Fprintf(out, "  - %s\n", v)
			if err != nil {
				return err
			}
		}
		return fmt.Errorf("policy denied: %v", result.Violations)
	}
	return fmt.Errorf("policy denied")
}

func (o *evaluateTrailOptions) printInput(out io.Writer, input map[string]interface{}) error {
	_, err := fmt.Fprintln(out, "Input:")
	if err != nil {
		return err
	}
	inputJSON, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal input: %v", err)
	}
	_, err = fmt.Fprintln(out, string(inputJSON))
	return err
}
