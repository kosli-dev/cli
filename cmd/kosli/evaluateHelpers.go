package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kosli-dev/cli/internal/evaluate"
	"github.com/kosli-dev/cli/internal/requests"
)

type commonEvaluateOptions struct {
	flowName     string
	policyFile   string
	output       string
	showInput    bool
	attestations []string
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
				continue
			}
			var wrapper map[string]interface{}
			if err := json.Unmarshal([]byte(detailResp.Body), &wrapper); err != nil {
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

	if outputFormat == "json" {
		auditResult := map[string]interface{}{
			"allow":      result.Allow,
			"violations": result.Violations,
		}
		if showInput {
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
		if showInput {
			if err := printEvaluateInput(out, input); err != nil {
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

func printEvaluateInput(out io.Writer, input map[string]interface{}) error {
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
