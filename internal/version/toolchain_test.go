package version

import (
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// Dependabot bumps the Dockerfile's golang image on its own, but Snyk reads
// the Go version from go.mod, so a bump that skips go.mod hides a stdlib CVE.
func TestDockerfileBuildsWithTheGoInGoMod(t *testing.T) {
	goMod, err := os.ReadFile("../../go.mod")
	require.NoError(t, err)
	dockerfile, err := os.ReadFile("../../Dockerfile")
	require.NoError(t, err)

	want := regexp.MustCompile(`(?m)^go (\S+)$`).FindSubmatch(goMod)
	require.NotNil(t, want, "no go directive in go.mod")
	got := regexp.MustCompile(`(?m)^FROM golang:([0-9.]+)-`).FindSubmatch(dockerfile)
	require.NotNil(t, got, "no golang builder image in Dockerfile")

	require.Equal(t, string(want[1]), string(got[1]),
		"Dockerfile builds with a different Go than go.mod declares; bump both together")
}
