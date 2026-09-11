package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// oversizeSbom writes a file past the size ceiling outside the repo, so a
// ten megabyte fixture is not committed.
func (suite *AttestSbomCommandTestSuite) oversizeSbom() (string, int64) {
	path := filepath.Join(suite.T().TempDir(), "oversize.json")
	padding := strings.Repeat("x", maxSbomFileBytes+1)
	body := fmt.Sprintf(
		`{"bomFormat":"CycloneDX","specVersion":"1.6","version":1,"pad":"%s"}`, padding,
	)
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		suite.T().Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		suite.T().Fatal(err)
	}
	return path, info.Size()
}

func (suite *AttestSbomCommandTestSuite) TestAttestSbomCmd() {
	tests := []cmdTestCase{
		{
			name:   "reports a CycloneDX SBOM against a trail",
			cmd:    fmt.Sprintf("attest sbom --name cli --sbom-file testdata/sbom/cyclonedx.json %s", suite.defaultKosliArguments),
			golden: "sbom attestation 'cli' is reported to trail: test-123\n",
		},
		{
			name:   "reports an SPDX SBOM against a trail",
			cmd:    fmt.Sprintf("attest sbom --name cli --sbom-file testdata/sbom/spdx.json %s", suite.defaultKosliArguments),
			golden: "sbom attestation 'cli' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "fails when --sbom-file is missing",
			cmd:       fmt.Sprintf("attest sbom --name cli %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"sbom-file\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when --attachments is used as well, because two attachments would be compressed",
			cmd:       fmt.Sprintf("attest sbom --name cli --sbom-file testdata/sbom/cyclonedx.json --attachments testdata/sbom/spdx.json %s", suite.defaultKosliArguments),
			golden:    "Error: only one of --sbom-file, --attachments is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when the file is not an SBOM",
			cmd:       fmt.Sprintf("attest sbom --name cli --sbom-file testdata/sbom/not-an-sbom.json %s", suite.defaultKosliArguments),
			golden:    "Error: failed to parse SBOM file [testdata/sbom/not-an-sbom.json]: not a CycloneDX SBOM: bomFormat is \"\", expected \"CycloneDX\"\n",
		},
		{
			wantError: true,
			name:      "fails when the file is compressed, because the format is read from it",
			cmd:       fmt.Sprintf("attest sbom --name cli --sbom-file testdata/sbom/compressed.json.gz %s", suite.defaultKosliArguments),
			golden:    "Error: failed to parse SBOM file [testdata/sbom/compressed.json.gz]: the file is gzip compressed; supply the uncompressed SBOM\n",
		},
		{
			wantError: true,
			name:      "fails when a reserved annotation key is supplied",
			cmd:       fmt.Sprintf("attest sbom --name cli --sbom-file testdata/sbom/cyclonedx.json --annotate sbom_format=mine %s", suite.defaultKosliArguments),
			golden:    "Error: annotation key 'sbom_format' is set by this command from the SBOM file and cannot be provided with --annotate\n",
		},
	}
	runTestCmd(suite.T(), tests)
}

func (suite *AttestSbomCommandTestSuite) TestAttestSbomRejectsAnOversizeFileBeforeUploading() {
	path, size := suite.oversizeSbom()

	runTestCmd(suite.T(), []cmdTestCase{
		{
			wantError: true,
			name:      "fails locally with the actual size rather than a bare 413 from the server",
			cmd:       fmt.Sprintf("attest sbom --name cli --sbom-file %s %s", path, suite.defaultKosliArguments),
			golden:    fmt.Sprintf("Error: SBOM file [%s] is %d bytes, above the %d byte limit for an SBOM attestation\n", path, size, maxSbomFileBytes),
		},
	})
}

func TestAttestSbomCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestSbomCommandTestSuite))
}
