package docgen

import (
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func TestCommandsInTable(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.StringP("name", "n", "", "The name")
	fs.Bool("verbose", false, "Enable verbose")

	got := CommandsInTable(fs)
	if !strings.Contains(got, "--name") {
		t.Error("expected --name flag")
	}
	if !strings.Contains(got, "-n") {
		t.Error("expected shorthand -n")
	}
	if !strings.Contains(got, "--verbose") {
		t.Error("expected --verbose flag")
	}
	if !strings.Contains(got, "|") {
		t.Error("expected table formatting")
	}
}

func TestCommandsInTableHiddenFlags(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("visible", "", "Visible flag")
	fs.String("hidden", "", "Hidden flag")
	_ = fs.MarkHidden("hidden")

	got := CommandsInTable(fs)
	if !strings.Contains(got, "--visible") {
		t.Error("expected --visible flag")
	}
	if strings.Contains(got, "--hidden") {
		t.Error("should not contain hidden flag")
	}
}

func TestCommandsInTableDefaultValues(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("dir", "/tmp", "The directory")

	got := CommandsInTable(fs)
	if !strings.Contains(got, `(default "/tmp")`) {
		t.Errorf("expected default value, got:\n%s", got)
	}
}

func TestHashTitledExamples(t *testing.T) {
	lines := []string{
		"# first example",
		"kosli attest snyk foo",
		"# second example",
		"kosli attest snyk bar",
	}
	groups := HashTitledExamples(lines)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0][0] != "# first example" {
		t.Errorf("expected first title, got: %s", groups[0][0])
	}
	if groups[0][1] != "kosli attest snyk foo" {
		t.Errorf("expected first command, got: %s", groups[0][1])
	}
}

func TestHashTitledExamplesFiltersEnvVars(t *testing.T) {
	lines := []string{
		"# example",
		"kosli attest snyk foo",
		"	--api-token yourToken",
		"	--org yourOrg",
	}
	groups := HashTitledExamples(lines)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	// Should only have title + the kosli command (env var lines filtered)
	if len(groups[0]) != 2 {
		t.Errorf("expected 2 lines (title + command), got %d", len(groups[0]))
	}
}

func TestIsSetWithEnvVar(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"	--api-token yourToken", true},
		{"	--host https://app.kosli.com", true},
		{"	--org yourOrg", true},
		{"	--flow yourFlow", true},
		{"	--trail yourTrail", true},
		{"	--name foo", false},
		{"kosli attest snyk", false},
	}
	for _, tt := range tests {
		got := IsSetWithEnvVar(tt.line)
		if got != tt.want {
			t.Errorf("IsSetWithEnvVar(%q) = %v, want %v", tt.line, got, tt.want)
		}
	}
}

func TestChoppedLineContinuation(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"kosli attest snyk foo \\", "kosli attest snyk foo "},
		{"kosli attest snyk foo", "kosli attest snyk foo"},
		{"kosli attest snyk foo  \t", "kosli attest snyk foo"},
	}
	for _, tt := range tests {
		got := ChoppedLineContinuation(tt.input)
		if got != tt.want {
			t.Errorf("ChoppedLineContinuation(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
