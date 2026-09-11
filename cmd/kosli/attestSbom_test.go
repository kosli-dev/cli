package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

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

// oversizeSbom makes a file past the size ceiling without allocating it. The
// guard only stats the file, so the content is never read.
func (suite *AttestSbomCommandTestSuite) oversizeSbom() (string, int64) {
	path := filepath.Join(suite.T().TempDir(), "oversize.json")
	size := int64(maxSbomFileBytes + 1)
	if err := os.WriteFile(path, nil, 0644); err != nil {
		suite.T().Fatal(err)
	}
	if err := os.Truncate(path, size); err != nil {
		suite.T().Fatal(err)
	}
	return path, size
}

// The dry-run cases are what prove the payload this command builds. They need a
// server for the flow and trail in SetupTest, but not one that knows the sbom
// type, so they stay green while the server side is unreleased.
func (suite *AttestSbomCommandTestSuite) TestAttestSbomBuildsTheRightRequest() {
	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:        "posts to the system endpoint with type_name sbom and the CycloneDX summary",
			cmd:         fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/cyclonedx.json --dry-run %s", suite.defaultKosliArguments),
			goldenRegex: `(?s)trail/test-123/system.*"type_name": "sbom".*"format": "cyclonedx-1\.6".*"original_fingerprint": "[a-f0-9]{64}".*"package_count": 1`,
		},
		{
			name:        "records the format and the file checksum as annotations",
			cmd:         fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/cyclonedx.json --dry-run %s", suite.defaultKosliArguments),
			goldenRegex: `(?s)"sbom_format": "cyclonedx-1\.6".*"sbom_sha256": "[a-f0-9]{64}"`,
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
			golden:    "Error: [kosli attest sbom flow=attest-sbom trail=test-123] required flag(s) \"sbom-file\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when --attachments is used too, because two attachments would be compressed",
			cmd:       fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/cyclonedx.json --attachments testdata/sbom/spdx.json %s", suite.defaultKosliArguments),
			golden:    "Error: [kosli attest sbom flow=attest-sbom trail=test-123] only one of --sbom-file, --attachments is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when the file is not an SBOM, naming the file",
			cmd:       fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/not-an-sbom.json --dry-run %s", suite.defaultKosliArguments),
			golden:    "Error: [kosli attest sbom flow=attest-sbom trail=test-123] failed to parse SBOM file [testdata/sbom/not-an-sbom.json]: not a CycloneDX SBOM: bomFormat is \"\", expected \"CycloneDX\"\n",
		},
		{
			wantError: true,
			name:      "fails when a reserved annotation key is supplied",
			cmd:       fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/cyclonedx.json --annotate sbom_format=mine --dry-run %s", suite.defaultKosliArguments),
			golden:    "Error: [kosli attest sbom flow=attest-sbom trail=test-123] annotation key 'sbom_format' is set by this command from the SBOM file and cannot be provided with --annotate\n",
		},
		{
			wantError: true,
			name:      "fails when the path is a directory rather than a file",
			cmd:       fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom --dry-run %s", suite.defaultKosliArguments),
			golden:    "Error: [kosli attest sbom flow=attest-sbom trail=test-123] SBOM file [testdata/sbom] is not a regular file\n",
		},
	}
	runTestCmd(suite.T(), tests)
}

func (suite *AttestSbomCommandTestSuite) TestAttestSbomRejectsAnOversizeFileBeforeReadingIt() {
	path, size := suite.oversizeSbom()

	runTestCmd(suite.T(), []cmdTestCase{
		{
			wantError: true,
			name:      "fails locally with the actual size, rather than a bare 413 from the server",
			cmd:       fmt.Sprintf("attest sbom --name my-sbom --sbom-file %s --dry-run %s", path, suite.defaultKosliArguments),
			golden:    fmt.Sprintf("Error: [kosli attest sbom flow=attest-sbom trail=test-123] SBOM file [%s] is %d bytes, above the %d byte limit for an SBOM attestation\n", path, size, maxSbomFileBytes),
		},
	})
}

// Reporting to a real server needs one that knows the sbom type, which arrives
// with kosli-dev/server#6863.
func (suite *AttestSbomCommandTestSuite) TestAttestSbomRoundTrip() {
	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:   "reports a CycloneDX SBOM against a trail",
			cmd:    fmt.Sprintf("attest sbom --name my-sbom --sbom-file testdata/sbom/cyclonedx.json %s", suite.defaultKosliArguments),
			golden: "sbom attestation 'my-sbom' is reported to trail: test-123\n",
		},
	})
}

func TestAttestSbomCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestSbomCommandTestSuite))
}
