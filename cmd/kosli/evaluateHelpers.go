package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/kosli-dev/cli/internal/evaluate"
	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type commonEvaluateOptions struct {
	flowName     string
	policyFile   string
	output       string
	showInput    bool
	attestations []string
	params       string
	assert       bool
	noAssert     bool
}

func (o *commonEvaluateOptions) addFlags(cmd *cobra.Command, policyDesc string) {
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.policyFile, "policy", "p", "", policyDesc)
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().BoolVar(&o.showInput, "show-input", false, "[optional] Include the policy input data in the output.")
	cmd.Flags().StringSliceVar(&o.attestations, "attestations", nil, "[optional] Limit which attestations are included. Plain name for trail-level, dot-qualified (artifact.name) for artifact-level.")
	cmd.Flags().StringVar(&o.params, "params", "", "[optional] Policy parameters as inline JSON or @file.json. Available in policies as data.params.")
	cmd.Flags().BoolVar(&o.assert, "assert", false, "[optional] Exit with a non-zero status when the policy denies. This is the current default; pass --assert to lock it in across future releases.")
	cmd.Flags().BoolVar(&o.noAssert, "no-assert", false, "[optional] Print the result and always exit 0, even when the policy denies. Use when this command feeds another tool as a policy decision point.")
	cmd.MarkFlagsMutuallyExclusive("assert", "no-assert")
}

// assertOnDeny resolves the --assert / --no-assert pair into a single bool.
// Today the default is true (assert); a future major release flips this by
// returning o.assert directly.
func (o *commonEvaluateOptions) assertOnDeny() bool {
	return !o.noAssert
}

func fetchAndEnrichTrail(flowName, trailName string, attestations []string) (interface{}, error) {
	trailURL, err := url.JoinPath(global.Host, "api/v2/trails", global.Org, flowName, trailName)
	if err != nil {
		return nil, err
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    trailURL,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return nil, err
	}

	var trailData interface{}
	err = json.Unmarshal([]byte(response.Body), &trailData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trail response: %v", err)
	}

	trailData = evaluate.TransformTrail(trailData)
	trailData = evaluate.FilterAttestations(trailData, attestations)

	ids := evaluate.CollectAttestationIDs(trailData)
	if len(ids) > 0 {
		details := make(map[string]interface{})
		for _, id := range ids {
			detailURL, err := url.JoinPath(global.Host, "api/v2/attestations", global.Org)
			if err != nil {
				return nil, err
			}
			q := url.Values{}
			q.Set("attestation_id", id)
			detailURL += "?" + q.Encode()
			detailResp, err := kosliClient.Do(&requests.RequestParams{
				Method: http.MethodGet,
				URL:    detailURL,
				Token:  global.ApiToken,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to fetch attestation detail for %s: %w", id, err)
			}
			var wrapper map[string]interface{}
			if err := json.Unmarshal([]byte(detailResp.Body), &wrapper); err != nil {
				return nil, fmt.Errorf("failed to parse attestation detail for %s: %w", id, err)
			}
			if data, ok := wrapper["data"].([]interface{}); ok && len(data) > 0 {
				if entry, ok := data[0].(map[string]interface{}); ok {
					details[id] = entry
				}
			}
		}
		trailData = evaluate.RehydrateTrail(trailData, details)
	}

	return trailData, nil
}

func parseParams(raw string) (map[string]interface{}, error) {
	if raw == "" {
		return nil, nil
	}

	var jsonBytes []byte
	if strings.HasPrefix(raw, "@") {
		var err error
		jsonBytes, err = os.ReadFile(raw[1:])
		if err != nil {
			return nil, fmt.Errorf("failed to read --params file: %w", err)
		}
	} else {
		jsonBytes = []byte(raw)
	}

	var params map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &params); err != nil {
		return nil, fmt.Errorf("failed to parse --params: %w", err)
	}
	return params, nil
}

func evaluateAndPrintResult(out io.Writer, policyFile string, input map[string]interface{}, outputFormat string, showInput bool, params map[string]interface{}, assertOnDeny bool) error {
	policySource, err := os.ReadFile(policyFile)
	if err != nil {
		return fmt.Errorf("failed to read policy file: %w", err)
	}

	result, err := evaluate.Evaluate(string(policySource), input, params)
	if err != nil {
		return err
	}

	auditResult := map[string]interface{}{
		"allow":      result.Allow,
		"violations": result.Violations,
	}
	if showInput {
		auditResult["input"] = input
	}
	if showInput && params != nil {
		auditResult["params"] = params
	}

	raw, err := json.Marshal(auditResult)
	if err != nil {
		return fmt.Errorf("failed to marshal output: %v", err)
	}

	return output.FormattedPrint(string(raw), outputFormat, out, 0,
		map[string]output.FormatOutputFunc{
			"json":  printEvaluateResultAsJsonFn(assertOnDeny),
			"table": printEvaluateResultAsTableFn(assertOnDeny),
		})
}

func printEvaluateResultAsJsonFn(assertOnDeny bool) output.FormatOutputFunc {
	return func(raw string, out io.Writer, _ int) error {
		if err := output.PrintJson(raw, out, 0); err != nil {
			return err
		}

		var result map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &result); err != nil {
			return err
		}
		if allow, ok := result["allow"].(bool); ok && !allow && assertOnDeny {
			return fmt.Errorf("policy denied")
		}
		return nil
	}
}

func printEvaluateResultAsTableFn(assertOnDeny bool) output.FormatOutputFunc {
	return func(raw string, out io.Writer, _ int) error {
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &result); err != nil {
			return err
		}

		allow, _ := result["allow"].(bool)

		var rows []string
		if allow {
			rows = append(rows, "RESULT:\tALLOWED")
			tabFormattedPrint(out, []string{}, rows)
			return nil
		}

		rows = append(rows, "RESULT:\tDENIED")

		if violations, ok := result["violations"].([]interface{}); ok && len(violations) > 0 {
			for i, v := range violations {
				if i == 0 {
					rows = append(rows, fmt.Sprintf("VIOLATIONS:\t%s", v))
				} else {
					rows = append(rows, fmt.Sprintf("\t%s", v))
				}
			}
			tabFormattedPrint(out, []string{}, rows)
			if assertOnDeny {
				return fmt.Errorf("policy denied: %v", violations)
			}
			return nil
		}
		tabFormattedPrint(out, []string{}, rows)
		if assertOnDeny {
			return fmt.Errorf("policy denied")
		}
		return nil
	}
}
