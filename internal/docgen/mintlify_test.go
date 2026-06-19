package docgen

import (
	"strings"
	"testing"
)

func TestMintlifyFrontMatter(t *testing.T) {
	f := MintlifyFormatter{}
	meta := CommandMeta{Name: "kosli attest snyk", Summary: "Report a snyk attestation"}
	got := f.FrontMatter(meta)
	if !strings.Contains(got, `title: "kosli attest snyk"`) {
		t.Errorf("expected title, got:\n%s", got)
	}
	if !strings.Contains(got, `description: "Report a snyk attestation"`) {
		t.Errorf("expected description, got:\n%s", got)
	}
}

func TestMintlifyFrontMatterSanitizesDescription(t *testing.T) {
	f := MintlifyFormatter{}
	meta := CommandMeta{Name: "cmd", Summary: "Use ^foo^ with \"quotes\""}
	got := f.FrontMatter(meta)
	if !strings.Contains(got, "description: \"Use `foo` with 'quotes'\"") {
		t.Errorf("expected sanitized description, got:\n%s", got)
	}
}

func TestMintlifyFrontMatterTruncatesLongDescription(t *testing.T) {
	f := MintlifyFormatter{}
	long := strings.Repeat("a", 250)
	meta := CommandMeta{Name: "cmd", Summary: long}
	got := f.FrontMatter(meta)
	if !strings.Contains(got, "...") {
		t.Error("expected truncated description")
	}
}

func TestMintlifyFrontMatterBetaTag(t *testing.T) {
	f := MintlifyFormatter{}
	got := f.FrontMatter(CommandMeta{Name: "kosli evaluate", Beta: true})
	if !strings.Contains(got, `tag: "BETA"`) {
		t.Errorf("expected BETA tag, got:\n%s", got)
	}
}

func TestMintlifyFrontMatterDeprecatedTag(t *testing.T) {
	f := MintlifyFormatter{}
	got := f.FrontMatter(CommandMeta{Name: "kosli report approval", Deprecated: true})
	if !strings.Contains(got, `tag: "DEPRECATED"`) {
		t.Errorf("expected DEPRECATED tag, got:\n%s", got)
	}
}

func TestMintlifyFrontMatterDeprecatedWinsOverBeta(t *testing.T) {
	f := MintlifyFormatter{}
	got := f.FrontMatter(CommandMeta{Name: "cmd", Beta: true, Deprecated: true})
	if !strings.Contains(got, `tag: "DEPRECATED"`) || strings.Contains(got, `tag: "BETA"`) {
		t.Errorf("expected DEPRECATED to win, got:\n%s", got)
	}
}

func TestMintlifyFrontMatterHidden(t *testing.T) {
	f := MintlifyFormatter{}
	got := f.FrontMatter(CommandMeta{Name: "kosli attest decision", Hidden: true})
	if !strings.Contains(got, "hidden: true") {
		t.Errorf("expected hidden: true, got:\n%s", got)
	}
}

func TestMintlifyFrontMatterNormalHasNoTagOrHidden(t *testing.T) {
	f := MintlifyFormatter{}
	got := f.FrontMatter(CommandMeta{Name: "kosli attest snyk"})
	if strings.Contains(got, "tag:") {
		t.Errorf("expected no tag for normal command, got:\n%s", got)
	}
	if strings.Contains(got, "hidden:") {
		t.Errorf("expected no hidden key for normal command, got:\n%s", got)
	}
}

func TestMintlifyBetaWarning(t *testing.T) {
	f := MintlifyFormatter{}
	got := f.BetaWarning("kosli evaluate")
	if !strings.Contains(got, `import CliBetaNotice from "/snippets/cli-beta-notice.mdx";`) {
		t.Errorf("expected beta snippet import, got:\n%s", got)
	}
	if !strings.Contains(got, "<CliBetaNotice />") {
		t.Errorf("expected beta snippet component, got:\n%s", got)
	}
	if strings.Contains(got, "<Warning>") {
		t.Errorf("notice prose should live in the snippet, not the generator, got:\n%s", got)
	}
}

func TestMintlifyTutorialTip(t *testing.T) {
	f := MintlifyFormatter{}
	got := f.TutorialTip("https://docs.kosli.com/tutorials/snyk")
	if !strings.Contains(got, "<Tip>") || !strings.Contains(got, "</Tip>") {
		t.Errorf("expected Tip component, got:\n%s", got)
	}
	if !strings.Contains(got, "[tutorial](https://docs.kosli.com/tutorials/snyk)") {
		t.Errorf("expected markdown link to tutorial URL, got:\n%s", got)
	}
}

func TestMintlifyDeprecatedWarning(t *testing.T) {
	f := MintlifyFormatter{}
	got := f.DeprecatedWarning("kosli snapshot server", "use 'kosli snapshot paths' instead")
	if !strings.Contains(got, `import CliDeprecatedNotice from "/snippets/cli-deprecated-notice.mdx";`) {
		t.Errorf("expected deprecated snippet import, got:\n%s", got)
	}
	if !strings.Contains(got, "<CliDeprecatedNotice />") {
		t.Errorf("expected deprecated snippet component, got:\n%s", got)
	}
	if !strings.Contains(got, "use 'kosli snapshot paths' instead") {
		t.Errorf("expected migration message as plain text, got:\n%s", got)
	}
	if strings.Contains(got, "<Warning>") {
		t.Errorf("notice prose should live in the snippet, not the generator, got:\n%s", got)
	}
}

func TestMintlifyDeprecatedWarningEmptyMessage(t *testing.T) {
	f := MintlifyFormatter{}
	got := f.DeprecatedWarning("cmd", "")
	if !strings.Contains(got, "<CliDeprecatedNotice />") {
		t.Errorf("expected snippet component, got:\n%s", got)
	}
}

func TestMintlifySynopsis(t *testing.T) {
	f := MintlifyFormatter{}
	meta := CommandMeta{
		Long:     "Report a ^snyk^ attestation.",
		UseLine:  "snyk [IMAGE-NAME] [flags]",
		Runnable: true,
	}
	got := f.Synopsis(meta)
	if !strings.Contains(got, "## Synopsis") {
		t.Error("expected Synopsis heading")
	}
	if !strings.Contains(got, "Report a `snyk` attestation.") {
		t.Error("expected carets replaced")
	}
}

func TestMintlifyExampleUseCases(t *testing.T) {
	f := MintlifyFormatter{}
	example := "# report a snyk attestation\nkosli attest snyk foo"
	got := f.ExampleUseCases("kosli attest snyk", example)
	if !strings.Contains(got, "<AccordionGroup>") {
		t.Error("expected AccordionGroup component")
	}
	if !strings.Contains(got, `<Accordion title="report a snyk attestation">`) {
		t.Error("expected Accordion with title")
	}
	if !strings.Contains(got, "</AccordionGroup>") {
		t.Error("expected closing AccordionGroup")
	}
	if strings.Contains(got, "https://docs.kosli.com") {
		t.Error("expected bare docs.kosli.com URLs to be linkified")
	}
}

func TestMintlifyLinkHandler(t *testing.T) {
	f := MintlifyFormatter{}
	got := f.LinkHandler("kosli_attest_snyk.md")
	if got != "/client_reference/kosli_attest_snyk" {
		t.Errorf("expected no trailing slash, got: %s", got)
	}
}

func TestLinkifyKosliDocsURLs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "bare URL becomes markdown link",
			input: "A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)",
			want:  "A boolean flag [docs](/faq/#boolean-flags) (default false)",
		},
		{
			name:  "URL followed by comma",
			input: "(defaulted in some CIs: https://docs.kosli.com/ci-defaults, otherwise defaults to HEAD ).",
			want:  "(defaulted in some CIs: [docs](/ci-defaults), otherwise defaults to HEAD ).",
		},
		{
			name:  "URL followed by space and closing paren",
			input: "(defaulted in some CIs: https://docs.kosli.com/ci-defaults ).",
			want:  "(defaulted in some CIs: [docs](/ci-defaults) ).",
		},
		{
			name:  "long path URL",
			input: "see https://docs.kosli.com/integrations/ci_cd/#defaulted-kosli-command-flags-from-ci-variables .",
			want:  "see [docs](/integrations/ci_cd/#defaulted-kosli-command-flags-from-ci-variables) .",
		},
		{
			name:  "non-kosli URL untouched",
			input: "see https://example.com/foo for details",
			want:  "see https://example.com/foo for details",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := linkifyKosliDocsURLs(tt.input)
			if got != tt.want {
				t.Errorf("linkifyKosliDocsURLs(%q):\ngot:  %q\nwant: %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEscapeMintlifyProse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "curly braces escaped",
			input: "Use {expression} here",
			want:  "Use \\{expression\\} here",
		},
		{
			name:  "uppercase angle brackets converted",
			input: "Use <IMAGE-NAME> as input",
			want:  "Use `IMAGE-NAME` as input",
		},
		{
			name:  "lowercase HTML tags preserved",
			input: "Use <a href=\"x\">link</a>",
			want:  "Use <a href=\"x\">link</a>",
		},
		{
			name:  "lowercase angle brackets with pipes converted",
			input: "Use <hours|days|weeks|months> for time",
			want:  "Use `hours|days|weeks|months` for time",
		},
		{
			name:  "uppercase angle brackets with pipes converted",
			input: "Format: <COMMIT_SHA1|ARTIFACT_FINGERPRINT>",
			want:  "Format: `COMMIT_SHA1|ARTIFACT_FINGERPRINT`",
		},
		{
			name:  "lowercase single-word placeholders converted",
			input: "flowName@<fingerprint> or flowName:<commit_sha>",
			want:  "flowName@`fingerprint` or flowName:`commit_sha`",
		},
		{
			name:  "double curly braces escaped",
			input: "--jira-secondary-source ${{ github.head_ref }}",
			want:  "--jira-secondary-source $\\{\\{ github.head_ref \\}\\}",
		},
		{
			name:  "code fence content not escaped",
			input: "text ```\n{code}\n<FOO>\n``` more {text}",
			want:  "text ```\n{code}\n<FOO>\n``` more \\{text\\}",
		},
		{
			name:  "single-quoted URL converted to backtick code",
			input: "e.g. 'http://proxy-server-ip:proxy-port'",
			want:  "e.g. `http://proxy-server-ip:proxy-port`",
		},
		{
			name:  "single-quoted https URL converted to backtick code",
			input: "use 'https://example.com:8080/path'",
			want:  "use `https://example.com:8080/path`",
		},
		{
			name:  "non-URL single quotes left alone",
			input: "e.g. 'some text'",
			want:  "e.g. 'some text'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeMintlifyProse(tt.input)
			if got != tt.want {
				t.Errorf("escapeMintlifyProse(%q):\ngot:  %q\nwant: %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBacktickFlags(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "double-dash flag at start",
			input: "--api-token",
			want:  "`--api-token`",
		},
		{
			name:  "single-dash flag at start",
			input: "-x",
			want:  "`-x`",
		},
		{
			name:  "flag in middle of sentence",
			input: "use --api-token to authenticate",
			want:  "use `--api-token` to authenticate",
		},
		{
			name:  "flag in table cell",
			input: "| --api-token | API token |",
			want:  "| `--api-token` | API token |",
		},
		{
			name:  "two adjacent flags",
			input: "--foo --bar",
			want:  "`--foo` `--bar`",
		},
		{
			name:  "comma-separated flags",
			input: "--foo,--bar",
			want:  "`--foo`,`--bar`",
		},
		{
			name:  "already-backticked flag is left alone",
			input: "use `--api-token` here",
			want:  "use `--api-token` here",
		},
		{
			name:  "hyphenated word is not a flag",
			input: "this is built-in behaviour",
			want:  "this is built-in behaviour",
		},
		{
			name:  "code fence content is untouched",
			input: "see ```\nkosli foo --bar\n``` for example",
			want:  "see ```\nkosli foo --bar\n``` for example",
		},
		{
			name:  "flag with hyphen in name",
			input: "use --jira-base-url here",
			want:  "use `--jira-base-url` here",
		},
		{
			name:  "short flag in middle",
			input: "use -h for help",
			want:  "use `-h` for help",
		},
		{
			name:  "bare double-dash is not a flag",
			input: "use -- to separate flags from args",
			want:  "use -- to separate flags from args",
		},
		{
			name:  "flag with equals value",
			input: "use --output=json here",
			want:  "use `--output`=json here",
		},
		{
			name:  "numeric arg is not a flag",
			input: "use -1 for default",
			want:  "use -1 for default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := backtickFlags(tt.input)
			if got != tt.want {
				t.Errorf("backtickFlags(%q):\ngot:  %q\nwant: %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMintlifyFlagsSectionBackticksFlags(t *testing.T) {
	f := MintlifyFormatter{}
	flags := "| --api-token string | required (default $KOSLI_API_TOKEN) |\n"
	got := f.FlagsSection(flags, "")
	if !strings.Contains(got, "`--api-token`") {
		t.Errorf("expected --api-token to be backticked, got:\n%s", got)
	}
	if strings.Contains(got, "| --api-token string |") {
		t.Errorf("expected bare --api-token to be replaced, got:\n%s", got)
	}
}

func TestSanitizeDescription(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "carets to backticks",
			input: "Use ^foo^ bar",
			want:  "Use `foo` bar",
		},
		{
			name:  "quotes escaped",
			input: `Say "hello"`,
			want:  "Say 'hello'",
		},
		{
			name:  "truncation",
			input: strings.Repeat("x", 250),
			want:  strings.Repeat("x", 197) + "...",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeDescription(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeDescription(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
