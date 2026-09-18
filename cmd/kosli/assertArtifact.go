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

const assertArtifactShortDesc = `Assert the compliance status of an artifact in Kosli. ` +
	`
There are three ways to choose what to assert against:

1. Against an environment. When ^--environment^ is specified,
asserts against all policies currently attached to the given environment.
2. Against one or more policies. When ^--policy^ is specified,
asserts against all the given policies.
3. Against flow templates. When neither ^--environment^ nor ^--policy^
is specified, asserts against the template files of the flows the artifact
is found in.

^--environment^ and ^--policy^ are mutually exclusive.

^--flow^ can be combined with any of the above to narrow the lookup
to a specific flow. Without ^--flow^, all flows containing the artifact
(by fingerprint) are considered.
`

const assertArtifactLongDesc = assertArtifactShortDesc + `
Exits with zero code if the artifact has compliant status,
non-zero code if non-compliant status.

` + kosliIgnoreDesc

const assertArtifactExample = `
# assert that an artifact meets all compliance requirements for an environment
kosli assert artifact \
	--fingerprint 184c799cd551dd1d8d5c5f9a5d593b2e931f5e36122ee5c793c1d08a19839cc0 \
	--environment prod \
	--api-token yourAPIToken \
	--org yourOrgName 

# assert that an artifact meets a set of policies
kosli assert artifact \
	--fingerprint 184c799cd551dd1d8d5c5f9a5d593b2e931f5e36122ee5c793c1d08a19839cc0 \
	--policy has-approval,has-been-integration-tested \
	--api-token yourAPIToken \
	--org yourOrgName 

# fail if an artifact has a non-compliant status in a single flow (using the artifact fingerprint)
export KOSLI_FLOW=yourFlowName
kosli assert artifact \
	--fingerprint 184c799cd551dd1d8d5c5f9a5d593b2e931f5e36122ee5c793c1d08a19839cc0 \
	--flow yourFlowName \
	--api-token yourAPIToken \
	--org yourOrgName 

# fail if an artifact has a non-compliant status in any flow (using the artifact name and type)
unset KOSLI_FLOW
kosli assert artifact library/nginx:1.21 \
	--artifact-type docker \
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
	cmd.Flags().StringSliceVar(&o.policyNames, "policy", []string{}, policyName)

	addFingerprintFlags(cmd, o.fingerprintOptions)
	addDryRunFlag(cmd)
	cmd.MarkFlagsMutuallyExclusive("environment", "policy")

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

	baseURL, err := url.JoinPath(global.Host, "api/v2/asserts", global.Org, "fingerprint", o.fingerprint)
	if err != nil {
		return err
	}
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

	err = output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printAssertAsTable,
			"json":  output.PrintJson,
		})
	if err != nil {
		return err
	}

	var result assertArtifactResult
	if err := json.Unmarshal([]byte(response.Body), &result); err != nil {
		return err
	}
	if !result.Compliant {
		return fmt.Errorf("Artifact is not compliant")
	}
	return nil
}

type assertArtifactResult struct {
	Scope             string             `json:"scope"`
	Compliant         bool               `json:"compliant"`
	Environment       string             `json:"environment"`
	HTMLURL           string             `json:"html_url"`
	PolicyEvaluations []policyEvaluation `json:"policy_evaluations"`
	Flows             []assertedFlow     `json:"flows"`
}

type policyEvaluation struct {
	PolicyName      string           `json:"policy_name"`
	Status          string           `json:"status"`
	RuleEvaluations []ruleEvaluation `json:"rule_evaluations"`
}

type ruleEvaluation struct {
	Ignored     bool             `json:"ignored"`
	Satisfied   bool             `json:"satisfied"`
	Rule        evaluatedRule    `json:"rule"`
	Resolutions []ruleResolution `json:"resolutions"`
}

type evaluatedRule struct {
	Definition struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"definition"`
}

type ruleResolution struct {
	Type    string `json:"type"`
	Context struct {
		ForControl string `json:"for_control"`
	} `json:"context"`
}

type assertedFlow struct {
	Flow             string `json:"flow"`
	Trail            string `json:"trail"`
	ComplianceStatus struct {
		AttestationsStatuses []attestationStatus `json:"attestations_statuses"`
	} `json:"compliance_status"`
}

type attestationStatus struct {
	AttestationName string `json:"attestation_name"`
	AttestationType string `json:"attestation_type"`
	Status          string `json:"status"`
	IsCompliant     bool   `json:"is_compliant"`
	Unexpected      bool   `json:"unexpected"`
}

func printAssertAsTable(raw string, out io.Writer, page int) error {
	var result assertArtifactResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return err
	}

	if result.Compliant {
		logger.Info("COMPLIANT")
	} else {
		logger.Info("Error: NON-COMPLIANT")
	}

	if result.Scope == "environment" || result.Scope == "policy" {
		if result.Scope == "environment" {
			logger.Info("Environment: %v", result.Environment)
		}
		logger.Info("%-32v %-30v", "Policy-name", "status")
		for _, pe := range result.PolicyEvaluations {
			logger.Info("  %-32v %-30v", pe.PolicyName, pe.Status)
			if pe.Status == "COMPLIANT" {
				continue
			}
			for _, failure := range policyFailures(pe) {
				logger.Info("    %v", failure)
			}
		}
		logger.Info("")
	}

	for _, flow := range result.Flows {
		logger.Info("Flow: %v\n  Trail: %v", flow.Flow, flow.Trail)
		logger.Info("  %-32v %-30v %-15v %-10v", "Attestation-name", "type", "status", "compliant")
		for _, att := range flow.ComplianceStatus.AttestationsStatuses {
			unexpectedStr := ""
			if att.Unexpected {
				unexpectedStr = "unexpected"
			}
			logger.Info("    %-32v %-30v %-15v %-10v %-10v", att.AttestationName, att.AttestationType, att.Status, att.IsCompliant, unexpectedStr)
		}
		logger.Info("  See more details at %s", result.HTMLURL)
	}

	return nil
}

func policyFailures(pe policyEvaluation) []string {
	var failures []string
	for _, re := range pe.RuleEvaluations {
		if re.Ignored || re.Satisfied {
			continue
		}
		name := re.Rule.Definition.Name
		attType := re.Rule.Definition.Type
		for _, res := range re.Resolutions {
			forControl := res.Context.ForControl
			switch res.Type {
			case "legacy_flow":
				failures = append(failures, "artifact comes from a legacy flow and does not have the new attestations")
			case "missing_attestation":
				if forControl != "" {
					failures = append(failures, fmt.Sprintf("artifact is missing required %v for control '%v'", attType, forControl))
				} else {
					failures = append(failures, fmt.Sprintf("artifact is missing required '%v' (type: %v) attestation in trail", name, attType))
				}
			case "non_compliant_attestation":
				if forControl != "" {
					failures = append(failures, fmt.Sprintf("decision for control '%v' is non-compliant in trail", forControl))
				} else {
					failures = append(failures, fmt.Sprintf("attestation '%v' is non-compliant in trail", name))
				}
			case "non_compliant_in_trail":
				failures = append(failures, "artifact is not compliant in trail")
			}
		}
	}
	return failures
}
