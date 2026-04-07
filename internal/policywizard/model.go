package policywizard

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/kosli-dev/cli/internal/policy"
)

const formWidth = 55

// Model is the bubbletea model for the policy wizard.
type Model struct {
	step   wizardStep
	form   *huh.Form
	styles styles
	width  int
	height int
	ctx    *Context

	// Public results — read after the program exits.
	Policy    *policy.Policy
	OutputFile string
	Cancelled  bool

	// Internal state for loops and expression building.
	exprTarget     exprTarget
	exprMode       string
	exprContext    string
	exprTagKey     string
	currentAttRule policy.AttestationRule
	requireProv    bool
	requireTrail   bool
	lastConfirm    bool
	validationErr  string
}

// NewModel creates a new policy wizard model.
func NewModel(ctx *Context) Model {
	m := Model{
		step:   stepProvConfirm,
		Policy: policy.NewPolicy(),
		ctx:    ctx,
		styles: newStyles(),
		width:  120,
	}
	m.form = m.buildForm()
	return m
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.Cancelled = true
			return m, tea.Quit
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if m.form.State == huh.StateAborted {
		m.Cancelled = true
		return m, tea.Quit
	}

	if m.form.State == huh.StateCompleted {
		m.processFormResults()
		if m.validationErr != "" {
			m.form = m.buildForm()
			return m, m.form.Init()
		}
		m.advanceStep()
		if m.step == stepDone {
			return m, tea.Quit
		}
		m.form = m.buildForm()
		return m, m.form.Init()
	}

	return m, cmd
}

func (m Model) View() string {
	if m.Cancelled || m.step == stepDone {
		return ""
	}

	s := m.styles
	fw := formWidth
	available := m.width - s.base.GetHorizontalFrameSize()
	pw := available - fw - 2
	if pw < 30 {
		pw = 0
	}

	header := s.title.Render("Kosli Policy Builder")

	formContent := m.form.View()
	if m.validationErr != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FE5F86")).Bold(true)
		formContent = errStyle.Render("⚠ "+m.validationErr) + "\n\n" + formContent
	}
	formView := lipgloss.NewStyle().Width(fw).Render(formContent)

	var body string
	if pw > 0 {
		yamlBytes, _ := m.Policy.ToYAML()
		yamlStr := strings.TrimRight(string(yamlBytes), "\n")
		if yamlStr == "" {
			yamlStr = "(empty)"
		}
		previewContent := s.previewText.Render(yamlStr)
		previewTitle := s.accent.Bold(true).Render("Live Preview")
		previewPanel := s.preview.Width(pw).
			Render(previewTitle + "\n\n" + previewContent)

		body = lipgloss.JoinHorizontal(lipgloss.Top, formView, "  ", previewPanel)
	} else {
		body = formView
	}

	footer := s.footer.Render("ctrl+c to cancel • enter to confirm")

	return s.base.Render(header + "\n\n" + body + "\n" + footer)
}
