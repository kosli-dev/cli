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
	if !strings.Contains(got, "| -n, --name | string | The name |") {
		t.Errorf("expected string type in its own column, got:\n%s", got)
	}
	if !strings.Contains(got, "| --verbose |  | Enable verbose |") {
		t.Errorf("expected empty type column for bool flag, got:\n%s", got)
	}
}

// TestCommandsInTableColumnCount guards the contract between the rows rendered
// here and the header in FlagTableHeader: both must have three columns.
func TestCommandsInTableColumnCount(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.StringP("name", "n", "", "The name")
	fs.Bool("verbose", false, "Enable verbose")
	fs.Int("count", 0, "How many")

	wantCells := strings.Count(strings.SplitN(FlagTableHeader, "\n", 2)[0], "|")
	for _, row := range strings.Split(strings.TrimSpace(CommandsInTable(fs)), "\n") {
		if got := strings.Count(row, "|"); got != wantCells {
			t.Errorf("row %q has %d pipes, header has %d", row, got, wantCells)
		}
	}
}

// TestCommandsInTableOptionalValue documents where the optional-value syntax
// from NoOptDefVal lands: in the type column, next to the type it modifies.
// No Kosli flag uses a non-bool NoOptDefVal today, so this pins the behaviour
// before something starts relying on it.
func TestCommandsInTableOptionalValue(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("colour", "", "When to colourise")
	fs.Lookup("colour").NoOptDefVal = "always"

	got := CommandsInTable(fs)
	if !strings.Contains(got, `| --colour | string[="always"] | When to colourise |`) {
		t.Errorf("expected optional-value syntax in the type column, got:\n%s", got)
	}
}

// TestCommandsInTableHiddenFlags and TestCommandsInTableDeprecatedFlags are a
// pair: a flag hidden via MarkHidden stays out of the docs, while a flag hidden
// as a side effect of MarkDeprecated is documented with its migration message.
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

func TestCommandsInTableDeprecatedFlags(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("new", "", "The replacement flag")
	fs.String("old", "", "The superseded flag")
	// MarkDeprecated also sets Hidden, which used to keep the flag out of the docs.
	_ = fs.MarkDeprecated("old", "use --new instead")

	got := CommandsInTable(fs)
	// Asserted as a whole row: the deprecation notice belongs in the
	// description, not in the type column.
	if !strings.Contains(got, "| --old | string | The superseded flag (DEPRECATED: use --new instead) |") {
		t.Errorf("expected deprecation message in the description, got:\n%s", got)
	}
	if !strings.Contains(got, "--new") {
		t.Errorf("expected replacement flag, got:\n%s", got)
	}
}

func TestCommandsInTableDefaultValues(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("dir", "/tmp", "The directory")

	got := CommandsInTable(fs)
	// Asserted as a whole row: the default belongs in the description, not in
	// the type column.
	if !strings.Contains(got, `| --dir | string | The directory (default "/tmp") |`) {
		t.Errorf("expected default value in the description, got:\n%s", got)
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
