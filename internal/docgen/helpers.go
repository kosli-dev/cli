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

	maxlen := 0
	f.VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}

		line := ""
		if flag.Shorthand != "" && flag.ShorthandDeprecated == "" {
			line = fmt.Sprintf("  -%s, --%s", flag.Shorthand, flag.Name)
		} else {
			line = fmt.Sprintf("      --%s", flag.Name)
		}

		varname, usage := pflag.UnquoteUsage(flag)
		if varname != "" {
			line += " " + varname
		}
		if flag.NoOptDefVal != "" {
			switch flag.Value.Type() {
			case "string":
				line += fmt.Sprintf("[=\"%s\"]", flag.NoOptDefVal)
			case "bool":
				if flag.NoOptDefVal != "true" {
					line += fmt.Sprintf("[=%s]", flag.NoOptDefVal)
				}
			case "count":
				if flag.NoOptDefVal != "+1" {
					line += fmt.Sprintf("[=%s]", flag.NoOptDefVal)
				}
			default:
				line += fmt.Sprintf("[=%s]", flag.NoOptDefVal)
			}
		}

		line += "\x00"
		if len(line) > maxlen {
			maxlen = len(line)
		}

		line += usage
		defaultZero := []string{"", "0", "[]", "<nil>", "0s", "false"}

		if !slices.Contains(defaultZero, flag.DefValue) {
			if flag.Value.Type() == "string" {
				line += fmt.Sprintf(" (default %q)", flag.DefValue)
			} else {
				line += fmt.Sprintf(" (default %s)", flag.DefValue)
			}
		}
		if len(flag.Deprecated) != 0 {
			line += fmt.Sprintf(" (DEPRECATED: %s)", flag.Deprecated)
		}

		lines = append(lines, line)
	})

	for _, line := range lines {
		sidx := strings.Index(line, "\x00")
		fmt.Fprintln(buf, "| ", line[:sidx], " | ", line[sidx+1:], " |")
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
