package intoto

import (
	"strings"
	"testing"

	"github.com/kosli-dev/cli/internal/sbom"
	"github.com/stretchr/testify/require"
)

const fixtureCommit = "4f9d2a7c1e6b8a3d5f0e2c9b7a1d4e6f8c3b5a70"

// statementWith reads a statement fixture with each old string replaced by
// its new one, for variations on a fixture that need no file of their own.
func statementWith(t *testing.T, name string, replacements ...string) *Attestation {
	t.Helper()
	content := string(statementFixture(t, name))
	for i := 0; i < len(replacements); i += 2 {
		require.Contains(t, content, replacements[i])
		content = strings.ReplaceAll(content, replacements[i], replacements[i+1])
	}
	attestations, err := Read([]byte(content))
	require.NoError(t, err)
	return attestations[0]
}

func TestProvenanceFromGitHubActions(t *testing.T) {
	p, err := ProvenanceFrom(statementWith(t, "github-provenance.json"), artifactDigest, "")
	require.NoError(t, err)
	require.Equal(t, &Provenance{
		Filename:  "artifact.bin",
		BuilderID: "https://github.com/actions/runner/github-hosted",
		BuildType: "https://actions.github.io/buildtypes/workflow/v1",
		BuildURL:  "https://github.com/acme/widget/actions/runs/1234567890/attempts/1",
		Commit:    fixtureCommit,
		Branch:    "main",
		Repository: Repository{
			URL:      "https://github.com/acme/widget",
			Name:     "acme/widget",
			ID:       "123456789",
			Host:     "github.com",
			Provider: "github",
		},
		CommitURL: "https://github.com/acme/widget/commit/" + fixtureCommit,
		Workflow:  ".github/workflows/build.yml",
	}, p)
}

func TestProvenanceFromGitHubActionsWithoutARunURL(t *testing.T) {
	attestation := statementWith(t, "github-provenance.json",
		`"invocationId": "https://github.com/acme/widget/actions/runs/1234567890/attempts/1"`, `"invocationId": "local-17"`)

	p, err := ProvenanceFrom(attestation, artifactDigest, "")
	require.NoError(t, err)
	require.Empty(t, p.BuildURL, "an invocation ID that is not a URL is not a build URL")
	require.Equal(t, []string{"build URL"}, p.Missing)
}

func TestProvenanceFromBuildKit(t *testing.T) {
	t.Run("reads the commit and repository from BuildKit's VCS metadata, and the run from the builder ID", func(t *testing.T) {
		p, err := ProvenanceFrom(statementWith(t, "buildkit-provenance.json"), artifactDigest, "")
		require.NoError(t, err)
		require.Equal(t, &Provenance{
			Filename:  "pkg:docker/acme/widget@1.2.3?platform=linux%2Famd64",
			BuilderID: "https://github.com/acme/widget/actions/runs/1234567890/attempts/1",
			BuildType: "https://github.com/moby/buildkit/blob/master/docs/attestations/slsa-definitions.md",
			BuildURL:  "https://github.com/acme/widget/actions/runs/1234567890/attempts/1",
			Commit:    fixtureCommit,
			Branch:    "main",
			Repository: Repository{
				URL:      "https://github.com/acme/widget",
				Name:     "acme/widget",
				Host:     "github.com",
				Provider: "github",
			},
			CommitURL: "https://github.com/acme/widget/commit/" + fixtureCommit,
		}, p)
	})

	t.Run("has no build URL when built outside CI, since its invocation ID is not one", func(t *testing.T) {
		p, err := ProvenanceFrom(statementWith(t, "buildkit-provenance.json",
			`"id": "https://github.com/acme/widget/actions/runs/1234567890/attempts/1"`, `"id": ""`), artifactDigest, "")
		require.NoError(t, err)
		require.Empty(t, p.BuildURL)
		require.Equal(t, []string{"build URL"}, p.Missing)
	})

	t.Run("falls back to the build's config source when there is no VCS metadata", func(t *testing.T) {
		p, err := ProvenanceFrom(statementWith(t, "buildkit-provenance.json",
			`"source": "https://github.com/acme/widget",`, ``,
			`"revision": "`+fixtureCommit+`"`, `"revision": ""`), artifactDigest, "")
		require.NoError(t, err)
		require.Equal(t, fixtureCommit, p.Commit)
		require.Equal(t, "https://github.com/acme/widget", p.Repository.URL)
		require.Equal(t, "main", p.Branch)
	})
}

func TestProvenanceFromSLSAv02(t *testing.T) {
	attestations, err := Read(statementFixture(t, "slsa-v0.2-provenance.json"))
	require.NoError(t, err)

	p, err := ProvenanceFrom(attestations[0], artifactDigest, "")
	require.NoError(t, err)
	require.Equal(t, &Provenance{
		Filename:  "artifact.bin",
		BuilderID: "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@refs/tags/v2.0.0",
		BuildType: "https://github.com/slsa-framework/slsa-github-generator/generic@v1",
		BuildURL:  "https://github.com/acme/widget/actions/runs/1234567890/attempts/1",
		Commit:    fixtureCommit,
		Branch:    "main",
		Repository: Repository{
			URL:      "https://github.com/acme/widget",
			Name:     "acme/widget",
			ID:       "123456789",
			Host:     "github.com",
			Provider: "github",
		},
		CommitURL: "https://github.com/acme/widget/commit/" + fixtureCommit,
		Workflow:  ".github/workflows/release.yml",
	}, p)
}

func TestProvenanceFromOtherSources(t *testing.T) {
	t.Run("uses the provider given for a host it does not recognise", func(t *testing.T) {
		attestation := statementWith(t, "github-provenance.json", "https://github.com/acme/widget", "https://github.example.com/acme/widget")

		p, err := ProvenanceFrom(attestation, artifactDigest, "github")
		require.NoError(t, err)
		require.Equal(t, "github.example.com", p.Repository.Host)
		require.Equal(t, "github", p.Repository.Provider)
		require.Equal(t, "https://github.example.com/acme/widget/commit/"+fixtureCommit, p.CommitURL)
	})

	t.Run("cannot build a commit URL for an unknown host with no provider", func(t *testing.T) {
		attestation := statementWith(t, "github-provenance.json", "https://github.com/acme/widget", "https://git.example.com/acme/widget")

		p, err := ProvenanceFrom(attestation, artifactDigest, "")
		require.NoError(t, err)
		require.Empty(t, p.Repository.Provider)
		require.Empty(t, p.CommitURL)
		require.Equal(t, []string{"commit URL"}, p.Missing)
	})

	t.Run("derives GitLab commit URLs", func(t *testing.T) {
		attestation := statementWith(t, "github-provenance.json", "https://github.com/acme/widget", "https://gitlab.com/acme/widget")

		p, err := ProvenanceFrom(attestation, artifactDigest, "")
		require.NoError(t, err)
		require.Equal(t, "gitlab", p.Repository.Provider)
		require.Equal(t, "https://gitlab.com/acme/widget/-/commit/"+fixtureCommit, p.CommitURL)
	})

	t.Run("records what an unknown build type offers and says what is missing", func(t *testing.T) {
		attestation := statementWith(t, "github-provenance.json",
			"https://actions.github.io/buildtypes/workflow/v1", "https://example.com/build/v1",
			`"invocationId": "https://github.com/acme/widget/actions/runs/1234567890/attempts/1"`, `"invocationId": "local-build-17"`)

		p, err := ProvenanceFrom(attestation, artifactDigest, "")
		require.NoError(t, err)
		require.Equal(t, "https://example.com/build/v1", p.BuildType)
		require.Equal(t, fixtureCommit, p.Commit)
		require.Equal(t, "https://github.com/acme/widget", p.Repository.URL)
		require.Equal(t, "main", p.Branch)
		require.Empty(t, p.Workflow)
		require.Equal(t, []string{"build URL"}, p.Missing)
	})

	t.Run("says everything is missing from an empty predicate", func(t *testing.T) {
		attestations, err := Read([]byte(`{"_type": "https://in-toto.io/Statement/v1", "subject": [{"name": "a", "digest": {"sha256": "` + artifactDigest + `"}}], "predicateType": "https://slsa.dev/provenance/v1", "predicate": {}}`))
		require.NoError(t, err)

		p, err := ProvenanceFrom(attestations[0], artifactDigest, "")
		require.NoError(t, err)
		require.Equal(t, []string{"build URL", "commit", "commit URL", "repository"}, p.Missing)
	})

	t.Run("refuses an attestation that is not provenance", func(t *testing.T) {
		_, err := ProvenanceFrom(statementWith(t, "spdx-sbom.json"), artifactDigest, "")
		require.EqualError(t, err, "the attestation is https://spdx.dev/Document/v2.3, not SLSA provenance")
	})
}

func TestSBOMDocument(t *testing.T) {
	t.Run("hands internal/sbom the SPDX document", func(t *testing.T) {
		document, err := SBOMDocument(statementWith(t, "spdx-sbom.json"))
		require.NoError(t, err)
		data, err := sbom.ProcessSBOM(document)
		require.NoError(t, err)
		require.Equal(t, "spdx-2.3", data.Format)
		require.Equal(t, 1, data.Document.PackageCount)
	})

	t.Run("hands internal/sbom the CycloneDX document", func(t *testing.T) {
		document, err := SBOMDocument(statementWith(t, "cyclonedx-sbom.json"))
		require.NoError(t, err)
		data, err := sbom.ProcessSBOM(document)
		require.NoError(t, err)
		require.Equal(t, "cyclonedx-1.6", data.Format)
	})

	t.Run("refuses an attestation that is not an SBOM", func(t *testing.T) {
		_, err := SBOMDocument(statementWith(t, "github-provenance.json"))
		require.EqualError(t, err, "the attestation is https://slsa.dev/provenance/v1, not an SPDX or CycloneDX SBOM")
	})
}
