package main

import (
	"io"
	"strings"
	"testing"

	"github.com/kosli-dev/cli/internal/docgen"
	"github.com/spf13/cobra"
)

func TestLifecycleEvaluateIsBeta(t *testing.T) {
	cmd := newEvaluateCmd(io.Discard)
	if !isBeta(cmd) {
		t.Error("expected evaluate command to be beta")
	}
	// subcommands inherit beta via parent walk
	for _, sub := range cmd.Commands() {
		if !isBeta(sub) {
			t.Errorf("expected subcommand %q to inherit beta", sub.Name())
		}
	}
}

func TestLifecycleAttestDecisionIsBetaAndDocHidden(t *testing.T) {
	global = &GlobalOpts{}
	cmd := newAttestDecisionCmd(io.Discard)
	if !cmd.Hidden {
		t.Error("expected attest decision to stay Hidden")
	}
	if !isBeta(cmd) {
		t.Error("expected attest decision to be beta")
	}
	if _, ok := cmd.Annotations[docgen.DocHiddenAnnotation]; !ok {
		t.Error("expected attest decision to carry the docHidden annotation")
	}
	if !isDocHidden(cmd) {
		t.Error("expected isDocHidden to be true for attest decision")
	}
}

func TestLifecycleNoBetaTextPrefix(t *testing.T) {
	global = &GlobalOpts{}
	cmds := []*cobra.Command{
		newEvaluateCmd(io.Discard),
		newAttestDecisionCmd(io.Discard),
	}
	for _, c := range cmds {
		if strings.Contains(c.Short, "[BETA]") {
			t.Errorf("command %q still has [BETA] text prefix in Short", c.Name())
		}
	}
}
