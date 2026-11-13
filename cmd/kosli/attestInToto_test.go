package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/kosli-dev/cli/internal/gitview"
	"github.com/kosli-dev/cli/internal/intoto"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	intotoFixtures = "../../internal/intoto/testdata"
	intotoArtifact = intotoFixtures + "/artifact.bin"
	intotoDigest   = "e88951291bca0807a41f3c0144009bfff5c6c77dc1310d52e8c6d62feee960a8"
	intotoCommit   = "4f9d2a7c1e6b8a3d5f0e2c9b7a1d4e6f8c3b5a70"
	intotoSigner   = "--certificate-identity https://github.com/acme/widget/.github/workflows/build.yml@refs/heads/main --certificate-oidc-issuer https://token.actions.githubusercontent.com"
	intotoBuildURL = "https://github.com/acme/widget/actions/runs/1234567890/attempts/1"
	intotoRerunURL = "https://github.com/acme/widget/actions/runs/1234567890/attempts/2"
	// intotoDeployURL is another run in the repository, such as a deployment
	// of what the build at intotoBuildURL made.
	intotoDeployURL      = "https://github.com/acme/widget/actions/runs/77/attempts/1"
	intotoBranchURL      = "https://github.com/acme/widget/actions/runs/2222222222/attempts/1"
	intotoBranchSigner   = "https://github.com/acme/widget/.github/workflows/build.yml@refs/heads/feature-x"
	intotoPullURL        = "https://github.com/acme/widget/actions/runs/3333333333/attempts/1"
	intotoPullSigner     = "https://github.com/acme/widget/.github/workflows/build.yml@refs/pull/12/merge"
	intotoMainSigner     = "https://github.com/acme/widget/.github/workflows/build.yml@refs/heads/main"
	intotoPinTheSigner   = "No signer is pinned, and the default only accepts an attestation made by the run this command is part of (%s). To record one made by another run, pin its signer: pass --certificate-identity or --certificate-identity-regexp with --certificate-oidc-issuer, or --key"
	intotoListingPath    = "/api/v3/repos/acme/widget/attestations/sha256:" + intotoDigest
	intotoCommitURL      = "https://github.com/acme/widget/commit/" + intotoCommit
	intotoLogEntryURL    = "https://rekor.localhost/api/v1/log/entries?logIndex=0"
	intotoSBOMTemplate   = "version: 1\ntrail:\n  artifacts:\n  - name: app\n    attestations:\n    - name: sbom\n      type: sbom\n"
	intotoDefaultCommand = "attest in-toto " + intotoArtifact + " --artifact-type file"
	// intotoGoneBundle, among the fixtures GitHub holds, is an attestation
	// listed at a download URL that no longer answers.
	intotoGoneBundle = "gone"
)

type AttestInTotoCommandTestSuite struct {
	suite.Suite
	flowName              string
	trailName             string
	defaultKosliArguments string
	github                *httptest.Server
	githubArguments       string
	savedTrustRoots       func() *intoto.TrustRoots
	savedGitHubRun        func() (string, bool)
	savedGitHubAPI        string

	// attestations names the fixtures GitHub's API lists for the artifact, in
	// order. The handler reads it on the server's goroutine.
	attestationsMu sync.Mutex
	attestations   []string
}

func (suite *AttestInTotoCommandTestSuite) SetupSuite() {
	// GitHub's API for acme/widget: the attestations a test stores for the
	// artifact, and the one commit the provenance fixtures name.
	suite.github = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == intotoListingPath {
			suite.listAttestations(w)
			return
		}
		if r.URL.Path != "/api/v3/repos/acme/widget/commits/"+intotoCommit {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, `{"message": "Not Found"}`)
			return
		}
		_, _ = fmt.Fprintf(w, `{
			"sha": %q,
			"html_url": %q,
			"author": {"login": "jdoe"},
			"commit": {
				"message": "Add the widget",
				"author": {"name": "Jane Doe", "email": "jane@example.com", "date": "2026-11-01T09:00:00Z"}
			}
		}`, intotoCommit, intotoCommitURL)
	}))
	suite.githubArguments = fmt.Sprintf(" --github-token a-token --github-base-url %s/", suite.github.URL)
}

func (suite *AttestInTotoCommandTestSuite) TearDownSuite() {
	suite.github.Close()
}

func (suite *AttestInTotoCommandTestSuite) listAttestations(w http.ResponseWriter) {
	suite.attestationsMu.Lock()
	defer suite.attestationsMu.Unlock()

	type listed struct {
		Bundle    json.RawMessage `json:"bundle,omitempty"`
		BundleURL string          `json:"bundle_url,omitempty"`
	}
	attestations := []listed{}
	for _, name := range suite.attestations {
		if name == intotoGoneBundle {
			attestations = append(attestations, listed{BundleURL: suite.github.URL + "/blobs/gone?sig=a-secret"})
			continue
		}
		content, err := os.ReadFile(filepath.Join(intotoFixtures, name))
		require.NoError(suite.T(), err)
		var bundle bytes.Buffer
		require.NoError(suite.T(), json.Compact(&bundle, content))
		attestations = append(attestations, listed{Bundle: bundle.Bytes()})
	}
	require.NoError(suite.T(), json.NewEncoder(w).Encode(map[string]any{"attestations": attestations}))
}

// githubHolds sets the attestations GitHub lists for the artifact.
func (suite *AttestInTotoCommandTestSuite) githubHolds(fixtures ...string) {
	suite.attestationsMu.Lock()
	defer suite.attestationsMu.Unlock()
	suite.attestations = fixtures
}

// inGitHubActions has the command believe it runs in the GitHub Actions run
// at runURI.
func (suite *AttestInTotoCommandTestSuite) inGitHubActions(runURI string) {
	currentGitHubRun = func() (string, bool) { return runURI, true }
}

// fromGitHub is a command reading acme/widget's attestations from github.com,
// which the fake GitHub stands in for.
func (suite *AttestInTotoCommandTestSuite) fromGitHub(flags string) string {
	return suite.command("--github-org acme --repository widget " + flags + " --github-token a-token")
}

// fromGitHubEnterprise is the same command reading them from the fake GitHub
// as a server named with --github-base-url.
func (suite *AttestInTotoCommandTestSuite) fromGitHubEnterprise(flags string) string {
	return suite.command("--github-org acme --repository widget " + flags + suite.githubArguments)
}

func (suite *AttestInTotoCommandTestSuite) SetupTest() {
	suite.flowName = "attest-in-toto"
	suite.trailName = "test-123"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --flow %s --trail %s --host %s --org %s --api-token %s",
		suite.flowName, suite.trailName, global.Host, global.Org, global.ApiToken)

	trustedRoot, err := root.NewTrustedRootFromPath(filepath.Join(intotoFixtures, "trusted_root.json"))
	require.NoError(suite.T(), err)
	suite.savedTrustRoots = intotoTrustRoots
	intotoTrustRoots = func() *intoto.TrustRoots { return intoto.FixedTrustRoots(trustedRoot) }
	// The tests run in GitHub Actions too, and must not take its defaults.
	suite.savedGitHubRun = currentGitHubRun
	currentGitHubRun = func() (string, bool) { return "", false }
	suite.savedGitHubAPI = intotoGitHubAPI
	intotoGitHubAPI = suite.github.URL + "/"
	suite.githubHolds()

	template := filepath.Join(suite.T().TempDir(), "template.yml")
	require.NoError(suite.T(), os.WriteFile(template, []byte(intotoSBOMTemplate), 0o600))
	CreateFlowWithTemplate(suite.flowName, template, suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.T())
}

func (suite *AttestInTotoCommandTestSuite) TearDownTest() {
	intotoTrustRoots = suite.savedTrustRoots
	currentGitHubRun = suite.savedGitHubRun
	intotoGitHubAPI = suite.savedGitHubAPI
}

func (suite *AttestInTotoCommandTestSuite) command(flags string) string {
	return intotoDefaultCommand + " " + flags + suite.defaultKosliArguments
}

func (suite *AttestInTotoCommandTestSuite) bundle(name string) string {
	return " --bundle " + filepath.Join(intotoFixtures, name)
}

func fixtureSHA256(t *testing.T, name string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(intotoFixtures, name))
	require.NoError(t, err)
	return fmt.Sprintf("%x", sha256.Sum256(content))
}

// compactFixtureSHA256 is the SHA-256 of a fixture as GitHub's API returns
// it: on one line, inside the listing.
func compactFixtureSHA256(t *testing.T, name string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(intotoFixtures, name))
	require.NoError(t, err)
	var compacted bytes.Buffer
	require.NoError(t, json.Compact(&compacted, content))
	return fmt.Sprintf("%x", sha256.Sum256(compacted.Bytes()))
}

// statementVariant writes a provenance statement fixture with each old string
// replaced by its new one, and returns the file's path.
func statementVariant(t *testing.T, replacements ...string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(intotoFixtures, "statements", "github-provenance.json"))
	require.NoError(t, err)
	statement := string(content)
	for i := 0; i < len(replacements); i += 2 {
		require.Contains(t, statement, replacements[i])
		statement = strings.ReplaceAll(statement, replacements[i], replacements[i+1])
	}
	path := filepath.Join(t.TempDir(), "provenance.json")
	require.NoError(t, os.WriteFile(path, []byte(statement), 0o600))
	return path
}

func (suite *AttestInTotoCommandTestSuite) TestProvenanceBuildsTheArtifactFromTheStatement() {
	bundleSHA256 := fixtureSHA256(suite.T(), "github-provenance.sigstore.json")
	signedCommand := suite.command("--predicate-type provenance --name app --dry-run " + intotoSigner +
		suite.bundle("github-provenance.sigstore.json") + suite.githubArguments)

	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:        "posts the artifact with the subject's fingerprint and filename",
			cmd:         signedCommand,
			goldenRegex: `(?s)artifacts/docs-cmd-test-user/attest-in-toto.*"fingerprint": "` + intotoDigest + `",\s+"filename": "artifact.bin",\s+"git_commit": "` + intotoCommit + `"`,
		},
		{
			name:        "records the build's run and commit URL from the statement",
			cmd:         signedCommand,
			goldenRegex: `(?s)"build_url": "` + intotoBuildURL + `",\s+"commit_url": "` + intotoCommitURL + `",\s+"repo_url": "https://github.com/acme/widget",\s+"template_reference_name": "app"`,
		},
		{
			name:        "records the repository the statement names",
			cmd:         signedCommand,
			goldenRegex: `(?s)"repo_info": \{\s+"url": "https://github.com/acme/widget",\s+"name": "acme/widget",\s+"id": "123456789",\s+"provider": "github"\s+\}`,
		},
		{
			name:        "looks the commit's details up on GitHub, keeping the statement's branch",
			cmd:         signedCommand,
			goldenRegex: `(?s)"git_commit_info": \{\s+"sha1": "` + intotoCommit + `",\s+"message": "Add the widget",\s+"author": "Jane Doe \\u003cjane@example.com\\u003e",\s+"timestamp": 1793523600,\s+"branch": "main",\s+"url": "` + intotoCommitURL + `"\s+\}`,
		},
		{
			name:        "links the transparency log entry, pinned by the bundle's SHA-256",
			cmd:         signedCommand,
			goldenRegex: `(?s)"external_urls": \{\s+"attestation": \{\s+"href": "` + regexpQuote(intotoLogEntryURL) + `",\s+"fingerprint": "` + bundleSHA256 + `"`,
		},
		{
			name:        "annotates the builder, the build type, the workflow and the verification",
			cmd:         signedCommand,
			goldenRegex: `(?s)"build_type": "https://actions.github.io/buildtypes/workflow/v1",\s+"builder_id": "https://github.com/actions/runner/github-hosted",\s+"integrated_time": "[^"]+",\s+"log_index": "0",\s+"signed_timestamp": "[^"]+",\s+"signer_identity": "https://github.com/acme/widget/.github/workflows/build.yml@refs/heads/main",\s+"signer_issuer": "https://token.actions.githubusercontent.com",\s+"verified": "true",\s+"workflow": ".github/workflows/build.yml"`,
		},
		{
			name:        "keeps the caller's own annotations and external URLs",
			cmd:         signedCommand + " --annotate team=platform --external-url docs=https://example.com/docs",
			goldenRegex: `(?s)"docs": \{\s+"href": "https://example.com/docs"\s+\}.*"team": "platform"`,
		},
		{
			name:        "reads BuildKit provenance",
			cmd:         suite.command("--predicate-type provenance --name app --dry-run " + intotoSigner + suite.bundle("buildkit-provenance.sigstore.json") + suite.githubArguments),
			goldenRegex: `(?s)"filename": "pkg:docker/acme/widget@1.2.3\?platform=linux%2Famd64".*"build_url": "` + intotoBuildURL + `".*"build_type": "https://github.com/moby/buildkit/blob/master/docs/attestations/slsa-definitions.md"`,
		},
		{
			name:        "reads SLSA v0.2 provenance",
			cmd:         suite.command("--predicate-type provenance --name app --dry-run " + intotoSigner + suite.bundle("slsa-v0.2-provenance.sigstore.json") + suite.githubArguments),
			goldenRegex: `(?s)"build_url": "` + intotoBuildURL + `".*"workflow": ".github/workflows/release.yml"`,
		},
		{
			name:        "finds the provenance among other attestations in a .intoto.jsonl file",
			cmd:         suite.command("--predicate-type provenance --name app --dry-run " + intotoSigner + suite.bundle("attestations.intoto.jsonl") + suite.githubArguments),
			goldenRegex: `(?s)"git_commit": "` + intotoCommit + `".*"verified": "true"`,
		},
		{
			name:        "verifies a DSSE envelope with the key that signed it",
			cmd:         suite.command("--predicate-type provenance --name app --dry-run --key " + filepath.Join(intotoFixtures, "key.pub") + suite.bundle("key-signed-provenance.dsse.json") + suite.githubArguments),
			goldenRegex: `(?s)"annotations": \{\s+"build_type": "[^"]+",\s+"builder_id": "[^"]+",\s+"verified": "true",`,
		},
		{
			name:        "records an unsigned statement as unverified when allowed",
			cmd:         suite.command("--predicate-type provenance --name app --dry-run --allow-unsigned --bundle " + filepath.Join(intotoFixtures, "statements", "github-provenance.json") + suite.githubArguments),
			goldenRegex: `(?s)"verified": "false"`,
		},
	})
}

func (suite *AttestInTotoCommandTestSuite) TestProvenanceWhenTheStatementLacksDetails() {
	// CI sets a real token, which would send the cases that expect no lookup to api.github.com.
	suite.T().Setenv("KOSLI_GITHUB_TOKEN", "")
	noURLs := statementVariant(suite.T(),
		`"invocationId": "`+intotoBuildURL+`"`, `"invocationId": "local-17"`,
		"https://github.com/acme/widget", "https://git.example.com/acme/widget")

	runTestCmd(suite.T(), []cmdTestCase{
		{
			wantError: true,
			name:      "fails naming the build URL and commit URL it needs, and the flags that supply them",
			cmd:       suite.command("--predicate-type provenance --name app --allow-unsigned --bundle " + noURLs),
			golden:    "Error: the provenance does not say the build URL or the commit URL, which Kosli records an artifact with. Pass --build-url and --commit-url\n",
		},
		{
			name:        "takes the build URL and commit URL from the flags when the statement lacks them",
			cmd:         suite.command("--predicate-type provenance --name app --dry-run --allow-unsigned --bundle " + noURLs + " --build-url https://ci.example.com/17 --commit-url https://git.example.com/acme/widget/commit/" + intotoCommit),
			goldenRegex: `(?s)"build_url": "https://ci.example.com/17",\s+"commit_url": "https://git.example.com/acme/widget/commit/` + intotoCommit + `"`,
		},
		{
			name:        "prefers the statement's build URL to the flag's",
			cmd:         suite.command("--predicate-type provenance --name app --dry-run " + intotoSigner + suite.bundle("github-provenance.sigstore.json") + " --build-url https://ci.example.com/this-job" + suite.githubArguments),
			goldenRegex: `"build_url": "` + intotoBuildURL + `"`,
		},
		{
			// POST /artifacts refuses commit info without every field, so none is sent.
			name:        "records only the commit's SHA when its details cannot be found",
			cmd:         suite.command("--predicate-type provenance --name app --dry-run " + intotoSigner + suite.bundle("github-provenance.sigstore.json")),
			goldenRegex: `(?s)^\[warning\] the details and branch of commit ` + intotoCommit + ` are not recorded, only its SHA.*"git_commit": "` + intotoCommit + `",\s+"git_commit_info": null`,
		},
		{
			name:   "has the server accept an artifact known only by its commit's SHA",
			cmd:    suite.command("--predicate-type provenance --name app " + intotoSigner + suite.bundle("github-provenance.sigstore.json")),
			golden: "[warning] the details and branch of commit " + intotoCommit + " are not recorded, only its SHA: they could not be looked up, and the git repository at --repo-root . does not hold the commit\nartifact artifact.bin was attested with fingerprint: " + intotoDigest + "\n",
		},
		{
			name:        "warns when GitHub cannot find the commit, then records only its SHA",
			cmd:         suite.command("--predicate-type provenance --name app --dry-run --allow-unsigned --bundle " + statementVariant(suite.T(), "acme/widget", "acme/gadget") + suite.githubArguments),
			goldenRegex: `(?s)^\[warning\] could not look up commit ` + intotoCommit + `: failed to get commit ` + intotoCommit + ` from GitHub repository acme/gadget.*"git_commit_info": null`,
		},
		{
			name:        "reads the commit's details from a local checkout that holds it",
			cmd:         suite.command("--predicate-type provenance --name app --dry-run --allow-unsigned --repo-root ../.. --bundle " + statementVariant(suite.T(), intotoCommit, localHeadCommit(suite.T()))),
			goldenRegex: `(?s)"git_commit_info": \{\s+"sha1": "` + localHeadCommit(suite.T()) + `",\s+"message": ` + jsonString + `,\s+"author": ` + jsonString + `,\s+"timestamp": \d+,\s+"branch": "main"`,
		},
	})
}

// jsonString matches any JSON-encoded string, including escaped quotes, because
// the local checkout's commit message and author are not under the test's control.
const jsonString = `"(?:[^"\\]|\\.)+"`

func (suite *AttestInTotoCommandTestSuite) TestSBOMIsRecordedLikeAttestSbom() {
	bundleSHA256 := fixtureSHA256(suite.T(), "spdx-sbom.sigstore.json")
	signedCommand := suite.command("--predicate-type sbom --name app.sbom --dry-run " + intotoSigner + suite.bundle("spdx-sbom.sigstore.json"))

	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:        "posts an sbom attestation to the system endpoint, named for the artifact in the template",
			cmd:         signedCommand,
			goldenRegex: `(?s)trail/test-123/system.*"artifact_fingerprint": "` + intotoDigest + `",\s+"attestation_name": "sbom",\s+"target_artifacts": \[\s+"app"\s+\].*"type_name": "sbom"`,
		},
		{
			name:        "summarises the SBOM and fingerprints the attached bundle",
			cmd:         signedCommand,
			goldenRegex: `(?s)"attestation_data": \{\s+"format": "spdx-2\.3",\s+"original_fingerprint": "` + bundleSHA256 + `",\s+"document": \{.*"package_count": 1`,
		},
		{
			name:        "annotates the SBOM and its verification",
			cmd:         signedCommand,
			goldenRegex: `(?s)"sbom_format": "spdx-2\.3",\s+"sbom_sha256": "` + bundleSHA256 + `",.*"verified": "true"`,
		},
		{
			name:        "uploads the bundle as the attachment",
			cmd:         signedCommand,
			goldenRegex: `(?s)"filename": "spdx-sbom.sigstore.json".*"sha256": "` + bundleSHA256 + `"`,
		},
		{
			name:        "reads CycloneDX",
			cmd:         suite.command("--predicate-type sbom --name container-sbom --dry-run " + intotoSigner + suite.bundle("cyclonedx-sbom.sigstore.json")),
			goldenRegex: `(?s)"attestation_name": "container-sbom".*"format": "cyclonedx-1\.6"`,
		},
		{
			name:        "attaches only the matching line of a .intoto.jsonl file",
			cmd:         suite.command("--predicate-type sbom --name app.sbom --dry-run " + intotoSigner + suite.bundle("attestations.intoto.jsonl")),
			goldenRegex: `"filename": "attestations.intoto.3.json"`,
		},
	})
}

func (suite *AttestInTotoCommandTestSuite) TestRecordsSatisfyTheTemplate() {
	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:   "records the artifact named app",
			cmd:    suite.command("--predicate-type provenance --name app " + intotoSigner + suite.bundle("github-provenance.sigstore.json") + suite.githubArguments),
			golden: fmt.Sprintf("artifact artifact.bin was attested with fingerprint: %s\n", intotoDigest),
		},
		{
			name:   "records the SBOM named app.sbom",
			cmd:    suite.command("--predicate-type sbom --name app.sbom " + intotoSigner + suite.bundle("spdx-sbom.sigstore.json")),
			golden: "sbom attestation 'sbom' is reported to trail: test-123\n",
		},
		{
			name: "leaves the trail with both of the template's records",
			cmd:  fmt.Sprintf("get trail %s --flow %s --output json --host %s --org %s --api-token %s", suite.trailName, suite.flowName, global.Host, global.Org, global.ApiToken),
			goldenJson: []jsonCheck{
				{"compliance_status.status", "COMPLIANT"},
				{"compliance_status.artifacts_statuses.app.artifact_fingerprint", intotoDigest},
				{"compliance_status.artifacts_statuses.app.attestations_statuses.[0].attestation_name", "sbom"},
				{"compliance_status.artifacts_statuses.app.attestations_statuses.[0].unexpected", false},
			},
		},
	})
}

func (suite *AttestInTotoCommandTestSuite) TestNameFromTheEnvironment() {
	suite.T().Setenv("KOSLI_NAME", "app")
	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:        "takes --name from KOSLI_NAME, as attest artifact and attest sbom do",
			cmd:         suite.command("--predicate-type provenance --dry-run " + intotoSigner + suite.bundle("github-provenance.sigstore.json") + suite.githubArguments),
			goldenRegex: `"template_reference_name": "app"`,
		},
	})
}

func (suite *AttestInTotoCommandTestSuite) TestRefusesWhatItCannotTrust() {
	runTestCmd(suite.T(), []cmdTestCase{
		{
			wantError: true,
			name:      "a statement about another artifact",
			cmd:       suite.command("--predicate-type provenance --name app " + intotoSigner + suite.bundle("wrong-subject-provenance.sigstore.json")),
			golden:    "Error: bundle file " + intotoFixtures + "/wrong-subject-provenance.sigstore.json: no provenance attestation lists the artifact (sha256:" + intotoDigest + ") as its subject; the subjects found were: other.bin (sha256:11de96226896b0fcf4ea66a11a63a5fd9f4e211fe94b62b7eebf4485c817b2d8)\n",
		},
		{
			wantError:   true,
			name:        "a bundle signed by another identity",
			cmd:         suite.command("--predicate-type provenance --name app " + intotoSigner + suite.bundle("wrong-signer-provenance.sigstore.json")),
			goldenRegex: `^Error: bundle file .*/wrong-signer-provenance.sigstore.json: the attestation failed verification: .*https://github.com/mallory/widget`,
		},
		{
			wantError:   true,
			name:        "a file where one attestation has the wrong signer, though another is unsigned and allowed",
			cmd:         suite.command("--predicate-type provenance --name app --allow-unsigned " + intotoSigner + suite.bundle("wrong-signer-and-unsigned.intoto.jsonl")),
			goldenRegex: `^Error: bundle file .*/wrong-signer-and-unsigned.intoto.jsonl: attestation 1 failed verification: .*https://github.com/mallory/widget`,
		},
		{
			wantError: true,
			name:      "a key-signed envelope checked against an identity, which cannot be verified rather than failing",
			cmd:       suite.command("--predicate-type provenance --name app " + intotoSigner + suite.bundle("key-signed-provenance.dsse.json")),
			golden:    "Error: bundle file " + intotoFixtures + "/key-signed-provenance.dsse.json: the attestation could not be verified: a DSSE envelope carries no certificate to check --certificate-identity against. Supply the Sigstore bundle instead, or verify the envelope with --key\n",
		},
		{
			wantError:   true,
			name:        "a statement changed after signing",
			cmd:         suite.command("--predicate-type provenance --name app " + intotoSigner + suite.bundle("tampered-provenance.sigstore.json")),
			goldenRegex: `^Error: bundle file .*/tampered-provenance.sigstore.json: the attestation failed verification: `,
		},
		{
			wantError: true,
			name:      "a signed bundle with no signer pinned, naming the flags that pin one",
			cmd:       suite.command("--predicate-type provenance --name app" + suite.bundle("github-provenance.sigstore.json")),
			golden:    "Error: bundle file " + intotoFixtures + "/github-provenance.sigstore.json: the attestation could not be verified: it is signed, so its signer must be pinned to verify it. Pass --certificate-identity or --certificate-identity-regexp with --certificate-oidc-issuer, or --key\n",
		},
		{
			wantError: true,
			name:      "a file where one attestation is signed and no signer is pinned, though another is unsigned and allowed",
			cmd:       suite.command("--predicate-type provenance --name app --allow-unsigned" + suite.bundle("wrong-signer-and-unsigned.intoto.jsonl")),
			golden:    "Error: bundle file " + intotoFixtures + "/wrong-signer-and-unsigned.intoto.jsonl: attestation 1 could not be verified: it is signed, so its signer must be pinned to verify it. Pass --certificate-identity or --certificate-identity-regexp with --certificate-oidc-issuer, or --key\n",
		},
		{
			wantError: true,
			name:      "an unsigned statement without --allow-unsigned",
			cmd:       suite.command("--predicate-type provenance --name app --bundle " + filepath.Join(intotoFixtures, "statements", "buildkit-provenance.json")),
			golden:    "Error: bundle file " + intotoFixtures + "/statements/buildkit-provenance.json: the attestation is not signed. Pass --allow-unsigned to record it as unverified\n",
		},
		{
			wantError: true,
			name:      "a file holding only other kinds of attestation, naming them",
			cmd:       suite.command("--predicate-type provenance --name app " + intotoSigner + suite.bundle("vuln-only.intoto.jsonl")),
			golden:    "Error: bundle file " + intotoFixtures + "/vuln-only.intoto.jsonl: found no provenance attestation (SLSA provenance v1 or v0.2); the predicate types found were: https://cosign.sigstore.dev/attestation/vuln/v1\n",
		},
	})
}

func (suite *AttestInTotoCommandTestSuite) TestGitHubSourceRecordsWhatGitHubHolds() {
	suite.githubHolds("cyclonedx-sbom.sigstore.json", "github-provenance.sigstore.json")
	bundleSHA256 := compactFixtureSHA256(suite.T(), "github-provenance.sigstore.json")
	sbomSHA256 := compactFixtureSHA256(suite.T(), "cyclonedx-sbom.sigstore.json")

	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:        "records the provenance GitHub holds for the artifact, verified against the pinned signer",
			cmd:         suite.fromGitHub("--source github --predicate-type provenance --name app --dry-run " + intotoSigner),
			goldenRegex: `(?s)"fingerprint": "` + intotoDigest + `",\s+"filename": "artifact.bin",\s+"git_commit": "` + intotoCommit + `".*"build_url": "` + intotoBuildURL + `".*"signer_identity": "https://github.com/acme/widget/.github/workflows/build.yml@refs/heads/main",.*"verified": "true"`,
		},
		{
			name:        "links the transparency log entry, pinned by the SHA-256 of the bundle GitHub returned",
			cmd:         suite.fromGitHub("--source github --predicate-type provenance --name app --dry-run " + intotoSigner),
			goldenRegex: `(?s)"external_urls": \{\s+"attestation": \{\s+"href": "` + regexpQuote(intotoLogEntryURL) + `",\s+"fingerprint": "` + bundleSHA256 + `"`,
		},
		{
			name:        "records the SBOM GitHub holds, attaching its bundle",
			cmd:         suite.fromGitHub("--source github --predicate-type sbom --name app.sbom --dry-run " + intotoSigner),
			goldenRegex: `(?s)"filename": "attestation.sigstore.json".*"sha256": "` + sbomSHA256 + `".*"attestation_name": "sbom".*"verified": "true".*"format": "cyclonedx-1\.6",\s+"original_fingerprint": "` + sbomSHA256 + `"`,
		},
		{
			name:        "takes the repository in the owner/name form GitHub Actions gives it in",
			cmd:         suite.command("--source github --github-org acme --repository acme/widget --predicate-type provenance --name app --dry-run " + intotoSigner + suite.githubArguments),
			goldenRegex: `(?s)"git_commit": "` + intotoCommit + `".*"verified": "true"`,
		},
		{
			name:   "has the server accept the artifact",
			cmd:    suite.fromGitHub("--source github --predicate-type provenance --name app " + intotoSigner),
			golden: fmt.Sprintf("artifact artifact.bin was attested with fingerprint: %s\n", intotoDigest),
		},
	})
}

func (suite *AttestInTotoCommandTestSuite) TestGitHubSourceInGitHubActions() {
	suite.githubHolds("github-provenance.sigstore.json")
	suite.inGitHubActions("https://github.com/acme/deploy/actions/runs/77/attempts/1")

	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:        "is the default source",
			cmd:         suite.fromGitHub("--predicate-type provenance --name app --dry-run " + intotoSigner),
			goldenRegex: `(?s)"git_commit": "` + intotoCommit + `".*"verified": "true"`,
		},
		{
			name:        "records the build's commit and run, not this job's, when run from another workflow",
			cmd:         suite.fromGitHub("--predicate-type provenance --name app --dry-run " + intotoSigner),
			goldenRegex: `(?s)"git_commit": "` + intotoCommit + `".*"build_url": "` + intotoBuildURL + `",\s+"commit_url": "` + intotoCommitURL + `"`,
		},
		{
			name:        "still reads a file when --bundle is given",
			cmd:         suite.command("--predicate-type provenance --name app --dry-run " + intotoSigner + suite.bundle("rerun-provenance.sigstore.json") + suite.githubArguments),
			goldenRegex: `"build_url": "` + intotoRerunURL + `"`,
		},
	})
}

func (suite *AttestInTotoCommandTestSuite) TestGitHubSourceDefaultSigner() {
	// Newest last. The repository's own workflow file signed all three: on
	// main in the build, then on a branch and in a pull request, which anyone
	// who can push a branch can have attest the same artifact.
	suite.githubHolds("github-provenance.sigstore.json", "branch-provenance.sigstore.json", "pull-request-provenance.sigstore.json")
	defaultSigner := suite.fromGitHub("--predicate-type provenance --name app")
	otherRuns := func(runURI string, madeBy ...string) string {
		named := make([]string, len(madeBy)/2)
		for i := range named {
			described := "the attestation"
			if len(named) > 1 {
				described = fmt.Sprintf("attestation %d", i+1)
			}
			named[i] = fmt.Sprintf("%s was made by another run: %s, signed as %s", described, madeBy[2*i], madeBy[2*i+1])
		}
		refused := named[0]
		if len(named) > 1 {
			refused = fmt.Sprintf("%d attestations could not be accepted: %s", len(named), strings.Join(named, "; "))
		}
		return "Error: GitHub repository acme/widget: " + refused + ". " + fmt.Sprintf(intotoPinTheSigner, runURI) + "\n"
	}

	suite.inGitHubActions(intotoBuildURL)
	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:        "is a workflow in the repository, for what the run this command is part of attested",
			cmd:         defaultSigner + " --dry-run",
			goldenRegex: `(?s)^[^\[].*"build_url": "` + intotoBuildURL + `".*"signer_identity": "` + intotoMainSigner + `",\s+"signer_issuer": "https://token.actions.githubusercontent.com",\s+"verified": "true"`,
		},
	})

	suite.inGitHubActions(intotoRerunURL)
	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:        "accepts what an earlier attempt of the run attested",
			cmd:         defaultSigner + " --dry-run",
			goldenRegex: `(?s)"build_url": "` + intotoBuildURL + `".*"verified": "true"`,
		},
	})

	suite.inGitHubActions(intotoDeployURL)
	runTestCmd(suite.T(), []cmdTestCase{
		{
			wantError: true,
			name:      "records nothing from outside the run that attested, naming each attestation and the flags that pin a signer",
			cmd:       defaultSigner,
			golden:    otherRuns(intotoDeployURL, intotoBuildURL, intotoMainSigner, intotoBranchURL, intotoBranchSigner, intotoPullURL, intotoPullSigner),
		},
		{
			name:        "records main's attestation from outside its run once that signer is pinned, warning about the other two",
			cmd:         defaultSigner + " --dry-run " + intotoSigner,
			goldenRegex: `(?s)^\[warning\] another attestation was not recorded: attestation 2 failed verification: .*refs/heads/feature-x.*\[warning\] another attestation was not recorded: attestation 3 failed verification: .*refs/pull/12/merge.*"build_url": "` + intotoBuildURL + `".*"signer_identity": "` + intotoMainSigner + `"`,
		},
	})

	suite.githubHolds("branch-provenance.sigstore.json")
	runTestCmd(suite.T(), []cmdTestCase{
		{
			wantError: true,
			name:      "does not accept an attestation a branch made in another run",
			cmd:       defaultSigner,
			golden:    otherRuns(intotoDeployURL, intotoBranchURL, intotoBranchSigner),
		},
	})

	suite.githubHolds("pull-request-provenance.sigstore.json")
	runTestCmd(suite.T(), []cmdTestCase{
		{
			wantError: true,
			name:      "does not accept an attestation a pull request made in another run",
			cmd:       defaultSigner,
			golden:    otherRuns(intotoDeployURL, intotoPullURL, intotoPullSigner),
		},
	})

	suite.githubHolds("wrong-signer-provenance.sigstore.json")
	suite.inGitHubActions(intotoBuildURL)
	runTestCmd(suite.T(), []cmdTestCase{
		{
			wantError:   true,
			name:        "refuses an attestation signed by a workflow in another repository",
			cmd:         defaultSigner,
			goldenRegex: `^Error: GitHub repository acme/widget: the attestation failed verification: .*https://github.com/mallory/widget[^\n]*"\n$`,
		},
		{
			name:        "accepts it when that workflow is pinned explicitly",
			cmd:         defaultSigner + " --dry-run --certificate-identity-regexp ^https://github.com/mallory/ --certificate-oidc-issuer https://token.actions.githubusercontent.com",
			goldenRegex: `"signer_identity": "https://github.com/mallory/widget/.github/workflows/build.yml@refs/heads/main"`,
		},
	})
}

func (suite *AttestInTotoCommandTestSuite) TestGitHubSourceOnGitHubEnterprise() {
	suite.githubHolds("github-provenance.sigstore.json")
	// Nothing answers here, so a command that records read the server it was given.
	intotoGitHubAPI = "http://127.0.0.1:1/"

	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:        "reads the attestations and the commit from the server --github-base-url names",
			cmd:         suite.fromGitHubEnterprise("--source github --predicate-type provenance --name app --dry-run " + intotoSigner),
			goldenRegex: `(?s)"git_commit": "` + intotoCommit + `".*"commit_url": "` + intotoCommitURL + `".*"verified": "true"`,
		},
	})

	suite.inGitHubActions("https://github.com/acme/deploy/actions/runs/77/attempts/1")
	runTestCmd(suite.T(), []cmdTestCase{
		{
			wantError: true,
			name:      "has no default signer there, naming the flags that pin one",
			cmd:       suite.fromGitHubEnterprise("--predicate-type provenance --name app"),
			golden:    "Error: --source github has no default signer when --github-base-url is given: the default is a workflow on github.com. Pass --certificate-identity or --certificate-identity-regexp with --certificate-oidc-issuer, or --key\n",
		},
		{
			name:        "records in GitHub Actions when the signer is pinned",
			cmd:         suite.fromGitHubEnterprise("--predicate-type provenance --name app --dry-run " + intotoSigner),
			goldenRegex: `(?s)"git_commit": "` + intotoCommit + `".*"verified": "true"`,
		},
	})

	suite.inGitHubActions("https://github.example.com/acme/deploy/actions/runs/77/attempts/1")
	intotoGitHubAPI = suite.github.URL + "/"
	runTestCmd(suite.T(), []cmdTestCase{
		{
			wantError: true,
			name:      "has no default signer in a run on that server when --github-base-url is not given",
			cmd:       suite.fromGitHub("--predicate-type provenance --name app"),
			golden:    "Error: --source github has no default signer in a GitHub Actions run on a server other than github.com: the default is a workflow on github.com. Name the server with --github-base-url, and pass --certificate-identity or --certificate-identity-regexp with --certificate-oidc-issuer, or --key\n",
		},
	})
}

func (suite *AttestInTotoCommandTestSuite) TestGitHubSourceChoosesAmongSeveral() {
	suite.githubHolds("github-provenance.sigstore.json", "rerun-provenance.sigstore.json")
	suite.inGitHubActions(intotoBuildURL)
	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:        "prefers the attestation made by the run it is part of to a newer one",
			cmd:         suite.fromGitHub("--predicate-type provenance --name app --dry-run"),
			goldenRegex: `"build_url": "` + intotoBuildURL + `"`,
		},
	})

	suite.inGitHubActions(intotoDeployURL)
	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:        "takes the newest by the pinned signer when none was made by the run it is part of",
			cmd:         suite.fromGitHub("--predicate-type provenance --name app --dry-run " + intotoSigner),
			goldenRegex: `"build_url": "` + intotoRerunURL + `"`,
		},
	})

	suite.inGitHubActions(intotoBuildURL)
	suite.githubHolds("wrong-signer-provenance.sigstore.json", "github-provenance.sigstore.json")
	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:        "records the pinned workflow's attestation, warning about one by another signer",
			cmd:         suite.fromGitHub("--predicate-type provenance --name app --dry-run"),
			goldenRegex: `(?s)^\[warning\] another attestation was not recorded: attestation 1 failed verification: .*https://github.com/mallory/widget.*"signer_identity": "https://github.com/acme/widget/.github/workflows/build.yml@refs/heads/main"`,
		},
	})

	suite.githubHolds("trusted_root.json", "github-provenance.sigstore.json")
	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:        "records the attestation that verifies, warning about one it cannot read",
			cmd:         suite.fromGitHub("--predicate-type provenance --name app --dry-run"),
			goldenRegex: `(?s)^\[warning\] another attestation was not recorded: attestation 1: not a Sigstore bundle, a DSSE envelope or an in-toto statement.*"signer_identity": "https://github.com/acme/widget/.github/workflows/build.yml@refs/heads/main"`,
		},
	})

	suite.githubHolds(intotoGoneBundle, "github-provenance.sigstore.json")
	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:        "records the attestation that verifies, warning about one whose bundle it cannot download",
			cmd:         suite.fromGitHub("--predicate-type provenance --name app --dry-run"),
			goldenRegex: `(?s)^\[warning\] another attestation was not recorded: attestation 1: its bundle could not be downloaded from http://[^ ]+/blobs/gone: 404 Not Found\n.*"signer_identity": "https://github.com/acme/widget/.github/workflows/build.yml@refs/heads/main"`,
		},
	})

	suite.githubHolds(intotoGoneBundle)
	runTestCmd(suite.T(), []cmdTestCase{
		{
			wantError:   true,
			name:        "records nothing when the only attestation's bundle cannot be downloaded",
			cmd:         suite.fromGitHub("--predicate-type provenance --name app"),
			goldenRegex: `^Error: GitHub repository acme/widget: the attestation: its bundle could not be downloaded from http://[^ ]+/blobs/gone: 404 Not Found\n$`,
		},
	})

	suite.githubHolds("wrong-signer-provenance.sigstore.json", "tampered-provenance.sigstore.json")
	runTestCmd(suite.T(), []cmdTestCase{
		{
			wantError:   true,
			name:        "records nothing when none verifies, naming each failure",
			cmd:         suite.fromGitHub("--predicate-type provenance --name app"),
			goldenRegex: `^Error: GitHub repository acme/widget: 2 attestations could not be accepted: attestation 1 failed verification: .*; attestation 2 failed verification: `,
		},
	})
}

func (suite *AttestInTotoCommandTestSuite) TestGitHubSourceWhenItCannotRecord() {
	// CI sets a real token, which would satisfy the case that expects none.
	suite.T().Setenv("KOSLI_GITHUB_TOKEN", "")
	suite.githubHolds("spdx-sbom.sigstore.json")

	runTestCmd(suite.T(), []cmdTestCase{
		{
			wantError: true,
			name:      "GitHub holds only other kinds of attestation for the artifact, naming them",
			cmd:       suite.fromGitHub("--source github --predicate-type provenance --name app " + intotoSigner),
			golden:    "Error: GitHub repository acme/widget: found no provenance attestation (SLSA provenance v1 or v0.2); the predicate types found were: https://spdx.dev/Document/v2.3\n",
		},
		{
			wantError:   true,
			name:        "GitHub answers the listing with not found, naming both causes",
			cmd:         suite.command("--source github --github-org acme --repository gadget --predicate-type provenance --name app " + intotoSigner + suite.githubArguments),
			goldenRegex: `^Error: GitHub repository acme/gadget: failed to list its attestations: .*404 Not Found.*Either GitHub holds no attestation for the artifact, or the token cannot read the repository and its attestations \(the attestations: read permission\)`,
		},
		{
			wantError: true,
			name:      "no signer is pinned outside GitHub Actions, naming the flags that pin one",
			cmd:       suite.fromGitHub("--source github --predicate-type provenance --name app"),
			golden:    "Error: --source github needs a signer to verify attestations against. Pass --certificate-identity or --certificate-identity-regexp with --certificate-oidc-issuer, or --key. In GitHub Actions the signer defaults to a workflow in the repository itself, for an attestation the same run made\n",
		},
		{
			wantError: true,
			name:      "no --github-token",
			cmd:       suite.command("--source github --github-org acme --repository widget --predicate-type provenance --name app " + intotoSigner),
			golden:    "Error: --github-token is required with --source github\n",
		},
		{
			wantError: true,
			name:      "no --github-org or --repository, naming both",
			cmd:       suite.command("--source github --predicate-type provenance --name app " + intotoSigner + suite.githubArguments),
			golden:    "Error: --github-org and --repository are required with --source github\n",
		},
		{
			wantError: true,
			name:      "a --bundle as well",
			cmd:       suite.fromGitHub("--source github --predicate-type provenance --name app " + intotoSigner + suite.bundle("github-provenance.sigstore.json")),
			golden:    "Error: --bundle cannot be used with --source github\n",
		},
	})

	suite.githubHolds()
	runTestCmd(suite.T(), []cmdTestCase{
		{
			wantError: true,
			name:      "GitHub holds no attestation for the artifact",
			cmd:       suite.fromGitHub("--source github --predicate-type provenance --name app " + intotoSigner),
			golden:    "Error: GitHub repository acme/widget: it holds no attestation for the artifact (sha256:" + intotoDigest + ")\n",
		},
	})
}

func (suite *AttestInTotoCommandTestSuite) TestRejectsBadFlags() {
	runTestCmd(suite.T(), []cmdTestCase{
		{
			wantError: true,
			name:      "missing --predicate-type and --name, listing both",
			cmd:       suite.command(suite.bundle("github-provenance.sigstore.json")),
			golden:    "Error: required flag(s) \"name\", \"predicate-type\" not set\n",
		},
		{
			wantError: true,
			name:      "a predicate type it does not record",
			cmd:       suite.command("--predicate-type vuln --name app" + suite.bundle("github-provenance.sigstore.json")),
			golden:    "Error: predicate type \"vuln\" is not supported. Must be one of: provenance, sbom for --predicate-type\n",
		},
		{
			wantError: true,
			name:      "a source it does not read",
			cmd:       suite.command("--source oci --predicate-type provenance --name app"),
			golden:    "Error: --source \"oci\" is not supported. Must be one of: bundle, github\n",
		},
		{
			wantError: true,
			name:      "no --bundle",
			cmd:       suite.command("--predicate-type provenance --name app"),
			golden:    "Error: --bundle is required with --source bundle\n",
		},
		{
			wantError: true,
			name:      "an artifact name with a dot for provenance",
			cmd:       suite.command("--predicate-type provenance --name app.sbom" + suite.bundle("github-provenance.sigstore.json")),
			golden:    "Error: --name 'app.sbom' is not a valid artifact name. With --predicate-type provenance it names the artifact in the flow template, which can only contain letters, digits, '-' and '_'\n",
		},
		{
			wantError: true,
			name:      "an identity without an issuer",
			cmd:       suite.command("--predicate-type provenance --name app --certificate-identity someone" + suite.bundle("github-provenance.sigstore.json")),
			golden:    "Error: --certificate-oidc-issuer is required with --certificate-identity or --certificate-identity-regexp\n",
		},
		{
			wantError: true,
			name:      "an annotation the command sets itself",
			cmd:       suite.command("--predicate-type provenance --name app --annotate verified=true" + suite.bundle("github-provenance.sigstore.json")),
			golden:    "Error: annotation key 'verified' is set by this command from the attestation and cannot be provided with --annotate\n",
		},
		{
			wantError: true,
			name:      "neither --fingerprint nor --artifact-type",
			cmd:       "attest in-toto --predicate-type provenance --name app" + suite.bundle("github-provenance.sigstore.json") + suite.defaultKosliArguments,
			golden:    "Error: at least one of --fingerprint, --artifact-type is required\n",
		},
	})
}

func TestAttestationLink(t *testing.T) {
	found := func(logEntryURL, heldAt string) *intoto.Found {
		return &intoto.Found{
			Attestation: &intoto.Attestation{URL: heldAt},
			Evidence:    &intoto.Evidence{LogEntryURL: logEntryURL, RawSHA256: "abc123"},
		}
	}
	const listing = "https://api.github.com/repos/acme/widget/attestations/sha256:" + intotoDigest

	t.Run("links the transparency log entry, which anyone can read, when there is one", func(t *testing.T) {
		require.Equal(t, &URLInfo{Href: intotoLogEntryURL, Fingerprint: "abc123"}, attestationLink(found(intotoLogEntryURL, listing)))
	})

	t.Run("links where the source holds the attestation when it has no log entry", func(t *testing.T) {
		require.Equal(t, &URLInfo{Href: listing, Fingerprint: "abc123"}, attestationLink(found("", listing)))
	})

	t.Run("links nothing for an attestation read from a file with no log entry", func(t *testing.T) {
		require.Nil(t, attestationLink(found("", "")))
	})
}

func TestCurrentGitHubRun(t *testing.T) {
	t.Run("is the run's invocation URI, as its certificates name it, in GitHub Actions", func(t *testing.T) {
		t.Setenv("GITHUB_RUN_NUMBER", "7")
		t.Setenv("GITHUB_SERVER_URL", "https://github.com")
		t.Setenv("GITHUB_REPOSITORY", "acme/widget")
		t.Setenv("GITHUB_RUN_ID", "1234567890")
		t.Setenv("GITHUB_RUN_ATTEMPT", "1")

		runURI, inActions := currentGitHubRun()
		require.True(t, inActions)
		require.Equal(t, intotoBuildURL, runURI)
	})

	t.Run("is nothing outside GitHub Actions", func(t *testing.T) {
		// Setenv first, so that the variable is restored after it is unset.
		t.Setenv("GITHUB_RUN_NUMBER", "")
		require.NoError(t, os.Unsetenv("GITHUB_RUN_NUMBER"))

		runURI, inActions := currentGitHubRun()
		require.False(t, inActions)
		require.Empty(t, runURI)
	})
}

// localHeadCommit is a commit the repository these tests run in is sure to
// hold, standing in for the one a build's provenance names.
func localHeadCommit(t *testing.T) string {
	t.Helper()
	gitView, err := gitview.New("../..")
	require.NoError(t, err)
	head, err := gitView.GetCommitInfoFromCommitSHA("HEAD", false, nil)
	require.NoError(t, err)
	return head.Sha1
}

func regexpQuote(s string) string {
	return strings.NewReplacer(".", `\.`, "?", `\?`).Replace(s)
}

func TestAttestInTotoCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestInTotoCommandTestSuite))
}
