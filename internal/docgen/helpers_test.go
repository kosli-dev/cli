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
	// pflag leaves the type empty for booleans; the table names it anyway so a
	// blank cell never leaves the reader guessing what the flag takes.
	if !strings.Contains(got, "| --verbose | bool | Enable verbose |") {
		t.Errorf("expected bool type column for bool flag, got:\n%s", got)
	}
	if strings.Contains(got, "|  |") {
		t.Errorf("expected no empty type cell, got:\n%s", got)
	}
}

// TestCommandsInTableBoolOptionalValue checks the bool naming lands before the
// optional-value suffix, so it reads like the string case (string[="x"]).
func TestCommandsInTableBoolOptionalValue(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.Bool("verbose", false, "Enable verbose")
	fs.Lookup("verbose").NoOptDefVal = "maybe"

	got := CommandsInTable(fs)
	if !strings.Contains(got, `| --verbose | bool[=maybe] | Enable verbose |`) {
		t.Errorf("expected bool[=maybe] in the type column, got:\n%s", got)
	}
}

// unescapedPipes counts only the pipes markdown reads as column separators,
// i.e. those not escaped as \|.
func unescapedPipes(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '|' && (i == 0 || s[i-1] != '\\') {
			n++
		}
	}
	return n
}

// TestCommandsInTableColumnCount guards the contract between the rows rendered
// here and the header in FlagTableHeader: both must have three columns. The
// pipe-carrying flags matter most — an unescaped pipe in a description or a
// default would silently split the row into extra columns.
func TestCommandsInTableColumnCount(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.StringP("name", "n", "", "The name")
	fs.Bool("verbose", false, "Enable verbose")
	fs.Int("count", 0, "How many")
	fs.String("age", "", "One of <hours|days|weeks|months>")
	fs.String("sep", "a|b", "The separator")

	wantCells := unescapedPipes(strings.SplitN(FlagTableHeader, "\n", 2)[0])
	rows := strings.Split(strings.TrimSpace(CommandsInTable(fs)), "\n")
	if len(rows) != 5 {
		t.Fatalf("expected 5 rows, got %d:\n%s", len(rows), strings.Join(rows, "\n"))
	}
	for _, row := range rows {
		if got := unescapedPipes(row); got != wantCells {
			t.Errorf("row %q has %d separators, header has %d", row, got, wantCells)
		}
	}
}

// TestCommandsInTableEscapesPipes covers where the pipes can come from: the
// description itself and a default value. No Kosli flag help contains a pipe
// today, so this is the guard that keeps it that way.
func TestCommandsInTableEscapesPipes(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("age", "", "One of <hours|days|weeks|months>")
	fs.String("sep", "a|b", "The separator")

	got := CommandsInTable(fs)
	if !strings.Contains(got, `| --age | string | One of <hours\|days\|weeks\|months> |`) {
		t.Errorf("expected pipes in the description to be escaped, got:\n%s", got)
	}
	if !strings.Contains(got, `| --sep | string | The separator (default "a\|b") |`) {
		t.Errorf("expected pipes in the default to be escaped, got:\n%s", got)
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
