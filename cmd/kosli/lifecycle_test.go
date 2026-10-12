package main

import (
	"io"
	"path/filepath"
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

func TestLifecycleAttestSbomIsNotBeta(t *testing.T) {
	global = &GlobalOpts{}
	cmd := newAttestSbomCmd(io.Discard)
	if isBeta(cmd) {
		t.Error("expected attest sbom to no longer be beta")
	}
	if cmd.Hidden {
		t.Error("expected attest sbom to be visible (not Hidden)")
	}
}

func TestLifecycleAttestDecisionIsBetaAndVisible(t *testing.T) {
	global = &GlobalOpts{}
	cmd := newAttestDecisionCmd(io.Discard)
	if cmd.Hidden {
		t.Error("expected attest decision to be visible (not Hidden)")
	}
	if !isBeta(cmd) {
		t.Error("expected attest decision to be beta")
	}
	if _, ok := cmd.Annotations[docgen.DocHiddenAnnotation]; ok {
		t.Error("expected attest decision to no longer carry the docHidden annotation")
	}
	if isDocHidden(cmd) {
		t.Error("expected isDocHidden to be false for attest decision")
	}
}

// Hidden while the server gates the endpoint behind is-server-side-evaluation-enabled.
func TestLifecycleCreateRegoPolicyIsBetaAndHidden(t *testing.T) {
	global = &GlobalOpts{}
	cmd := newCreateRegoPolicyCmd(io.Discard)
	if !cmd.Hidden {
		t.Error("expected create rego-policy to be Hidden")
	}
	if !isBeta(cmd) {
		t.Error("expected create rego-policy to be beta")
	}

	dir := t.TempDir()
	if _, _, _, _, err := executeCommandC("docs --dir " + dir); err != nil {
		t.Fatalf("docs generation failed: %v", err)
	}
	pages, err := filepath.Glob(filepath.Join(dir, "*rego-policy*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) > 0 {
		t.Errorf("expected no docs page for a hidden rego-policy command, got %v", pages)
	}
}

func TestLifecycleSnapshotGKEIsBetaAndVisible(t *testing.T) {
	cmd := newSnapshotGKECmd(io.Discard)
	if cmd.Hidden {
		t.Error("expected snapshot gke to be visible (not Hidden)")
	}
	if !isBeta(cmd) {
		t.Error("expected snapshot gke to be beta")
	}
}

func TestLifecycleControlCommandsAreBeta(t *testing.T) {
	global = &GlobalOpts{}
	cmds := map[string]*cobra.Command{
		"create control":    newCreateControlCmd(io.Discard),
		"list controls":     newListControlsCmd(io.Discard),
		"get control":       newGetControlCmd(io.Discard),
		"update control":    newUpdateControlCmd(io.Discard),
		"archive control":   newArchiveControlCmd(io.Discard),
		"unarchive control": newUnarchiveControlCmd(io.Discard),
	}
	for name, cmd := range cmds {
		if !isBeta(cmd) {
			t.Errorf("expected %q to be marked beta while controls is behind a feature flag", name)
		}
	}
}

func TestDeprecationHint(t *testing.T) {
	generic := &cobra.Command{Use: "x", Deprecated: deprecatedCommandMsg}
	if got := deprecationHint(generic); got != "" {
		t.Errorf("expected generic deprecation message suppressed, got %q", got)
	}
	custom := &cobra.Command{Use: "y", Deprecated: "use 'kosli snapshot paths' instead"}
	if got := deprecationHint(custom); got != "use 'kosli snapshot paths' instead" {
		t.Errorf("expected custom hint preserved, got %q", got)
	}
	none := &cobra.Command{Use: "z"}
	if got := deprecationHint(none); got != "" {
		t.Errorf("expected empty for non-deprecated command, got %q", got)
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
