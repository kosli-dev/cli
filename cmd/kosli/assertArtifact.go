package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const assertArtifactShortDesc = `Assert the compliance status of an artifact in Kosli (in its flow or against an environment).  `

const assertArtifactLongDesc = assertArtifactShortDesc + `
Exits with non-zero code if the artifact has a non-compliant status.`

const assertArtifactExample = `
# assert that an artifact meets all compliance requirements for an environment
kosli assert artifact \
	--fingerprint 184c799cd551dd1d8d5c5f9a5d593b2e931f5e36122ee5c793c1d08a19839cc0 \
	--flow yourFlowName \
	--environment prod \
	--api-token yourAPIToken \
	--org yourOrgName 

# fail if an artifact has a non-compliant status (using the artifact fingerprint)
kosli assert artifact \
	--fingerprint 184c799cd551dd1d8d5c5f9a5d593b2e931f5e36122ee5c793c1d08a19839cc0 \
	--flow yourFlowName \
	--api-token yourAPIToken \
	--org yourOrgName 

# fail if an artifact has a non-compliant status (using the artifact name and type)
kosli assert artifact library/nginx:1.21 \
	--artifact-type docker \
	--flow yourFlowName \
	--api-token yourAPIToken \
	--org yourOrgName 
`

type assertArtifactOptions struct {
	fingerprintOptions *fingerprintOptions
	fingerprint        string // This is calculated or provided by the user
	flowName           string
	envName            string
	policyNames        []string
	output             string
}

func newAssertArtifactCmd(out io.Writer) *cobra.Command {
	o := &assertArtifactOptions{}
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "artifact [IMAGE-NAME | FILE-PATH | DIR-PATH]",
		Short:   assertArtifactShortDesc,
		Long:    assertArtifactLongDesc,
		Example: assertArtifactExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.fingerprint, false)
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return ValidateRegistryFlags(cmd, o.fingerprintOptions)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.fingerprint, "fingerprint", "F", "", fingerprintFlag)
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVar(&o.envName, "environment", "", envNameFlag)
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().StringSliceVarP(&o.policyNames, "policy", "", []string{}, policyName)

	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)

	return cmd
}

func (o *assertArtifactOptions) run(out io.Writer, args []string) error {
	var err error
	if o.fingerprint == "" {
		o.fingerprint, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	}

	baseURL := fmt.Sprintf("%s/api/v2/asserts/%s/fingerprint/%s", global.Host, global.Org, o.fingerprint)
	params := url.Values{}

	if o.flowName != "" {
		params.Add("flow_name", o.flowName)
	}

	if o.envName != "" {
		params.Add("environment_name", o.envName)
	}

	if len(o.policyNames) > 0 {
		for _, policy := range o.policyNames {
			params.Add("policy_name", policy)
		}
	}

	fullURL := baseURL
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    fullURL,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printAssertAsTable,
			"json":  output.PrintJson,
		})
}

func printAssertAsTable(raw string, out io.Writer, page int) error {
	var evaluationResult map[string]interface{}
	err := json.Unmarshal([]byte(raw), &evaluationResult)
	if err != nil {
		return err
	}

	flow, _ := evaluationResult["flow"].(string)
	trail, _ := evaluationResult["trail"].(string)
	scope := evaluationResult["scope"].(string)
	complianceStatus, _ := evaluationResult["compliance_status"].(map[string]interface{})
	attestationsStatuses, _ := complianceStatus["attestations_statuses"].([]interface{})

	if evaluationResult["compliant"].(bool) {
		logger.Info("COMPLIANT")
	} else {
		logger.Info("Error: NON-COMPLIANT")
	}
	logger.Info("Flow: %v\nTrail %v", flow, trail)
	logger.Info("%-32v %-30v %-15v %-10v", "Attestation-name", "type", "status", "compliant")

	for _, item := range attestationsStatuses {
		attestation := item.(map[string]interface{})
		name := attestation["attestation_name"]
		attType := attestation["attestation_type"]
		status := attestation["status"]
		isCompliant, _ := attestation["is_compliant"].(bool)
		unexpected, _ := attestation["unexpected"].(bool)
		unexpectedStr := ""
		if unexpected {
			unexpectedStr = "unexpected"
		}

		logger.Info("  %-32v %-30v %-15v %-10v %-10v", name, attType, status, isCompliant, unexpectedStr)
	}
	if scope == "environment" || scope == "policy" {
		logger.Info("%-32v %-30v", "Policy-name", "status")
		policyEvaluations := evaluationResult["policy_evaluations"].([]interface{})
		for _, item := range policyEvaluations {
			policyEvaluation := item.(map[string]interface{})
			policyName := policyEvaluation["policy_name"]
			policyStatus := policyEvaluation["status"]
			logger.Info("  %-32v %-30v", policyName, policyStatus)
			if policyStatus != "COMPLIANT" {
				ruleEvaluations := policyEvaluation["rule_evaluations"].([]interface{})
				var failures []string
				for _, item2 := range ruleEvaluations {
					ruleEvaluation := item2.(map[string]interface{})
					ignored := ruleEvaluation["ignored"].(bool)
					satisfied, _ := ruleEvaluation["satisfied"].(bool)
					if !ignored && !satisfied {
						rule := ruleEvaluation["rule"].(map[string]interface{})
						resolutions := ruleEvaluation["resolutions"].([]interface{})
						for _, item3 := range resolutions {
							resolution := item3.(map[string]interface{})
							resolutionType := resolution["type"].(string)
							ruleDefinition := rule["definition"].(map[string]interface{})
							attestationName := ruleDefinition["name"]
							attestationType := ruleDefinition["type"]
							switch resolutionType {
							case "legacy_flow":
								failures = append(failures, "artifact comes from a legacy flow and does not have the new attestations")
							case "missing_attestation":
								failures = append(failures, fmt.Sprintf("artifact is missing required '%v' (type: %v) attestation in trail", attestationName, attestationType))
							case "non_compliant_attestation":
								failures = append(failures, fmt.Sprintf("attestation '%v' is non-compliant in trail", attestationName))
							case "non_compliant_in_trail":
								failures = append(failures, "artifact is not compliant in trail")
							}
						}
					}
				}
				for _, fail := range failures {
					logger.Info("    %v", fail)
				}
			}
		}
	}
	logger.Info("\nSee more details at %s", evaluationResult["html_url"].(string))

	return nil
}
