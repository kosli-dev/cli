package docgen

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommandsInTable renders a pflag.FlagSet as markdown table rows.
func CommandsInTable(f *pflag.FlagSet) string {
	buf := new(bytes.Buffer)

	lines := make([]string, 0, 100)

	f.VisitAll(func(flag *pflag.Flag) {
		// pflag.MarkDeprecated sets Hidden as a side effect, which would keep
		// deprecated-but-working flags out of the reference docs along with
		// their migration message. Only flags hidden on purpose (MarkHidden,
		// e.g. internal aliases) are skipped; deprecated ones are documented
		// below with a (DEPRECATED: ...) suffix.
		if flag.Hidden && flag.Deprecated == "" {
			return
		}

		flagName := ""
		if flag.Shorthand != "" && flag.ShorthandDeprecated == "" {
			flagName = fmt.Sprintf("  -%s, --%s", flag.Shorthand, flag.Name)
		} else {
			flagName = fmt.Sprintf("      --%s", flag.Name)
		}

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

		lines = append(lines, fmt.Sprintf("| %s | %s | %s |", flagName, varname, usage))
	})

	for _, line := range lines {
		fmt.Fprintln(buf, line)
	}

	return buf.String()
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
