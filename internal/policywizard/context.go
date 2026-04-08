package policywizard

import "github.com/charmbracelet/lipgloss"

// FetchResult holds the data returned by the API fetch.
type FetchResult struct {
	FlowNames         []string
	CustomAttestTypes []string
}

// Context holds data fetched from the API to populate wizard options.
// WriteRequest contains the parameters for writing and uploading a policy.
type WriteRequest struct {
	YAMLBytes   []byte
	Filename    string
	PolicyName  string
	Description string
	Org         string
	Upload      bool
}

// WriteResult is returned after writing/uploading.
type WriteResult struct {
	Filename   string
	PolicyName string
	PolicyURL  string // e.g. https://app.kosli.com/my-org/policies/my-policy
	Uploaded   bool
	Err        error
}

type Context struct {
	FlowNames         []string
	CustomAttestTypes []string
	HasAPICredentials bool
	Org               string // current org (e.g. from $KOSLI_ORG)
	Host              string // e.g. https://app.kosli.com
	// FetchFunc is called asynchronously to fetch API data. If nil, no fetch is performed.
	FetchFunc func() FetchResult
	// WriteFunc writes the file and optionally uploads the policy. Called as a tea.Cmd.
	WriteFunc func(WriteRequest) WriteResult
}

// Kosli brand colors for terminal UI.
const (
	colorBlue    = lipgloss.Color("#1C4BC6") // Blue 600 — primary accent
	colorGreen   = lipgloss.Color("#45A26D") // Success green
	colorRed     = lipgloss.Color("#C13D33") // Error red
	colorTextDim = lipgloss.Color("#646A71") // Tertiary text
)

type styles struct {
	base        lipgloss.Style
	title       lipgloss.Style
	preview     lipgloss.Style
	previewText lipgloss.Style
	footer      lipgloss.Style
	accent      lipgloss.Style
	err         lipgloss.Style
}

func newStyles() styles {
	return styles{
		base: lipgloss.NewStyle().Padding(1, 2),
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBlue).
			Padding(0, 1),
		preview: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBlue).
			Padding(1, 2),
		previewText: lipgloss.NewStyle().
			Foreground(colorGreen),
		footer: lipgloss.NewStyle().
			Foreground(colorTextDim).
			Padding(1, 1, 0, 1),
		accent: lipgloss.NewStyle().
			Foreground(colorBlue),
		err: lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true),
	}
}
