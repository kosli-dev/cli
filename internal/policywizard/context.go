package policywizard

import "github.com/charmbracelet/lipgloss"

// Context holds data fetched from the API to populate wizard options.
type Context struct {
	FlowNames         []string
	CustomAttestTypes []string
}

type styles struct {
	base        lipgloss.Style
	title       lipgloss.Style
	preview     lipgloss.Style
	previewText lipgloss.Style
	footer      lipgloss.Style
	accent      lipgloss.Style
}

func newStyles() styles {
	accent := lipgloss.Color("#7571F9")
	green := lipgloss.Color("#02BF87")
	return styles{
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
