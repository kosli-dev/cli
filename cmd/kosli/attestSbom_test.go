package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kosli-dev/cli/internal/intoto"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AttestSbomCommandTestSuite struct {
	flowName  string
	trailName string
	suite.Suite
	defaultKosliArguments string
}

func (suite *AttestSbomCommandTestSuite) SetupTest() {
	suite.flowName = "attest-sbom"
	suite.trailName = "test-123"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(
		" --flow %s --trail %s --repo-root ../.. --host %s --org %s --api-token %s",
		suite.flowName, suite.trailName, global.Host, global.Org, global.ApiToken,
	)
	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.T())
}

// The dry-run cases are what prove the payload this command builds. They need a
// server for the flow and trail in SetupTest, but not one that knows the sbom
// type, so they stay green while the server side is unreleased.
func (suite *AttestSbomCommandTestSuite) TestAttestSbomBuildsTheRightRequest() {
	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:        "posts to the system endpoint with type_name sbom and the CycloneDX summary",
			cmd:         fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/cyclonedx.json --dry-run %s", suite.defaultKosliArguments),
			goldenRegex: `(?s)trail/test-123/system.*"type_name": "sbom".*"format": "cyclonedx-1\.6".*"original_fingerprint": "db09ef115d88e48a5ef553b21a88ccdc15b3df700e0b7c3736e2ef1024d26d9c".*"package_count": 1`,
		},
		{
			name:        "records the format and the file checksum as annotations",
			cmd:         fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/cyclonedx.json --dry-run %s", suite.defaultKosliArguments),
			goldenRegex: `(?s)"sbom_format": "cyclonedx-1\.6".*"sbom_sha256": "db09ef115d88e48a5ef553b21a88ccdc15b3df700e0b7c3736e2ef1024d26d9c"`,
		},
		{
			name:        "keeps the caller's own annotations alongside the two it derives",
			cmd:         fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/cyclonedx.json --annotate team=platform --dry-run %s", suite.defaultKosliArguments),
			goldenRegex: `"team": "platform"`,
		},
		{
			name:        "still derives its own annotations when the caller supplies one",
			cmd:         fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/cyclonedx.json --annotate team=platform --dry-run %s", suite.defaultKosliArguments),
			goldenRegex: `"sbom_format": "cyclonedx-1\.6"`,
		},
		{
			name:        "reads SPDX as well, and reports the version the file declares",
			cmd:         fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/spdx.json --dry-run %s", suite.defaultKosliArguments),
			goldenRegex: `(?s)"type_name": "sbom".*"format": "spdx-2\.3"`,
		},
	})
}

// An SBOM signed into an attestation is recorded with the attestation as the
// attachment, and verified when a signer is pinned.
func (suite *AttestSbomCommandTestSuite) TestAttestSbomUnwrapsAnAttestation() {
	trustedRoot, err := root.NewTrustedRootFromPath(filepath.Join(intotoFixtures, "trusted_root.json"))
	require.NoError(suite.T(), err)
	saved := intotoTrustRoots
	intotoTrustRoots = func() *intoto.TrustRoots { return intoto.FixedTrustRoots(trustedRoot) }
	defer func() { intotoTrustRoots = saved }()

	bundle := filepath.Join(intotoFixtures, "spdx-sbom.sigstore.json")
	bundleSHA256 := fixtureSHA256(suite.T(), "spdx-sbom.sigstore.json")
	envelope := filepath.Join(intotoFixtures, "unsigned-spdx-sbom.dsse.json")
	command := func(flags string) string {
		return "attest sbom --name my-sbom " + flags + suite.defaultKosliArguments
	}

	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:        "reads the SBOM from a Sigstore bundle and fingerprints the bundle it attaches",
			cmd:         command("--dry-run --sbom-file " + bundle),
			goldenRegex: `(?s)"filename": "spdx-sbom.sigstore.json".*"sbom_sha256": "` + bundleSHA256 + `",\s+"verified": "false".*"format": "spdx-2\.3",\s+"original_fingerprint": "` + bundleSHA256 + `"`,
		},
		{
			name:        "verifies the bundle against the pinned signer and records the evidence",
			cmd:         command("--dry-run --sbom-file " + bundle + " " + intotoSigner),
			goldenRegex: `(?s)"external_urls": \{\s+"attestation": \{\s+"href": "` + regexpQuote(intotoLogEntryURL) + `",\s+"fingerprint": "` + bundleSHA256 + `".*"signer_identity": "https://github.com/acme/widget/.github/workflows/build.yml@refs/heads/main".*"verified": "true"`,
		},
		{
			name:        "checks the subject when the SBOM is about an artifact",
			cmd:         command("--dry-run --sbom-file " + bundle + " " + intotoSigner + " --fingerprint " + intotoDigest),
			goldenRegex: `"verified": "true"`,
		},
		{
			wantError: true,
			name:      "refuses a verified SBOM about another artifact",
			cmd:       command("--sbom-file " + bundle + " " + intotoSigner + " --fingerprint 11de96226896b0fcf4ea66a11a63a5fd9f4e211fe94b62b7eebf4485c817b2d8"),
			golden:    "Error: SBOM file [" + bundle + "]: the attestation does not list the artifact (sha256:11de96226896b0fcf4ea66a11a63a5fd9f4e211fe94b62b7eebf4485c817b2d8) as its subject\n",
		},
		{
			wantError:   true,
			name:        "refuses a bundle signed by someone other than the pinned signer",
			cmd:         command("--sbom-file " + bundle + " --certificate-identity https://github.com/mallory/widget/.github/workflows/build.yml@refs/heads/main --certificate-oidc-issuer https://token.actions.githubusercontent.com"),
			goldenRegex: `^Error: SBOM file \[.*spdx-sbom.sigstore.json\]: the attestation failed verification: `,
		},
		{
			name:        "reads the SBOM from an unsigned DSSE envelope, marked unverified",
			cmd:         command("--dry-run --sbom-file " + envelope),
			goldenRegex: `(?s)"verified": "false".*"format": "spdx-2\.3"`,
		},
		{
			wantError: true,
			name:      "refuses to verify an unsigned DSSE envelope",
			cmd:       command("--sbom-file " + envelope + " " + intotoSigner),
			golden:    "Error: SBOM file [" + envelope + "] is not signed, so it cannot be verified against --certificate-identity, --certificate-identity-regexp or --key\n",
		},
		{
			wantError: true,
			name:      "refuses an attestation that is not an SBOM",
			cmd:       command("--sbom-file " + filepath.Join(intotoFixtures, "github-provenance.sigstore.json")),
			golden:    "Error: SBOM file [" + filepath.Join(intotoFixtures, "github-provenance.sigstore.json") + "]: the attestation is https://slsa.dev/provenance/v1, not an SPDX or CycloneDX SBOM\n",
		},
		{
			wantError: true,
			name:      "refuses to verify a plain SBOM file, which has no signature",
			cmd:       command("--sbom-file testdata/sbom/cyclonedx.json " + intotoSigner),
			golden:    "Error: SBOM file [testdata/sbom/cyclonedx.json] is not a Sigstore bundle or DSSE envelope, so it has no signature to verify against --certificate-identity, --certificate-identity-regexp or --key\n",
		},
		{
			name:        "still lets a plain SBOM file be annotated with a key the verification would set",
			cmd:         command("--dry-run --sbom-file testdata/sbom/cyclonedx.json --annotate verified=by-hand"),
			goldenRegex: `"verified": "by-hand"`,
		},
		{
			wantError: true,
			name:      "reserves the verification annotations for an SBOM in an attestation",
			cmd:       command("--sbom-file " + bundle + " --annotate verified=by-hand"),
			golden:    "Error: annotation key 'verified' is set by this command from the attestation and cannot be provided with --annotate\n",
		},
	})
}

func (suite *AttestSbomCommandTestSuite) TestAttestSbomRejectsBadInput() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when --sbom-file is missing",
			cmd:       fmt.Sprintf("attest sbom --name my-sbom %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"sbom-file\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when --attachments is used too, because two attachments would be compressed",
			cmd:       fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/cyclonedx.json --attachments testdata/sbom/spdx.json %s", suite.defaultKosliArguments),
			golden:    "Error: --attachments cannot be used with attest sbom: the SBOM file is the only attachment, and a second one would be compressed\n",
		},
		{
			// The command wraps the parser's message, and only this exercises
			// that wrapping.
			wantError: true,
			name:      "fails when the file is gzipped, because the format is read from it",
			cmd:       fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/compressed.json.gz %s", suite.defaultKosliArguments),
			golden:    "Error: failed to parse SBOM file [testdata/sbom/compressed.json.gz]: the file is gzip compressed; supply the uncompressed SBOM\n",
		},
		{
			wantError: true,
			name:      "fails when the file is not an SBOM, naming the file",
			cmd:       fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/not-an-sbom.json %s", suite.defaultKosliArguments),
			golden:    "Error: failed to parse SBOM file [testdata/sbom/not-an-sbom.json]: not a CycloneDX SBOM: bomFormat is \"\", expected \"CycloneDX\"\n",
		},
		{
			wantError: true,
			name:      "fails when a reserved annotation key is supplied",
			cmd:       fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/cyclonedx.json --annotate sbom_format=mine %s", suite.defaultKosliArguments),
			golden:    "Error: annotation key 'sbom_format' is set by this command from the SBOM file and cannot be provided with --annotate\n",
		},
		{
			wantError: true,
			name:      "fails when the checksum annotation key is supplied, not only the format one",
			cmd:       fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/cyclonedx.json --annotate sbom_sha256=mine %s", suite.defaultKosliArguments),
			golden:    "Error: annotation key 'sbom_sha256' is set by this command from the SBOM file and cannot be provided with --annotate\n",
		},
		{
			wantError: true,
			name:      "fails when the path is a directory rather than a file",
			cmd:       fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom %s", suite.defaultKosliArguments),
			golden:    "Error: SBOM file [testdata/sbom] is a directory; supply the SBOM file itself\n",
		},
	}
	runTestCmd(suite.T(), tests)
}

// A shared CI env block is how several attest steps get configured at once, so
// KOSLI_ATTACHMENTS reaches this command without anyone typing the flag. The
// flag is hidden from help, so the error has to say where the value came from,
// and only when that is where it came from.
func (suite *AttestSbomCommandTestSuite) TestAttestSbomRejectsAttachmentsFromTheEnvironment() {
	suite.T().Setenv("KOSLI_ATTACHMENTS", "testdata/sbom/spdx.json")
	runTestCmd(suite.T(), []cmdTestCase{
		{
			wantError: true,
			name:      "names the environment variable the user did not type",
			cmd:       fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/cyclonedx.json %s", suite.defaultKosliArguments),
			golden:    "Error: --attachments cannot be used with attest sbom (set by environment variable KOSLI_ATTACHMENTS): the SBOM file is the only attachment, and a second one would be compressed\n",
		},
		{
			// The typed flag wins and the variable is never applied, so blaming
			// it would send the user to edit something that had no effect.
			wantError: true,
			name:      "does not blame the environment when the flag was typed as well",
			cmd:       fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/cyclonedx.json --attachments testdata/sbom/spdx.json %s", suite.defaultKosliArguments),
			golden:    "Error: --attachments cannot be used with attest sbom: the SBOM file is the only attachment, and a second one would be compressed\n",
		},
	})
}

func (suite *AttestSbomCommandTestSuite) TestAttestSbomRejectsAttachmentsFromAConfigFile() {
	configFile := filepath.Join(suite.T().TempDir(), "kosli.yml")
	if err := os.WriteFile(configFile, []byte("attachments: testdata/sbom/spdx.json\n"), 0644); err != nil {
		suite.T().Fatal(err)
	}
	runTestCmd(suite.T(), []cmdTestCase{
		{
			wantError: true,
			name:      "names the config file the value came from",
			cmd:       fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/cyclonedx.json --config-file %s %s", configFile, suite.defaultKosliArguments),
			golden:    fmt.Sprintf("Error: --attachments cannot be used with attest sbom (set by config file [%s]): the SBOM file is the only attachment, and a second one would be compressed\n", configFile),
		},
	})
}

// The only test that exercises the command against a real server. It fails
// with "System attestation type 'sbom' does not exist" until the server that
// CI tests against carries the type; that red is accurate and is not to be
// skipped around.
func (suite *AttestSbomCommandTestSuite) TestAttestSbomRoundTrip() {
	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:   "reports a CycloneDX SBOM against a trail",
			cmd:    fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/cyclonedx.json %s", suite.defaultKosliArguments),
			golden: "sbom attestation 'my-sbom' is reported to trail: test-123\n",
		},
	})
}

// The recorded checksum is only verifiable by hand if the bytes the server
// receives are the bytes that were hashed. The uploader is handed those bytes,
// not a path it would read again, so the same buffer serves the fingerprint,
// the summary and the body. Outside the suite because it asserts local logic
// and should still report when no server is running.
func TestSbomUploadsTheBytesItHashed(t *testing.T) {
	o := &attestSbomOptions{
		CommonAttestationOptions: &CommonAttestationOptions{fingerprintOptions: &fingerprintOptions{}},
		sbomFilePath:             "testdata/sbom/cyclonedx.json",
		payload:                  SbomAttestationPayload{CommonAttestationPayload: &CommonAttestationPayload{}, TypeName: "sbom"},
	}
	content, err := o.loadSbom()
	require.NoError(t, err)

	fb := o.evidence(content)
	require.Equal(t, "cyclonedx.json", fb.Name)

	uploaded := fmt.Sprintf("%x", sha256.Sum256(fb.Data))
	require.Equal(t, o.payload.AttestationData.OriginalFingerprint, uploaded, "original_fingerprint must describe the uploaded bytes")
	require.Equal(t, o.payload.Annotations[sbomSha256Annotation], uploaded, "sbom_sha256 must describe the uploaded bytes")
	// Anchor: the digest is of real content, not of an empty buffer.
	require.Equal(t, "db09ef115d88e48a5ef553b21a88ccdc15b3df700e0b7c3736e2ef1024d26d9c", uploaded)
}

func TestAttestSbomCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestSbomCommandTestSuite))
}
