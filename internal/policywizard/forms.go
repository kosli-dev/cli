package policywizard

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/kosli-dev/cli/internal/policy"
)

// ---------------------------------------------------------------------------
// Step and target enums
// ---------------------------------------------------------------------------

type wizardStep int

const (
	stepLoading wizardStep = iota
	stepProvConfirm
	stepProvExcConfirm
	stepTrailConfirm
	stepTrailExcConfirm
	stepAttConfirm
	stepAttDetails
	stepAttCondConfirm
	stepExprMode
	stepExprFlowName
	stepExprFlowTag
	stepExprFlowTagOp
	stepExprArtifactName
	stepExprCustomCtx
	stepExprCustomTagKey
	stepExprCustomOp
	stepExprRaw
	stepSaveFile
	stepDone
)

type exprTarget int

const (
	targetProvException exprTarget = iota
	targetTrailException
	targetAttCondition
)

// ---------------------------------------------------------------------------
// Form builders
// ---------------------------------------------------------------------------

var builtInAttestationTypes = []string{
	"generic", "junit", "snyk", "pull_request", "jira", "sonar", "*",
}

func (m *Model) buildForm() *huh.Form {
	var f *huh.Form
	switch m.step {
	case stepProvConfirm:
		f = confirmForm("Require artifact provenance?",
			"All artifacts must belong to a Kosli flow")

	case stepTrailConfirm:
		f = confirmForm("Require trail compliance?",
			"All artifacts must be part of compliant trails")

	case stepProvExcConfirm:
		f = confirmForm(m.excConfirmTitle("provenance"),
			"Exceptions waive this requirement for matching artifacts")

	case stepTrailExcConfirm:
		f = confirmForm(m.excConfirmTitle("trail compliance"),
			"Exceptions waive this requirement for matching artifacts")

	case stepAttConfirm:
		title := "Add a required attestation?"
		if m.Policy.Artifacts != nil && len(m.Policy.Artifacts.Attestations) > 0 {
			title = "Add another required attestation?"
		}
		f = confirmForm(title, "")

	case stepAttDetails:
		allTypes := slices.Clone(builtInAttestationTypes)
		allTypes = append(allTypes, m.ctx.CustomAttestTypes...)
		opts := make([]huh.Option[string], len(allTypes))
		for i, t := range allTypes {
			opts[i] = huh.NewOption(t, t)
		}
		f = huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().Key("type").
				Title("Attestation type").
				Options(opts...),
			huh.NewInput().Key("name").
				Title("Attestation name").
				Description("Use * to match any name for this type").
				Placeholder("*"),
		))

	case stepAttCondConfirm:
		f = confirmForm("Add a condition for this attestation?",
			"Only require when condition is met")

	case stepExprMode:
		f = huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().Key("mode").
				Title("How do you want to define this condition?").
				Options(
					huh.NewOption("Match by flow name", "flow_name"),
					huh.NewOption("Match by flow tag", "flow_tag"),
					huh.NewOption("Match by artifact name pattern", "artifact_name"),
					huh.NewOption("Custom comparison", "custom"),
					huh.NewOption("Write raw expression", "raw"),
				),
		))

	case stepExprFlowName:
		if len(m.ctx.FlowNames) > 0 {
			opts := make([]huh.Option[string], len(m.ctx.FlowNames))
			for i, n := range m.ctx.FlowNames {
				opts[i] = huh.NewOption(n, n)
			}
			f = huh.NewForm(huh.NewGroup(
				huh.NewSelect[string]().Key("value").
					Title("Select a flow").Options(opts...),
			))
		} else {
			f = inputForm("value", "Flow name", "The flow name to match", "", "flow name")
		}

	case stepExprFlowTag:
		f = inputForm("value", "Tag key", "e.g. team, risk-level, key.with.dots", "", "tag key")

	case stepExprFlowTagOp:
		f = huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().Key("op").Title("Operator").
				Options(huh.NewOptions("==", "!=", ">", "<", ">=", "<=")...),
			huh.NewInput().Key("value").Title("Value").
				Description("The value to compare against").
				Validate(notEmpty("value")),
		))

	case stepExprArtifactName:
		f = inputForm("value", "Artifact name regex", "e.g. ^datadog:.*", "^datadog:.*", "regex")

	case stepExprCustomCtx:
		f = huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().Key("value").Title("Context field").
				Options(
					huh.NewOption("flow.name", "flow.name"),
					huh.NewOption("flow.tags.<key>", "flow.tags.<key>"),
					huh.NewOption("artifact.name", "artifact.name"),
					huh.NewOption("artifact.fingerprint", "artifact.fingerprint"),
				),
		))

	case stepExprCustomTagKey:
		f = inputForm("value", "Tag key", "The flow tag key (e.g. team, risk-level)", "", "tag key")

	case stepExprCustomOp:
		f = huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().Key("op").Title("Operator").
				Options(huh.NewOptions("==", "!=", ">", "<", ">=", "<=", "matches")...),
			huh.NewInput().Key("value").Title("Value").
				Description("The value to compare against").
				Validate(notEmpty("value")),
		))

	case stepExprRaw:
		f = inputForm("value", "Raw expression",
			`e.g. flow.name == "prod" and artifact.name == "svc"`,
			`flow.name == "prod"`, "expression")

	case stepSaveFile:
		f = huh.NewForm(huh.NewGroup(
			huh.NewInput().Key("filename").
				Title("Save policy to file").
				Description("Press enter to accept default").
				Placeholder("policy.yaml").
				Validate(validateYAMLExtension),
		))

	default:
		f = huh.NewForm(huh.NewGroup())
	}

	return f.WithWidth(formWidth).WithShowHelp(true).WithShowErrors(true)
}

// ---------------------------------------------------------------------------
// Form helpers
// ---------------------------------------------------------------------------

func confirmForm(title, description string) *huh.Form {
	c := huh.NewConfirm().Key("confirm").
		Title(title).
		Affirmative("Yes").Negative("No")
	if description != "" {
		c = c.Description(description)
	}
	return huh.NewForm(huh.NewGroup(c))
}

func inputForm(key, title, description, placeholder, requiredName string) *huh.Form {
	inp := huh.NewInput().Key(key).Title(title).
		Description(description).
		Placeholder(placeholder).
		Validate(notEmpty(requiredName))
	return huh.NewForm(huh.NewGroup(inp))
}

func notEmpty(field string) func(string) error {
	return func(s string) error {
		if s == "" {
			return fmt.Errorf("%s is required", field)
		}
		return nil
	}
}

func validateYAMLExtension(s string) error {
	if s == "" {
		return nil // placeholder "policy.yaml" will be used
	}
	ext := strings.ToLower(filepath.Ext(s))
	if ext != ".yaml" && ext != ".yml" {
		return fmt.Errorf("file must have a .yaml or .yml extension")
	}
	return nil
}

func (m *Model) excConfirmTitle(rule string) string {
	var count int
	if rule == "provenance" && m.Policy.Artifacts != nil && m.Policy.Artifacts.Provenance != nil {
		count = len(m.Policy.Artifacts.Provenance.Exceptions)
	}
	if rule == "trail compliance" && m.Policy.Artifacts != nil && m.Policy.Artifacts.TrailCompliance != nil {
		count = len(m.Policy.Artifacts.TrailCompliance.Exceptions)
	}
	if count > 0 {
		return fmt.Sprintf("Add another exception to %s?", rule)
	}
	return fmt.Sprintf("Add an exception to %s?", rule)
}

// ---------------------------------------------------------------------------
// Form values — extracted from huh form for testability
// ---------------------------------------------------------------------------

type formValues struct {
	confirm  bool
	str      string // generic string: value, filename, mode
	attType  string
	attName  string
	operator string
}

func extractFormValues(f *huh.Form) formValues {
	return formValues{
		confirm:  f.GetBool("confirm"),
		str:      firstNonEmpty(f.GetString("value"), f.GetString("filename"), f.GetString("mode")),
		attType:  f.GetString("type"),
		attName:  f.GetString("name"),
		operator: f.GetString("op"),
	}
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// State transitions: processFormResults
// ---------------------------------------------------------------------------

func (m *Model) processFormResults() {
	m.applyFormValues(extractFormValues(m.form))
}

func (m *Model) applyFormValues(fv formValues) {
	m.lastConfirm = fv.confirm
	switch m.step {
	case stepProvConfirm:
		m.requireProv = fv.confirm
		if m.requireProv {
			if m.Policy.Artifacts == nil {
				m.Policy.Artifacts = &policy.ArtifactRules{}
			}
			m.Policy.Artifacts.Provenance = &policy.BooleanRule{Required: true}
		}

	case stepTrailConfirm:
		m.requireTrail = fv.confirm
		if m.requireTrail {
			if m.Policy.Artifacts == nil {
				m.Policy.Artifacts = &policy.ArtifactRules{}
			}
			m.Policy.Artifacts.TrailCompliance = &policy.BooleanRule{Required: true}
		}

	case stepAttDetails:
		name := fv.attName
		if name == "" {
			name = "*"
		}
		if fv.attType == "*" && name == "*" {
			m.validationErr = "when type is *, name must not be * — please specify a name"
			return
		}
		m.validationErr = ""
		m.currentAttRule = policy.AttestationRule{
			Type: fv.attType,
			Name: name,
		}

	case stepExprMode:
		m.exprMode = fv.str

	case stepExprFlowName:
		m.applyExpression(policy.FlowNameExpr(fv.str))

	case stepExprFlowTag:
		m.exprTagKey = fv.str

	case stepExprFlowTagOp:
		m.applyExpression(policy.FlowTagExpr(m.exprTagKey, fv.operator, fv.str))

	case stepExprArtifactName:
		m.applyExpression(policy.ArtifactNameMatchExpr(fv.str))

	case stepExprCustomCtx:
		m.exprContext = fv.str

	case stepExprCustomTagKey:
		m.exprContext = "flow.tags." + fv.str

	case stepExprCustomOp:
		if fv.operator == "matches" {
			m.applyExpression(policy.MatchesExpr(m.exprContext, fv.str))
		} else {
			m.applyExpression(policy.ComparisonExpr(m.exprContext, fv.operator, fv.str))
		}

	case stepExprRaw:
		m.applyExpression(policy.WrapExpr(fv.str))

	case stepSaveFile:
		m.OutputFile = fv.str
		if m.OutputFile == "" {
			m.OutputFile = "policy.yaml"
		}
	}
}

func (m *Model) applyExpression(expr string) {
	switch m.exprTarget {
	case targetProvException:
		if m.Policy.Artifacts == nil || m.Policy.Artifacts.Provenance == nil {
			return
		}
		m.Policy.Artifacts.Provenance.Exceptions = append(
			m.Policy.Artifacts.Provenance.Exceptions,
			policy.ExceptionRule{If: expr},
		)
	case targetTrailException:
		if m.Policy.Artifacts == nil || m.Policy.Artifacts.TrailCompliance == nil {
			return
		}
		m.Policy.Artifacts.TrailCompliance.Exceptions = append(
			m.Policy.Artifacts.TrailCompliance.Exceptions,
			policy.ExceptionRule{If: expr},
		)
	case targetAttCondition:
		m.currentAttRule.If = expr
		m.commitAttestation()
	}
}

func (m *Model) commitAttestation() {
	if m.Policy.Artifacts == nil {
		m.Policy.Artifacts = &policy.ArtifactRules{}
	}
	m.Policy.Artifacts.Attestations = append(m.Policy.Artifacts.Attestations, m.currentAttRule)
	m.currentAttRule = policy.AttestationRule{}
}

// ---------------------------------------------------------------------------
// State transitions: advanceStep
// ---------------------------------------------------------------------------

func (m *Model) advanceStep() {
	switch m.step {
	case stepProvConfirm:
		if m.requireProv {
			m.exprTarget = targetProvException
			m.step = stepProvExcConfirm
		} else {
			m.step = stepTrailConfirm
		}

	case stepProvExcConfirm:
		if m.lastConfirm {
			m.exprTarget = targetProvException
			m.step = stepExprMode
		} else {
			m.step = stepTrailConfirm
		}

	case stepTrailConfirm:
		if m.requireTrail {
			m.exprTarget = targetTrailException
			m.step = stepTrailExcConfirm
		} else {
			m.step = stepAttConfirm
		}

	case stepTrailExcConfirm:
		if m.lastConfirm {
			m.exprTarget = targetTrailException
			m.step = stepExprMode
		} else {
			m.step = stepAttConfirm
		}

	case stepAttConfirm:
		if m.lastConfirm {
			m.step = stepAttDetails
		} else {
			m.step = stepSaveFile
		}

	case stepAttDetails:
		m.step = stepAttCondConfirm

	case stepAttCondConfirm:
		if m.lastConfirm {
			m.exprTarget = targetAttCondition
			m.step = stepExprMode
		} else {
			m.commitAttestation()
			m.step = stepAttConfirm
		}

	case stepExprMode:
		switch m.exprMode {
		case "flow_name":
			m.step = stepExprFlowName
		case "flow_tag":
			m.step = stepExprFlowTag
		case "artifact_name":
			m.step = stepExprArtifactName
		case "custom":
			m.step = stepExprCustomCtx
		case "raw":
			m.step = stepExprRaw
		}

	case stepExprFlowName:
		m.advanceAfterExpr()
	case stepExprFlowTag:
		m.step = stepExprFlowTagOp
	case stepExprFlowTagOp:
		m.advanceAfterExpr()
	case stepExprArtifactName:
		m.advanceAfterExpr()

	case stepExprCustomCtx:
		if m.exprContext == "flow.tags.<key>" {
			m.step = stepExprCustomTagKey
		} else {
			m.step = stepExprCustomOp
		}

	case stepExprCustomTagKey:
		m.step = stepExprCustomOp
	case stepExprCustomOp:
		m.advanceAfterExpr()
	case stepExprRaw:
		m.advanceAfterExpr()

	case stepSaveFile:
		m.step = stepDone
	}
}

func (m *Model) advanceAfterExpr() {
	switch m.exprTarget {
	case targetProvException:
		m.step = stepProvExcConfirm
	case targetTrailException:
		m.step = stepTrailExcConfirm
	case targetAttCondition:
		m.step = stepAttConfirm
	}
}
