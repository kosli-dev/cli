package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

// The empty-flag audit in empty-flag-audit/ runs every command with every one
// of its flags emptied. It can only do that for commands it knows about,
// and it cannot find them by reading --help: cobra leaves a Hidden or a
// Deprecated command out of every listing it prints, and pflag leaves a hidden
// flag out too. The audit found `attest override` only because someone named it
// by hand, and missed `report artifact`, `snapshot server` and `completion`
// because nobody knew to.
//
// The command tree itself has no such blind spot. This test walks it and
// compares it against the surface the audit is expected to cover, so a command
// or flag added later fails here rather than being silently unaudited.
const auditCoverageFile = "../../empty-flag-audit/coverage.json"

// The global flags are declared once on the root command and behave the same
// wherever they appear, so the audit measures them on one command rather than
// on all of them. The same command is used here.
const auditGlobalsOn = "archive flow"

// auditCoverageExempt lists commands the audit is not expected to cover.
// `docs` generates this repository's documentation and is not something a
// customer runs. `help` is cobra's own.
var auditCoverageExempt = map[string]bool{
	"docs": true,
	"help": true,
}

// localFlags returns the flags a command declares itself, hidden ones included,
// without the flags it inherits from its parents, each against the type pflag
// parses it as. The audit needs the type to invent a real value for the run it
// compares an empty value against, and a boolean is the one that has to be told
// apart: it takes no value of its own.
func localFlags(cmd *cobra.Command) map[string]string {
	flags := map[string]string{}
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if flag.Name != "help" {
			flags[flag.Name] = flag.Value.Type()
		}
	})
	return flags
}

// commandSurface walks the command tree below cmd, collecting every runnable
// command against the flags it declares. Hidden and deprecated commands are
// collected like any other, which is the whole point of walking the tree rather
// than reading a help listing.
func commandSurface(cmd *cobra.Command, path string, surface map[string]map[string]string) {
	for _, child := range cmd.Commands() {
		childPath := strings.TrimSpace(path + " " + child.Name())
		if child.Runnable() && !auditCoverageExempt[childPath] {
			surface[childPath] = localFlags(child)
		}
		commandSurface(child, childPath, surface)
	}
}

// auditableSurface returns every command the audit is expected to cover, each
// against the flags it must be run with, the global flags among those of the
// one command they are measured on.
func auditableSurface(t *testing.T) map[string]map[string]string {
	root, err := newRootCmd(io.Discard, io.Discard, nil)
	require.NoError(t, err)

	surface := map[string]map[string]string{}
	commandSurface(root, "", surface)

	require.Contains(t, surface, auditGlobalsOn,
		"the command the global flags are measured on must exist")
	root.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		surface[auditGlobalsOn][flag.Name] = flag.Value.Type()
	})

	return surface
}

// combinations renders a surface as one "command --flag" line per pair, which
// is the unit the audit measures and so the unit a difference is reported in.
// A command declaring no flags is a line of its own, so that adding one cannot
// go unnoticed just because it has nothing to empty.
func combinations(surface map[string]map[string]string) map[string]bool {
	pairs := map[string]bool{}
	for command, flags := range surface {
		if len(flags) == 0 {
			pairs[command] = true
		}
		for flag := range flags {
			pairs[fmt.Sprintf("%s --%s", command, flag)] = true
		}
	}
	return pairs
}

// missingFrom returns the members of want that have are absent from got.
func missingFrom(want, got map[string]bool) []string {
	absent := []string{}
	for pair := range want {
		if !got[pair] {
			absent = append(absent, pair)
		}
	}
	sort.Strings(absent)
	return absent
}

// auditCoverageUpdateVar names the variable that rewrites coverage.json from
// the command tree instead of checking against it:
//
//	UPDATE_AUDIT_COVERAGE=1 go test ./cmd/kosli -run TestEmptyFlagAudit
//
// Rewriting it says the new command or flag is real and the audit must cover
// it. The audit reads the same file and refuses to run until it does, so the
// rewrite records the work rather than dismissing it.
const auditCoverageUpdateVar = "UPDATE_AUDIT_COVERAGE"

// TestEmptyFlagAuditCoversEveryCommandAndFlag pins the audit's coverage to the
// command tree, so that adding a command or a flag without extending the audit
// fails here.
func TestEmptyFlagAuditCoversEveryCommandAndFlag(t *testing.T) {
	if os.Getenv(auditCoverageUpdateVar) != "" {
		content, err := json.MarshalIndent(auditableSurface(t), "", " ")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(auditCoverageFile, append(content, '\n'), 0644))
		t.Logf("rewrote %s from the command tree", auditCoverageFile)
		return
	}

	content, err := os.ReadFile(auditCoverageFile)
	require.NoError(t, err)
	var covered map[string]map[string]string
	require.NoError(t, json.Unmarshal(content, &covered))

	real := combinations(auditableSurface(t))
	claimed := combinations(covered)

	unaudited := missingFrom(real, claimed)
	require.Empty(t, unaudited, "these exist but the audit does not cover them:\n%s",
		strings.Join(unaudited, "\n"))

	gone := missingFrom(claimed, real)
	require.Empty(t, gone, "the audit covers these but they no longer exist:\n%s",
		strings.Join(gone, "\n"))
}
