package policywizard

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/kosli-dev/cli/internal/policy"
)

const formWidth = 55

// fetchDoneMsg is sent when the async API fetch completes.
type fetchDoneMsg struct {
	result FetchResult
}

// Model is the bubbletea model for the policy wizard.
type Model struct {
	step    wizardStep
	form    *huh.Form
	spinner spinner.Model
	styles  styles
	width   int
	height  int
	ctx     *Context

	// Public results — read after the program exits.
	Policy     *policy.Policy
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
	s := spinner.New(spinner.WithSpinner(spinner.Dot))
	s.Style = lipgloss.NewStyle().Foreground(colorBlue)

	startStep := stepProvConfirm
	if ctx.FetchFunc != nil {
		startStep = stepLoading
	}

	m := Model{
		step:    startStep,
		Policy:  policy.NewPolicy(),
		ctx:     ctx,
		styles:  newStyles(),
		spinner: s,
		width:   120,
	}
	if startStep != stepLoading {
		m.form = m.buildForm()
	}
	return m
}

func (m Model) Init() tea.Cmd {
	if m.step == stepLoading {
		return tea.Batch(m.spinner.Tick, m.startFetch())
	}
	return m.form.Init()
}

func (m Model) startFetch() tea.Cmd {
	fetchFn := m.ctx.FetchFunc
	return func() tea.Msg {
		result := fetchFn()
		return fetchDoneMsg{result: result}
	}
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
	case fetchDoneMsg:
		m.ctx.FlowNames = msg.result.FlowNames
		m.ctx.CustomAttestTypes = msg.result.CustomAttestTypes
		m.step = stepProvConfirm
		m.form = m.buildForm()
		return m, m.form.Init()
	}

	// During loading, only update the spinner
	if m.step == stepLoading {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
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
	header := s.title.Render("Kosli Policy Builder")

	if m.step == stepLoading {
		loading := m.spinner.View() + " Fetching flows and attestation types from Kosli..."
		return s.base.Render(header + "\n\n" + loading)
	}

	fw := formWidth
	available := m.width - s.base.GetHorizontalFrameSize()
	pw := available - fw - 2
	if pw < 30 {
		pw = 0
	}

	formContent := m.form.View()
	if m.validationErr != "" {
		formContent = s.err.Render("⚠ "+m.validationErr) + "\n\n" + formContent
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
