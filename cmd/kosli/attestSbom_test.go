package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/kosli-dev/cli/internal/requests"
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

// sizedSbom makes a file of the given size without allocating it: past the
// first byte it is zeros the filesystem never stored. The leading "{" sends the
// parser down the JSON branch, which fails on the second byte, rather than the
// fallback that regex-scans the whole buffer three times.
func (suite *AttestSbomCommandTestSuite) sizedSbom(name string, size int64) string {
	path := filepath.Join(suite.T().TempDir(), name)
	if err := os.WriteFile(path, []byte("{"), 0644); err != nil {
		suite.T().Fatal(err)
	}
	if err := os.Truncate(path, size); err != nil {
		suite.T().Fatal(err)
	}
	return path
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

func (suite *AttestSbomCommandTestSuite) TestAttestSbomSizeLimit() {
	over := suite.sizedSbom("oversize.json", maxSbomFileBytes+1)
	atLimit := suite.sizedSbom("at-limit.json", maxSbomFileBytes)

	runTestCmd(suite.T(), []cmdTestCase{
		{
			wantError: true,
			name:      "fails locally on an oversize file, rather than with a bare 413 from the server",
			cmd:       fmt.Sprintf("attest sbom --name my-sbom --sbom-file %s %s", over, suite.defaultKosliArguments),
			golden:    fmt.Sprintf("Error: SBOM file [%s] is above the %d byte limit for an SBOM attestation\n", over, maxSbomFileBytes),
		},
		{
			// Reaching the parser is the point: it proves the limit is the
			// largest accepted size and not the smallest rejected one.
			wantError:   true,
			name:        "a file of exactly the limit gets past the size guard",
			cmd:         fmt.Sprintf("attest sbom --name my-sbom --sbom-file %s %s", atLimit, suite.defaultKosliArguments),
			goldenRegex: `failed to parse SBOM file .* not valid JSON`,
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

	form := o.attestationForm(content)
	require.Len(t, form, 2, "the JSON payload and exactly one attachment")
	require.Equal(t, "file-bytes", form[1].Type, "a path here would be read a second time by the uploader")
	fb, ok := form[1].Content.(requests.FileBytes)
	require.True(t, ok)
	require.Equal(t, "cyclonedx.json", fb.Name)

	uploaded := fmt.Sprintf("%x", sha256.Sum256(fb.Data))
	require.Equal(t, o.payload.AttestationData.OriginalFingerprint, uploaded, "original_fingerprint must describe the uploaded bytes")
	require.Equal(t, o.payload.Annotations[sbomSha256Annotation], uploaded, "sbom_sha256 must describe the uploaded bytes")
	// Anchor: the digest is of real content, not of an empty buffer.
	require.Equal(t, "db09ef115d88e48a5ef553b21a88ccdc15b3df700e0b7c3736e2ef1024d26d9c", uploaded)
}

// The only guard in loadSbom without a test, and it could not have one while the
// check ran on an already-open handle: open(2) on a fifo with no writer blocks,
// so a reversed order hangs rather than failing. The timeout turns that hang
// into a named failure, which makes the ordering itself something a red run can
// report, not just the branch.
func TestSbomRejectsAFifoWithoutBlocking(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no mkfifo on windows")
	}
	path := filepath.Join(t.TempDir(), "bom.json")
	require.NoError(t, syscall.Mkfifo(path, 0644))

	o := &attestSbomOptions{
		CommonAttestationOptions: &CommonAttestationOptions{fingerprintOptions: &fingerprintOptions{}},
		sbomFilePath:             path,
		payload:                  SbomAttestationPayload{CommonAttestationPayload: &CommonAttestationPayload{}},
	}
	done := make(chan error, 1)
	go func() { _, err := o.loadSbom(); done <- err }()

	select {
	case err := <-done:
		require.ErrorContains(t, err, "is not a regular file")
	case <-time.After(5 * time.Second):
		t.Fatal("loadSbom blocked on a fifo: the stat must precede the open")
	}
}

func TestAttestSbomCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestSbomCommandTestSuite))
}
