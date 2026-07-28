package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// normalizeBoolFlagArgs rewrites boolean flags written in the space form
// ("--compliant false") into the "=" form ("--compliant=false"), which is the
// only form pflag accepts for an explicitly-valued boolean flag. All other
// tokens are returned unchanged, as is everything following a "--" terminator.
//
// The rewrite is positional: it matches on the tokens themselves and does not
// track which of them pflag will consume as the value of an earlier flag. Two
// pathological inputs are therefore read differently from the way pflag reads
// them, and both are accepted. A positional argument literally named "true" or
// "false" directly after a bare boolean flag becomes that flag's value. And in
// "--name --compliant false", where pflag gives --name the value "--compliant"
// and leaves "false" positional, this rewrite instead produces
// "--compliant=false". Kosli positionals are artifact names, fingerprints and
// file paths, and flag values are not flag tokens, so both inputs are far
// enough outside real usage to be worth the plain forms this rescues.
// The rewrite runs in two passes because the flags in scope depend on where a
// token sits. A flag before the subcommand can only be one of root's own, and
// root.Find cannot resolve the command while such a flag is still in the space
// form: the value is left as a stray positional, which cobra reads as an unknown
// subcommand. The first pass therefore works from root's flags alone, which is
// what lets the second pass resolve the command and reach the flags declared on
// it. Joining is idempotent, since a joined token no longer matches any flag
// token, so the passes cannot rewrite the same token twice.
func normalizeBoolFlagArgs(root *cobra.Command, args []string) []string {
	args = joinBoolFlagValues(boolFlagTokens(root.LocalFlags()), args)
	cmd, _, err := root.Find(args)
	if err != nil {
		return args
	}
	return joinBoolFlagValues(boolFlagTokens(cmd.Flags()), args)
}

// joinBoolFlagValues rewrites every "<flag> <literal>" pair in args, where flag
// is a token in boolFlags and literal is a boolean literal, into the single
// "<flag>=<literal>" token. All other tokens are returned unchanged, as is
// everything following a "--" terminator.
func joinBoolFlagValues(boolFlags map[string]bool, args []string) []string {
	normalized := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			// pflag stops parsing flags here, so everything left is a
			// positional argument and must be passed through untouched.
			normalized = append(normalized, args[i:]...)
			break
		}
		if boolFlags[args[i]] && i+1 < len(args) && isBoolLiteral(args[i+1]) {
			normalized = append(normalized, args[i]+"="+args[i+1])
			i++
			continue
		}
		normalized = append(normalized, args[i])
	}
	return normalized
}

// boolFlagTokens returns the command-line tokens ("--compliant", "-C") of every
// boolean flag in flags.
//
// Shorthands are listed individually, so a grouped token such as "-qC" is not a
// key here and "-qC false" is left alone: it keeps failing with the arg-count
// error and its link to the boolean flags FAQ. That is a deliberate choice.
// Grouping raises questions this rewrite has no good answer to, such as which
// member of the group the value belongs to and what to do when a group mixes
// boolean and non-boolean shorthands. Rescuing the plain forms is worth a
// low-level rewrite; guessing intent inside a grouped token is not.
func boolFlagTokens(flags *pflag.FlagSet) map[string]bool {
	tokens := map[string]bool{}
	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Value.Type() != "bool" {
			return
		}
		tokens["--"+flag.Name] = true
		if flag.Shorthand != "" {
			tokens["-"+flag.Shorthand] = true
		}
	})
	return tokens
}

// isBoolLiteral reports whether token is one of the two values a rewritten
// boolean flag may be given.
func isBoolLiteral(token string) bool {
	return token == "true" || token == "false"
}
