package docgen

import "testing"

// Compile-time interface satisfaction checks.
var _ Formatter = (*HugoFormatter)(nil)
var _ Formatter = (*MintlifyFormatter)(nil)
var _ LiveDocProvider = (*NullLiveDocProvider)(nil)

func TestNullLiveDocProvider(t *testing.T) {
	p := NullLiveDocProvider{}
	if p.YamlDocExists("GitHub", "cmd") {
		t.Error("expected false")
	}
	if p.EventDocExists("GitHub", "cmd") {
		t.Error("expected false")
	}
	if p.YamlURL("GitHub", "cmd") != "" {
		t.Error("expected empty string")
	}
	if p.EventURL("GitHub", "cmd") != "" {
		t.Error("expected empty string")
	}
	fc, u, exists := p.CLIDocExists("cmd")
	if fc != "" || u != "" || exists {
		t.Error("expected empty/false")
	}
}
