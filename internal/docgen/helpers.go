package docgen

import (
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// FlagTableHeader is the markdown header row (plus alignment row) for the table
// whose body CommandsInTable renders. It lives next to the row builder so the
// column count cannot drift between the header and the rows.
const FlagTableHeader = "| Flag | Type | Description |\n| :--- | :--- | :--- |\n"

// escapeTableCell escapes the pipes in a cell's content so markdown cannot read
// them as column separators. GFM unescapes \| in a table cell before inline
// parsing, so this survives even inside a code span — which matters because
// MintlifyFormatter later turns placeholders like <hours|days> into `hours|days`.
// It is deliberately applied here rather than in escapeMintlifyProse: that
// escaper also runs over prose, where a table row's \| would render as a
// literal backslash.
func escapeTableCell(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}

// CommandsInTable renders a pflag.FlagSet as the body of the table headed by
// FlagTableHeader: one row per flag, with the flag name(s), the value type
// (empty for booleans, which pflag leaves unnamed), and the description.
// Defaults and deprecation notices are appended to the description. Every cell
// is pipe-escaped, so each row has exactly as many separators as the header.
func CommandsInTable(f *pflag.FlagSet) string {
	var b strings.Builder

	f.VisitAll(func(flag *pflag.Flag) {
		// pflag.MarkDeprecated sets Hidden as a side effect, which would keep
		// deprecated-but-working flags out of the reference docs along with
		// their migration message. Only flags hidden on purpose (MarkHidden,
		// e.g. internal aliases) are skipped; deprecated ones are documented
		// below with a (DEPRECATED: ...) suffix.
		if flag.Hidden && flag.Deprecated == "" {
			return
		}

		// No alignment padding: markdown strips leading whitespace inside a
		// table cell, so it would only add noise to the generated MDX.
		var flagName string
		if flag.Shorthand != "" && flag.ShorthandDeprecated == "" {
			flagName = fmt.Sprintf("-%s, --%s", flag.Shorthand, flag.Name)
		} else {
			flagName = "--" + flag.Name
		}

		// varname is the value type ("string", "strings", ...) and is empty for
		// booleans. NoOptDefVal describes the optional-value syntax, so it is
		// appended here rather than to the description.
		varname, usage := pflag.UnquoteUsage(flag)
		if flag.NoOptDefVal != "" {
			switch flag.Value.Type() {
			case "string":
				varname += fmt.Sprintf("[=\"%s\"]", flag.NoOptDefVal)
			case "bool":
				if flag.NoOptDefVal != "true" {
					varname += fmt.Sprintf("[=%s]", flag.NoOptDefVal)
				}
			case "count":
				if flag.NoOptDefVal != "+1" {
					varname += fmt.Sprintf("[=%s]", flag.NoOptDefVal)
				}
			default:
				varname += fmt.Sprintf("[=%s]", flag.NoOptDefVal)
			}
		}
		defaultZero := []string{"", "0", "[]", "<nil>", "0s", "false"}

		if !slices.Contains(defaultZero, flag.DefValue) {
			if flag.Value.Type() == "string" {
				usage += fmt.Sprintf(" (default %q)", flag.DefValue)
			} else {
				usage += fmt.Sprintf(" (default %s)", flag.DefValue)
			}
		}
		if len(flag.Deprecated) != 0 {
			usage += fmt.Sprintf(" (DEPRECATED: %s)", flag.Deprecated)
		}

		fmt.Fprintf(&b, "| %s | %s | %s |\n",
			escapeTableCell(flagName), escapeTableCell(varname), escapeTableCell(usage))
	})

	return b.String()
}

// RenderFlagsTables returns the rendered flag tables for a command's own flags
// and its inherited flags as separate strings.
func RenderFlagsTables(cmd *cobra.Command) (flags, inherited string) {
	f := cmd.NonInheritedFlags()
	if f.HasAvailableFlags() {
		flags = CommandsInTable(f)
	}
	pf := cmd.InheritedFlags()
	if pf.HasAvailableFlags() {
		inherited = CommandsInTable(pf)
	}
	return
}

// HashTitledExamples splits example lines into groups where each group starts
// with a line beginning with '#'.
func HashTitledExamples(lines []string) [][]string {
	result := make([][]string, 0)
	example := make([]string, 0)
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			result = append(result, example)
			example = make([]string, 0)
		}
		if !IsSetWithEnvVar(line) {
			example = append(example, ChoppedLineContinuation(line))
		}
	}
	result = append(result, example)
	return result[1:]
}

// IsSetWithEnvVar returns true if the line sets a flag that is typically
// provided via environment variable.
func IsSetWithEnvVar(line string) bool {
	trimmed := strings.TrimSpace(line)
	for _, prefix := range []string{"--api-token ", "--host ", "--org ", "--flow ", "--trail "} {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	return false
}

// ChoppedLineContinuation trims trailing whitespace and removes trailing backslash.
func ChoppedLineContinuation(line string) string {
	trimmed := strings.TrimRightFunc(line, unicode.IsSpace)
	return strings.TrimSuffix(trimmed, "\\")
}
