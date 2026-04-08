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

// writeDoneMsg is sent when the file write/upload completes.
type writeDoneMsg struct {
	result WriteResult
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
	Policy            *policy.Policy
	OutputFile        string
	Cancelled         bool
	UploadPolicy      bool
	UploadPolicyName  string
	UploadDescription string
	UploadOrg         string

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
	writeResult    *WriteResult
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

func (m Model) completionView() string {
	s := m.styles
	if m.writeResult == nil {
		return ""
	}
	r := m.writeResult
	if r.Err != nil {
		return s.err.Render("✗ Error: " + r.Err.Error()) + "\n\n" +
			s.footer.Render("Press any key to exit")
	}

	var b strings.Builder
	b.WriteString(s.accent.Bold(true).Render("✓ Policy file saved") + "  " + r.Filename + "\n")
	if r.Uploaded {
		b.WriteString(s.accent.Bold(true).Render("✓ Policy created") + "    " + r.PolicyName + "\n")
		if r.PolicyURL != "" {
			b.WriteString("\n" + s.accent.Render("→") + " " + r.PolicyURL + "\n")
		}
	}
	b.WriteString("\n" + s.footer.Render("Press any key to exit"))
	return b.String()
}

func (m Model) startWrite() tea.Cmd {
	yamlBytes, _ := m.Policy.ToYAML()
	req := WriteRequest{
		YAMLBytes:   yamlBytes,
		Filename:    m.OutputFile,
		PolicyName:  m.UploadPolicyName,
		Description: m.UploadDescription,
		Org:         m.UploadOrg,
		Upload:      m.UploadPolicy,
	}
	writeFn := m.ctx.WriteFunc
	return func() tea.Msg {
		result := writeFn(req)
		return writeDoneMsg{result: result}
	}
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
	case writeDoneMsg:
		m.writeResult = &msg.result
		m.step = stepComplete
		return m, nil
	}

	// During loading/writing, only update the spinner
	if m.step == stepLoading || m.step == stepWriting {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// On the completion screen, any key quits
	if m.step == stepComplete {
		if _, ok := msg.(tea.KeyMsg); ok {
			return m, tea.Quit
		}
		return m, nil
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
		if m.step == stepWriting {
			return m, m.startWrite()
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

	if m.step == stepWriting {
		loading := m.spinner.View() + " Saving policy..."
		return s.base.Render(header + "\n\n" + loading)
	}

	if m.step == stepComplete {
		return s.base.Render(header + "\n\n" + m.completionView() + "\n")
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
