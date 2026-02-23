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

func TestMintlifyBetaWarning(t *testing.T) {
	f := MintlifyFormatter{}
	got := f.BetaWarning("kosli foo")
	if !strings.Contains(got, "<Warning>") {
		t.Error("expected Warning component")
	}
	if !strings.Contains(got, "</Warning>") {
		t.Error("expected closing Warning component")
	}
	if !strings.Contains(got, "**kosli foo** is a beta feature") {
		t.Error("expected command name in warning")
	}
}

func TestMintlifyDeprecatedWarning(t *testing.T) {
	f := MintlifyFormatter{}
	got := f.DeprecatedWarning("kosli artifact", "see kosli attest commands")
	if !strings.Contains(got, "<Warning>") {
		t.Error("expected Warning component")
	}
	if !strings.Contains(got, "**kosli artifact** is deprecated") {
		t.Error("expected deprecation message")
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

func TestMintlifyLiveCIExamples(t *testing.T) {
	f := MintlifyFormatter{}
	examples := []CIExample{
		{CI: "GitHub", YamlURL: "http://yaml", EventURL: "http://event"},
		{CI: "GitLab", YamlURL: "http://yaml2"},
	}
	got := f.LiveCIExamples(examples, "kosli attest snyk")
	if !strings.Contains(got, "<Tabs>") {
		t.Error("expected Tabs component")
	}
	if !strings.Contains(got, `<Tab title="GitHub">`) {
		t.Error("expected GitHub tab")
	}
	if !strings.Contains(got, `<Tab title="GitLab">`) {
		t.Error("expected GitLab tab")
	}
	if !strings.Contains(got, "</Tabs>") {
		t.Error("expected closing Tabs")
	}
}

func TestMintlifyLiveCLIExample(t *testing.T) {
	f := MintlifyFormatter{}
	got := f.LiveCLIExample("kosli list environments", "kosli list environments --output=json", "http://example.com")
	// Should NOT contain Hugo shortcode wrappers
	if strings.Contains(got, "{{< raw-html >}}") {
		t.Error("should not contain Hugo shortcode")
	}
	if !strings.Contains(got, "<pre>") {
		t.Error("expected raw HTML pre tag")
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
			name:  "code fence content not escaped",
			input: "text ```\n{code}\n<FOO>\n``` more {text}",
			want:  "text ```\n{code}\n<FOO>\n``` more \\{text\\}",
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
