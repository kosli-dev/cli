package docgen

import "github.com/spf13/cobra"

// CommandMeta holds metadata about a cobra command for doc generation.
type CommandMeta struct {
	Name       string
	Beta       bool
	Deprecated bool
	DeprecMsg  string
	Summary    string
	Long       string
	UseLine    string
	Runnable   bool
	Example    string
}

// CIExample holds data for a single CI system's live example.
type CIExample struct {
	CI       string
	YamlURL  string
	EventURL string
}

// LiveExampleData holds all live example data for a command.
type LiveExampleData struct {
	CIExamples []CIExample
	CLICommand string
	CLIURL     string
	CLIExists  bool
}

// Formatter defines the interface for generating doc output in different formats.
type Formatter interface {
	Title(name string) string
	FrontMatter(meta CommandMeta) string
	BetaWarning(name string) string
	DeprecatedWarning(name, message string) string
	Synopsis(meta CommandMeta) string
	FlagsSection(flags, inherited string) string
	LiveCIExamples(examples []CIExample, commandName string) string
	LiveCLIExample(commandName, fullCommand, url string) string
	ExampleUseCases(commandName, example string) string
	LinkHandler(name string) string
}

// CommandMetaFunc is a callback that returns metadata for a cobra command.
// It bridges the cmd/kosli package (which knows about isBeta/isDeprecated)
// with the docgen package.
type CommandMetaFunc func(cmd *cobra.Command) CommandMeta

// LiveDocProvider abstracts the HTTP calls to check for live documentation.
type LiveDocProvider interface {
	YamlDocExists(ci, command string) bool
	EventDocExists(ci, command string) bool
	YamlURL(ci, command string) string
	EventURL(ci, command string) string
	CLIDocExists(command string) (fullCommand, url string, exists bool)
}

// NullLiveDocProvider is a no-op implementation for testing.
type NullLiveDocProvider struct{}

func (NullLiveDocProvider) YamlDocExists(ci, command string) bool              { return false }
func (NullLiveDocProvider) EventDocExists(ci, command string) bool             { return false }
func (NullLiveDocProvider) YamlURL(ci, command string) string                  { return "" }
func (NullLiveDocProvider) EventURL(ci, command string) string                 { return "" }
func (NullLiveDocProvider) CLIDocExists(command string) (string, string, bool) { return "", "", false }
