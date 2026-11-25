package docgen

import "github.com/spf13/cobra"

// DocHiddenAnnotation marks a hidden command that should still get a generated
// docs page (with hidden: true front matter), as opposed to internal commands
// that are never documented. cmd/kosli sets this annotation; the tree walker
// checks for its presence.
const DocHiddenAnnotation = "docHidden"

// JSONOutputAPIAnnotation marks a command whose --output json prints an API
// response body unchanged. Its value is the operation's path under
// /api-reference on the docs site, e.g. "trails/get-trail". A command whose
// flags or arguments choose between operations lists them, comma-separated.
const JSONOutputAPIAnnotation = "jsonOutputAPI"

// CommandMeta holds metadata about a cobra command for doc generation.
type CommandMeta struct {
	Name       string
	Beta       bool
	Deprecated bool
	DeprecMsg  string
	Hidden     bool
	Summary    string
	Long       string
	UseLine    string
	Runnable   bool
	Example    string
	Tutorial   string
	// JSONOutputAPI is the value of JSONOutputAPIAnnotation, empty when unset.
	JSONOutputAPI string
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
	JSONOutput(operations []string) string
	ExampleUseCases(commandName, example string) string
	LinkHandler(name string) string
}

// CommandMetaFunc is a callback that returns metadata for a cobra command.
// It bridges the cmd/kosli package (which knows about isBeta/isDeprecated)
// with the docgen package.
type CommandMetaFunc func(cmd *cobra.Command) CommandMeta
