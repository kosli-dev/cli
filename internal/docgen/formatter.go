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
	Tutorial   string
}

// Formatter defines the interface for generating doc output in different formats.
type Formatter interface {
	Title(name string) string
	FrontMatter(meta CommandMeta) string
	BetaWarning(name string) string
	DeprecatedWarning(name, message string) string
	TutorialTip(url string) string
	Synopsis(meta CommandMeta) string
	FlagsSection(flags, inherited string) string
	ExampleUseCases(commandName, example string) string
	LinkHandler(name string) string
}

// CommandMetaFunc is a callback that returns metadata for a cobra command.
// It bridges the cmd/kosli package (which knows about isBeta/isDeprecated)
// with the docgen package.
type CommandMetaFunc func(cmd *cobra.Command) CommandMeta
