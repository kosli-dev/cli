package docgen

import (
	"strings"
	"testing"
)

func TestHugoFrontMatter(t *testing.T) {
	f := HugoFormatter{}
	meta := CommandMeta{Name: "kosli attest snyk", Summary: "Report a snyk attestation"}
	got := f.FrontMatter(meta)
	if !strings.Contains(got, `title: "kosli attest snyk"`) {
		t.Errorf("expected title in front matter, got:\n%s", got)
	}
	if !strings.Contains(got, "beta: false") {
		t.Errorf("expected beta: false, got:\n%s", got)
	}
	if !strings.Contains(got, `summary: "Report a snyk attestation"`) {
		t.Errorf("expected summary, got:\n%s", got)
	}
}

func TestHugoBetaWarning(t *testing.T) {
	f := HugoFormatter{}
	got := f.BetaWarning("kosli foo")
	if !strings.Contains(got, "{{% hint warning %}}") {
		t.Error("expected Hugo hint warning shortcode")
	}
	if !strings.Contains(got, "**kosli foo** is a beta feature") {
		t.Error("expected command name in warning")
	}
}

func TestHugoDeprecatedWarning(t *testing.T) {
	f := HugoFormatter{}
	got := f.DeprecatedWarning("kosli artifact", "see kosli attest commands")
	if !strings.Contains(got, "{{% hint danger %}}") {
		t.Error("expected Hugo hint danger shortcode")
	}
	if !strings.Contains(got, "**kosli artifact** is deprecated. see kosli attest commands") {
		t.Error("expected deprecation message")
	}
}

func TestHugoSynopsis(t *testing.T) {
	f := HugoFormatter{}
	meta := CommandMeta{
		Long:     "Report a ^snyk^ attestation.",
		UseLine:  "snyk [IMAGE-NAME] [flags]",
		Runnable: true,
	}
	got := f.Synopsis(meta)
	if !strings.Contains(got, "## Synopsis") {
		t.Error("expected Synopsis heading")
	}
	if !strings.Contains(got, "```shell\nsnyk [IMAGE-NAME] [flags]\n```") {
		t.Error("expected shell code block with usage line")
	}
	if !strings.Contains(got, "Report a `snyk` attestation.") {
		t.Error("expected carets replaced with backticks")
	}
}

func TestHugoSynopsisNotRunnable(t *testing.T) {
	f := HugoFormatter{}
	meta := CommandMeta{Long: "Some description", Runnable: false}
	got := f.Synopsis(meta)
	if strings.Contains(got, "```shell") {
		t.Error("should not contain code block for non-runnable command")
	}
}

func TestHugoLiveCIExamples(t *testing.T) {
	f := HugoFormatter{}
	examples := []CIExample{
		{CI: "GitHub", YamlURL: "http://yaml", EventURL: "http://event"},
	}
	got := f.LiveCIExamples(examples, "kosli attest snyk")
	if !strings.Contains(got, `{{< tabs "live-examples"`) {
		t.Error("expected Hugo tabs shortcode")
	}
	if !strings.Contains(got, `{{< tab "GitHub" >}}`) {
		t.Error("expected GitHub tab")
	}
}

func TestHugoLiveCIExamplesEmpty(t *testing.T) {
	f := HugoFormatter{}
	got := f.LiveCIExamples(nil, "cmd")
	if got != "" {
		t.Error("expected empty string for no examples")
	}
}

func TestHugoLiveCLIExample(t *testing.T) {
	f := HugoFormatter{}
	got := f.LiveCLIExample("kosli list environments", "kosli list environments --output=json", "http://example.com")
	if !strings.Contains(got, "{{< raw-html >}}") {
		t.Error("expected raw-html shortcode")
	}
	if !strings.Contains(got, "kosli list environments --output=json") {
		t.Error("expected CLI command in output")
	}
}

func TestHugoExampleUseCases(t *testing.T) {
	f := HugoFormatter{}
	example := "# report a snyk attestation\nkosli attest snyk foo"
	got := f.ExampleUseCases("kosli attest snyk", example)
	if !strings.Contains(got, "## Examples Use Cases") {
		t.Error("expected heading")
	}
	if !strings.Contains(got, "##### report a snyk attestation") {
		t.Error("expected hash-titled example")
	}
	if !strings.Contains(got, "https://docs.kosli.com/getting_started") {
		t.Error("expected full URL for Hugo format")
	}
}

func TestHugoLinkHandler(t *testing.T) {
	f := HugoFormatter{}
	got := f.LinkHandler("kosli_attest_snyk.md")
	if got != "/client_reference/kosli_attest_snyk/" {
		t.Errorf("expected trailing slash link, got: %s", got)
	}
}
