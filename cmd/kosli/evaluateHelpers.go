package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

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
}

func (o *commonEvaluateOptions) addFlags(cmd *cobra.Command, policyDesc string) {
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.policyFile, "policy", "p", "", policyDesc)
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().BoolVar(&o.showInput, "show-input", false, "[optional] Include the policy input data in the output.")
	cmd.Flags().StringSliceVar(&o.attestations, "attestations", nil, "[optional] Limit which attestations are included. Plain name for trail-level, dot-qualified (artifact.name) for artifact-level.")

	err := RequireFlags(cmd, []string{"flow", "policy"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}
}

func fetchAndEnrichTrail(flowName, trailName string, attestations []string) (interface{}, error) {
	url := fmt.Sprintf("%s/api/v2/trails/%s/%s/%s", global.Host, global.Org, flowName, trailName)

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    url,
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
			detailURL := fmt.Sprintf("%s/api/v2/attestations/%s?attestation_id=%s", global.Host, global.Org, id)
			detailResp, err := kosliClient.Do(&requests.RequestParams{
				Method: http.MethodGet,
				URL:    detailURL,
				Token:  global.ApiToken,
			})
			if err != nil {
				logger.Debug("failed to fetch attestation detail for %s: %v", id, err)
				continue
			}
			var wrapper map[string]interface{}
			if err := json.Unmarshal([]byte(detailResp.Body), &wrapper); err != nil {
				logger.Debug("failed to parse attestation detail for %s: %v", id, err)
				continue
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

func evaluateAndPrintResult(out io.Writer, policyFile string, input map[string]interface{}, outputFormat string, showInput bool) error {
	policySource, err := os.ReadFile(policyFile)
	if err != nil {
		return fmt.Errorf("failed to read policy file: %w", err)
	}

	result, err := evaluate.Evaluate(string(policySource), input)
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

	raw, err := json.Marshal(auditResult)
	if err != nil {
		return fmt.Errorf("failed to marshal output: %v", err)
	}

	return output.FormattedPrint(string(raw), outputFormat, out, 0,
		map[string]output.FormatOutputFunc{
			"json":  printEvaluateResultAsJson,
			"table": printEvaluateResultAsTable,
		})
}

func printEvaluateResultAsJson(raw string, out io.Writer, _ int) error {
	if err := output.PrintJson(raw, out, 0); err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return err
	}
	if allow, ok := result["allow"].(bool); ok && !allow {
		return fmt.Errorf("policy denied")
	}
	return nil
}

func printEvaluateResultAsTable(raw string, out io.Writer, _ int) error {
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
		return fmt.Errorf("policy denied: %v", violations)
	}
	tabFormattedPrint(out, []string{}, rows)
	return fmt.Errorf("policy denied")
}
