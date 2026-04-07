package policywizard

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/kosli-dev/cli/internal/policy"
)

// ---------------------------------------------------------------------------
// Step and target enums
// ---------------------------------------------------------------------------

type wizardStep int

const (
	stepProvConfirm     wizardStep = iota
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
		allTypes := append([]string{}, builtInAttestationTypes...)
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
				Options(huh.NewOptions("==", "!=", "in", "matches")...),
			huh.NewInput().Key("value").Title("Value").
				Description("The value to compare against").
				Validate(notEmpty("value")),
		))

	case stepExprRaw:
		f = inputForm("value", "Raw expression",
			`e.g. flow.name == "prod" and artifact.name == "svc"`,
			`flow.name == "prod"`, "expression")

	case stepSaveFile:
		f = inputForm("filename", "Save policy to file",
			"Enter filename (e.g. policy.yaml)", "policy.yaml", "filename")

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
// State transitions: processFormResults
// ---------------------------------------------------------------------------

func (m *Model) processFormResults() {
	switch m.step {
	case stepProvConfirm:
		m.requireProv = m.form.GetBool("confirm")
		if m.requireProv {
			if m.Policy.Artifacts == nil {
				m.Policy.Artifacts = &policy.ArtifactRules{}
			}
			m.Policy.Artifacts.Provenance = &policy.BooleanRule{Required: true}
		}

	case stepTrailConfirm:
		m.requireTrail = m.form.GetBool("confirm")
		if m.requireTrail {
			if m.Policy.Artifacts == nil {
				m.Policy.Artifacts = &policy.ArtifactRules{}
			}
			m.Policy.Artifacts.TrailCompliance = &policy.BooleanRule{Required: true}
		}

	case stepAttDetails:
		attType := m.form.GetString("type")
		name := m.form.GetString("name")
		if name == "" {
			name = "*"
		}
		if attType == "*" && name == "*" {
			m.validationErr = "when type is *, name must not be * — please specify a name"
			return
		}
		m.validationErr = ""
		m.currentAttRule = policy.AttestationRule{
			Type: attType,
			Name: name,
		}

	case stepExprMode:
		m.exprMode = m.form.GetString("mode")

	case stepExprFlowName:
		m.applyExpression(policy.FlowNameExpr(m.form.GetString("value")))

	case stepExprFlowTag:
		m.exprTagKey = m.form.GetString("value")

	case stepExprFlowTagOp:
		m.applyExpression(policy.FlowTagExpr(m.exprTagKey, m.form.GetString("op"), m.form.GetString("value")))

	case stepExprArtifactName:
		m.applyExpression(policy.ArtifactNameMatchExpr(m.form.GetString("value")))

	case stepExprCustomCtx:
		m.exprContext = m.form.GetString("value")

	case stepExprCustomTagKey:
		m.exprContext = "flow.tags." + m.form.GetString("value")

	case stepExprCustomOp:
		m.applyExpression(policy.ComparisonExpr(m.exprContext, m.form.GetString("op"), m.form.GetString("value")))

	case stepExprRaw:
		m.applyExpression(policy.WrapExpr(m.form.GetString("value")))

	case stepSaveFile:
		m.OutputFile = m.form.GetString("filename")
	}
}

func (m *Model) applyExpression(expr string) {
	switch m.exprTarget {
	case targetProvException:
		m.Policy.Artifacts.Provenance.Exceptions = append(
			m.Policy.Artifacts.Provenance.Exceptions,
			policy.ExceptionRule{If: expr},
		)
	case targetTrailException:
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
		if m.form.GetBool("confirm") {
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
		if m.form.GetBool("confirm") {
			m.exprTarget = targetTrailException
			m.step = stepExprMode
		} else {
			m.step = stepAttConfirm
		}

	case stepAttConfirm:
		if m.form.GetBool("confirm") {
			m.step = stepAttDetails
		} else {
			m.step = stepSaveFile
		}

	case stepAttDetails:
		m.step = stepAttCondConfirm

	case stepAttCondConfirm:
		if m.form.GetBool("confirm") {
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
