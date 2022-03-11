package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

// // PrintDefaults prints, to standard error unless configured
// // otherwise, the default values of all defined flags in the set.
// func PrintDefaults(f *pflag.FlagSet, out io.Writer) {
// 	usages := f.FlagUsages()
// 	fmt.Fprint(out, usages)
// }

// // FlagUsages returns a string containing the usage information for all flags in
// // the FlagSet
// func (f *FlagSet) FlagUsages() string {
// 	return f.FlagUsagesWrapped(0)
// }

// FlagUsagesWrapped returns a string containing the usage information
// for all flags in the FlagSet. Wrapped to `cols` columns (0 for no
// wrapping)
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

		// This special character will be replaced with spacing once the
		// correct alignment is calculated
		line += "\x00"
		if len(line) > maxlen {
			maxlen = len(line)
		}

		line += usage
		// if !flag.defaultIsZeroValue() {
		if flag.Value.Type() == "string" {
			line += fmt.Sprintf(" (default %q)", flag.DefValue)
		} else {
			line += fmt.Sprintf(" (default %s)", flag.DefValue)
		}
		// }
		if len(flag.Deprecated) != 0 {
			line += fmt.Sprintf(" (DEPRECATED: %s)", flag.Deprecated)
		}

		lines = append(lines, line)
	})

	fmt.Printf("printed lines: %v", lines)
	for _, line := range lines {
		sidx := strings.Index(line, "\x00")
		// spacing := strings.Repeat(" ", maxlen-sidx)
		// maxlen + 2 comes from + 1 for the \x00 and + 1 for the (deliberate) off-by-one in maxlen-sidx
		// fmt.Fprintln(buf, line[:sidx], spacing, wrap(maxlen+2, 0, line[sidx+1:]))
		fmt.Fprintln(buf, "| ", line[:sidx], " | ", line[sidx+1:], " |")
	}

	return buf.String()
}
