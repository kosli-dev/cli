package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/kosli-dev/cli/internal/policy"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const createPolicyFileShortDesc = `Interactively create a Kosli environment policy YAML file.`

const createPolicyFileLongDesc = createPolicyFileShortDesc + `
Launches an interactive wizard that guides you through building a policy file
conforming to the Kosli environment policy schema. The generated YAML is
written to stdout by default, or to a file with ^--output-file^.

This command does not upload the policy to Kosli. Use ^kosli create policy^
to upload the generated file.

If ^--api-token^ and ^--org^ are set, the wizard will fetch flow names and
custom attestation types from the Kosli API to offer as suggestions.
`

const createPolicyFileExample = `
# create a policy file interactively:
kosli create policy-file
`

func newCreatePolicyFileCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "policy-file",
		Short:   createPolicyFileShortDesc,
		Long:    createPolicyFileLongDesc,
		Example: createPolicyFileExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreatePolicyFile()
		},
	}

	return cmd
}

func runCreatePolicyFile() error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("this command requires an interactive terminal; write policy YAML manually or use 'kosli create policy' directly")
	}

	wctx := &wizardContext{}
	if global.ApiToken != "" && global.Org != "" {
		wctx.fetchFromAPI()
	}

	m := newPolicyWizardModel(wctx)
	finalModel, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return fmt.Errorf("wizard error: %w", err)
	}

	wm := finalModel.(policyWizardModel)
	if wm.cancelled {
		logger.Info("policy file creation cancelled")
		return nil
	}

	yamlBytes, err := wm.policy.ToYAML()
	if err != nil {
		return fmt.Errorf("failed to generate policy YAML: %w", err)
	}

	filename := wm.outputFile
	if filename == "" {
		filename = "policy.yaml"
	}
	err = os.WriteFile(filename, yamlBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write policy file: %w", err)
	}
	logger.Info("policy file written to %s", filename)
	return nil
}

// ---------------------------------------------------------------------------
// Wizard step enum
// ---------------------------------------------------------------------------

type wizardStep int

const (
	stepProvConfirm     wizardStep = iota // require provenance?
	stepProvExcConfirm                    // add provenance exception?
	stepTrailConfirm                      // require trail compliance?
	stepTrailExcConfirm                   // add trail compliance exception?
	stepAttConfirm                         // add attestation?
	stepAttDetails                         // attestation type + name
	stepAttCondConfirm                     // add condition for attestation?
	stepExprMode                           // choose expression mode
	stepExprFlowName                       // flow name input/select
	stepExprFlowTag                        // tag key input
	stepExprFlowTagOp                      // tag operator + value
	stepExprArtifactName                   // artifact regex input
	stepExprCustomCtx                      // custom context select
	stepExprCustomTagKey                   // tag key for custom context
	stepExprCustomOp                       // custom operator + value
	stepExprRaw                            // raw expression input
	stepSaveFile                           // ask for filename
	stepDone
)

// exprTarget tracks what we're building an expression for.
type exprTarget int

const (
	targetProvException exprTarget = iota
	targetTrailException
	targetAttCondition
)

// ---------------------------------------------------------------------------
// Wizard context (API data)
// ---------------------------------------------------------------------------

type wizardContext struct {
	flowNames         []string
	customAttestTypes []string
}

// ---------------------------------------------------------------------------
// Styles
// ---------------------------------------------------------------------------

type wizardStyles struct {
	base        lipgloss.Style
	title       lipgloss.Style
	preview     lipgloss.Style
	previewText lipgloss.Style
	footer      lipgloss.Style
	accent      lipgloss.Style
}

func newWizardStyles() wizardStyles {
	accent := lipgloss.Color("#7571F9")
	green := lipgloss.Color("#02BF87")
	return wizardStyles{
		base: lipgloss.NewStyle().Padding(1, 2),
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(accent).
			Padding(0, 1),
		preview: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Padding(1, 2),
		previewText: lipgloss.NewStyle().
			Foreground(green),
		footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(1, 1, 0, 1),
		accent: lipgloss.NewStyle().
			Foreground(accent),
	}
}

// ---------------------------------------------------------------------------
// Bubbletea model
// ---------------------------------------------------------------------------

const formWidth = 55

type policyWizardModel struct {
	step   wizardStep
	form   *huh.Form
	policy *policy.Policy
	wctx   *wizardContext
	styles wizardStyles
	width  int
	height int

	// State for loops and expression building
	exprTarget     exprTarget
	exprMode       string
	exprContext    string // for custom expressions
	exprTagKey     string // for flow tag / custom tag
	currentAttRule policy.AttestationRule
	cancelled      bool
	outputFile     string
	requireProv    bool
	requireTrail   bool
}

func newPolicyWizardModel(wctx *wizardContext) policyWizardModel {
	m := policyWizardModel{
		step:   stepProvConfirm,
		policy: policy.NewPolicy(),
		wctx:   wctx,
		styles: newWizardStyles(),
		width:  120, // sensible default until WindowSizeMsg arrives
	}
	m.form = m.buildForm()
	return m
}

func (m policyWizardModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m policyWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.cancelled = true
			return m, tea.Quit
		}
	}

	// Always forward to form (including WindowSizeMsg)
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if m.form.State == huh.StateAborted {
		m.cancelled = true
		return m, tea.Quit
	}

	if m.form.State == huh.StateCompleted {
		m.processFormResults()
		m.advanceStep()
		if m.step == stepDone {
			return m, tea.Quit
		}
		m.form = m.buildForm()
		return m, m.form.Init()
	}

	return m, cmd
}

func (m policyWizardModel) View() string {
	if m.cancelled {
		return ""
	}
	if m.step == stepDone {
		return ""
	}

	s := m.styles
	fw := formWidth
	available := m.width - s.base.GetHorizontalFrameSize()
	// Give remaining width to preview, with a gap of 2
	pw := available - fw - 2
	if pw < 30 {
		pw = 0 // hide preview if terminal too narrow
	}

	// Header
	header := s.title.Render("Kosli Policy Builder")

	// Form (left)
	formView := lipgloss.NewStyle().
		Width(fw).
		Render(m.form.View())

	var body string
	if pw > 0 {
		// YAML preview (right)
		yamlBytes, _ := m.policy.ToYAML()
		yamlStr := strings.TrimRight(string(yamlBytes), "\n")
		if yamlStr == "" {
			yamlStr = "(empty)"
		}
		previewContent := s.previewText.Render(yamlStr)
		previewTitle := s.accent.Bold(true).Render("Live Preview")
		previewPanel := s.preview.
			Width(pw).
			Render(previewTitle + "\n\n" + previewContent)

		body = lipgloss.JoinHorizontal(lipgloss.Top, formView, "  ", previewPanel)
	} else {
		body = formView
	}

	// Footer
	footer := s.footer.Render("ctrl+c to cancel • enter to confirm")

	return s.base.Render(header + "\n\n" + body + "\n" + footer)
}

// ---------------------------------------------------------------------------
// Form builders — one per step
// ---------------------------------------------------------------------------

var builtInAttestationTypes = []string{
	"generic", "junit", "snyk", "pull_request", "jira", "sonar",
}

func (m *policyWizardModel) buildForm() *huh.Form {
	var f *huh.Form
	switch m.step {
	case stepProvConfirm:
		f = huh.NewForm(huh.NewGroup(
			huh.NewConfirm().Key("confirm").
				Title("Require artifact provenance?").
				Description("All artifacts must belong to a Kosli flow").
				Affirmative("Yes").Negative("No"),
		))

	case stepTrailConfirm:
		f = huh.NewForm(huh.NewGroup(
			huh.NewConfirm().Key("confirm").
				Title("Require trail compliance?").
				Description("All artifacts must be part of compliant trails").
				Affirmative("Yes").Negative("No"),
		))

	case stepProvExcConfirm:
		f = huh.NewForm(huh.NewGroup(
			huh.NewConfirm().Key("confirm").
				Title(m.excConfirmTitle("provenance")).
				Description("Exceptions waive this requirement for matching artifacts").
				Affirmative("Yes").Negative("No"),
		))

	case stepTrailExcConfirm:
		f = huh.NewForm(huh.NewGroup(
			huh.NewConfirm().Key("confirm").
				Title(m.excConfirmTitle("trail compliance")).
				Description("Exceptions waive this requirement for matching artifacts").
				Affirmative("Yes").Negative("No"),
		))

	case stepAttConfirm:
		title := "Add a required attestation?"
		if m.policy.Artifacts != nil && len(m.policy.Artifacts.Attestations) > 0 {
			title = "Add another required attestation?"
		}
		f = huh.NewForm(huh.NewGroup(
			huh.NewConfirm().Key("confirm").
				Title(title).
				Affirmative("Yes").Negative("No"),
		))

	case stepAttDetails:
		allTypes := append([]string{}, builtInAttestationTypes...)
		allTypes = append(allTypes, m.wctx.customAttestTypes...)
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
		f = huh.NewForm(huh.NewGroup(
			huh.NewConfirm().Key("confirm").
				Title("Add a condition for this attestation?").
				Description("Only require when condition is met").
				Affirmative("Yes").Negative("No"),
		))

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
		if len(m.wctx.flowNames) > 0 {
			opts := make([]huh.Option[string], len(m.wctx.flowNames))
			for i, n := range m.wctx.flowNames {
				opts[i] = huh.NewOption(n, n)
			}
			f = huh.NewForm(huh.NewGroup(
				huh.NewSelect[string]().Key("value").
					Title("Select a flow").
					Options(opts...),
			))
		} else {
			f = huh.NewForm(huh.NewGroup(
				huh.NewInput().Key("value").
					Title("Flow name").
					Description("The flow name to match").
					Validate(notEmpty("flow name")),
			))
		}

	case stepExprFlowTag:
		f = huh.NewForm(huh.NewGroup(
			huh.NewInput().Key("value").
				Title("Tag key").
				Description("e.g. team, risk-level, key.with.dots").
				Validate(notEmpty("tag key")),
		))

	case stepExprFlowTagOp:
		f = huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().Key("op").
				Title("Operator").
				Options(huh.NewOptions("==", "!=", ">", "<", ">=", "<=")...),
			huh.NewInput().Key("value").
				Title("Value").
				Description("The value to compare against").
				Validate(notEmpty("value")),
		))

	case stepExprArtifactName:
		f = huh.NewForm(huh.NewGroup(
			huh.NewInput().Key("value").
				Title("Artifact name regex").
				Description("e.g. ^datadog:.*").
				Placeholder("^datadog:.*").
				Validate(notEmpty("regex")),
		))

	case stepExprCustomCtx:
		f = huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().Key("value").
				Title("Context field").
				Options(
					huh.NewOption("flow.name", "flow.name"),
					huh.NewOption("flow.tags.<key>", "flow.tags.<key>"),
					huh.NewOption("artifact.name", "artifact.name"),
					huh.NewOption("artifact.fingerprint", "artifact.fingerprint"),
				),
		))

	case stepExprCustomTagKey:
		f = huh.NewForm(huh.NewGroup(
			huh.NewInput().Key("value").
				Title("Tag key").
				Description("The flow tag key (e.g. team, risk-level)").
				Validate(notEmpty("tag key")),
		))

	case stepExprCustomOp:
		f = huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().Key("op").
				Title("Operator").
				Options(huh.NewOptions("==", "!=", "in", "matches")...),
			huh.NewInput().Key("value").
				Title("Value").
				Description("The value to compare against").
				Validate(notEmpty("value")),
		))

	case stepExprRaw:
		f = huh.NewForm(huh.NewGroup(
			huh.NewInput().Key("value").
				Title("Raw expression").
				Description(`e.g. flow.name == "prod" and artifact.name == "svc"`).
				Placeholder(`flow.name == "prod"`).
				Validate(notEmpty("expression")),
		))

	case stepSaveFile:
		f = huh.NewForm(huh.NewGroup(
			huh.NewInput().Key("filename").
				Title("Save policy to file").
				Description("Enter filename (e.g. policy.yaml)").
				Placeholder("policy.yaml").
				Validate(notEmpty("filename")),
		))

	default:
		f = huh.NewForm(huh.NewGroup())
	}

	return f.WithWidth(formWidth).WithShowHelp(true).WithShowErrors(true)
}

func notEmpty(field string) func(string) error {
	return func(s string) error {
		if s == "" {
			return fmt.Errorf("%s is required", field)
		}
		return nil
	}
}

func (m *policyWizardModel) excConfirmTitle(rule string) string {
	var count int
	if rule == "provenance" && m.policy.Artifacts != nil && m.policy.Artifacts.Provenance != nil {
		count = len(m.policy.Artifacts.Provenance.Exceptions)
	}
	if rule == "trail compliance" && m.policy.Artifacts != nil && m.policy.Artifacts.TrailCompliance != nil {
		count = len(m.policy.Artifacts.TrailCompliance.Exceptions)
	}
	if count > 0 {
		return fmt.Sprintf("Add another exception to %s?", rule)
	}
	return fmt.Sprintf("Add an exception to %s?", rule)
}

// ---------------------------------------------------------------------------
// State transitions
// ---------------------------------------------------------------------------

func (m *policyWizardModel) processFormResults() {
	switch m.step {
	case stepProvConfirm:
		m.requireProv = m.form.GetBool("confirm")
		if m.requireProv {
			if m.policy.Artifacts == nil {
				m.policy.Artifacts = &policy.ArtifactRules{}
			}
			m.policy.Artifacts.Provenance = &policy.BooleanRule{Required: true}
		}

	case stepTrailConfirm:
		m.requireTrail = m.form.GetBool("confirm")
		if m.requireTrail {
			if m.policy.Artifacts == nil {
				m.policy.Artifacts = &policy.ArtifactRules{}
			}
			m.policy.Artifacts.TrailCompliance = &policy.BooleanRule{Required: true}
		}

	case stepAttDetails:
		name := m.form.GetString("name")
		if name == "" {
			name = "*"
		}
		m.currentAttRule = policy.AttestationRule{
			Type: m.form.GetString("type"),
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
		m.outputFile = m.form.GetString("filename")
	}
}

func (m *policyWizardModel) applyExpression(expr string) {
	switch m.exprTarget {
	case targetProvException:
		m.policy.Artifacts.Provenance.Exceptions = append(
			m.policy.Artifacts.Provenance.Exceptions,
			policy.ExceptionRule{If: expr},
		)
	case targetTrailException:
		m.policy.Artifacts.TrailCompliance.Exceptions = append(
			m.policy.Artifacts.TrailCompliance.Exceptions,
			policy.ExceptionRule{If: expr},
		)
	case targetAttCondition:
		m.currentAttRule.If = expr
		m.commitAttestation()
	}
}

func (m *policyWizardModel) commitAttestation() {
	if m.policy.Artifacts == nil {
		m.policy.Artifacts = &policy.ArtifactRules{}
	}
	m.policy.Artifacts.Attestations = append(m.policy.Artifacts.Attestations, m.currentAttRule)
	m.currentAttRule = policy.AttestationRule{}
}

func (m *policyWizardModel) advanceStep() {
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

func (m *policyWizardModel) advanceAfterExpr() {
	switch m.exprTarget {
	case targetProvException:
		m.step = stepProvExcConfirm
	case targetTrailException:
		m.step = stepTrailExcConfirm
	case targetAttCondition:
		// attestation was committed in applyExpression
		m.step = stepAttConfirm
	}
}

// ---------------------------------------------------------------------------
// API fetching
// ---------------------------------------------------------------------------

func (wctx *wizardContext) fetchFromAPI() {
	wctx.flowNames = fetchFlowNames()
	wctx.customAttestTypes = fetchCustomAttestationTypes()
}

func fetchFlowNames() []string {
	u, err := url.JoinPath(global.Host, "api/v2/flows", global.Org)
	if err != nil {
		logger.Debug("failed to build flows URL: %v", err)
		return nil
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    u,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		logger.Debug("failed to fetch flows: %v", err)
		return nil
	}

	var flows []map[string]any
	if err := json.Unmarshal([]byte(response.Body), &flows); err != nil {
		logger.Debug("failed to parse flows response: %v", err)
		return nil
	}

	names := make([]string, 0, len(flows))
	for _, flow := range flows {
		if name, ok := flow["name"].(string); ok {
			names = append(names, name)
		}
	}
	return names
}

func fetchCustomAttestationTypes() []string {
	u, err := url.JoinPath(global.Host, "api/v2/custom-attestation-types", global.Org)
	if err != nil {
		logger.Debug("failed to build attestation types URL: %v", err)
		return nil
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    u,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		logger.Debug("failed to fetch attestation types: %v", err)
		return nil
	}

	var types []map[string]any
	if err := json.Unmarshal([]byte(response.Body), &types); err != nil {
		logger.Debug("failed to parse attestation types response: %v", err)
		return nil
	}

	names := make([]string, 0, len(types))
	for _, t := range types {
		if name, ok := t["name"].(string); ok {
			names = append(names, "custom:"+name)
		}
	}
	return names
}
