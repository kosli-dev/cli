package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kosli-dev/cli/internal/cloudrun"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubCloudRunLister struct {
	services []cloudrun.Service
	err      error
}

func (s stubCloudRunLister) ListServices(_ context.Context, _, _ string) ([]cloudrun.Service, error) {
	return s.services, s.err
}

var origNewCloudRunClient = newCloudRunClient

type SnapshotCloudRunTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	envName               string
}

// stubServices returns two Cloud Run services so filter tests can verify
// inclusion and exclusion in a single run. Digests are full 64-char hex
// because the server's CloudRunReport model rejects anything else.
func stubServices() []cloudrun.Service {
	const (
		alphaDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		betaDigest  = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	)
	return []cloudrun.Service{
		{
			Name: "alpha",
			URI:  "https://alpha.run.app",
			Revisions: []cloudrun.Revision{
				{
					Name:      "alpha-rev1",
					Digests:   map[string]string{"gcr.io/x/alpha@sha256:" + alphaDigest: alphaDigest},
					CreatedAt: time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			Name: "beta",
			URI:  "https://beta.run.app",
			Revisions: []cloudrun.Revision{
				{
					Name:      "beta-rev1",
					Digests:   map[string]string{"gcr.io/x/beta@sha256:" + betaDigest: betaDigest},
					CreatedAt: time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC),
				},
			},
		},
	}
}

func (suite *SnapshotCloudRunTestSuite) SetupTest() {
	suite.envName = "snapshot-cloud-run-env"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	newCloudRunClient = func(_ context.Context) (cloudRunLister, error) {
		return stubCloudRunLister{services: stubServices()}, nil
	}

	CreateEnv(global.Org, suite.envName, "cloud-run", suite.T())
}

func (suite *SnapshotCloudRunTestSuite) TearDownTest() {
	newCloudRunClient = origNewCloudRunClient
}

func (suite *SnapshotCloudRunTestSuite) TestSnapshotCloudRunCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "01 snapshot cloud-run fails if no args are provided",
			cmd:       fmt.Sprintf(`snapshot cloud-run --project p --region r %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "02 snapshot cloud-run fails if 2 args are provided",
			cmd:       fmt.Sprintf(`snapshot cloud-run %s xxx --project p --region r %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "03 snapshot cloud-run fails if --project is missing",
			cmd:       fmt.Sprintf(`snapshot cloud-run %s --region r %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"project\" not set\n",
		},
		{
			wantError: true,
			name:      "04 snapshot cloud-run fails if --region is missing",
			cmd:       fmt.Sprintf(`snapshot cloud-run %s --project p %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"region\" not set\n",
		},
		{
			name:        "05 snapshot cloud-run dry-runs the report URL and payload built from the GCP client",
			cmd:         fmt.Sprintf(`snapshot cloud-run %s --project proj-x --region europe-west1 --dry-run %s`, suite.envName, suite.defaultKosliArguments),
			goldenRegex: `(?s)THIS IS A DRY-RUN.*report/cloud-run.*"type": "cloud-run".*"service_name": "alpha".*"service_name": "beta"`,
		},
		{
			wantError: true,
			name:      "06 snapshot cloud-run fails if --services and --exclude are set",
			cmd:       fmt.Sprintf(`snapshot cloud-run %s --project p --region r --services alpha --exclude beta %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: only one of --services, --exclude is allowed\n",
		},
		{
			wantError: true,
			name:      "07 snapshot cloud-run fails if --services and --exclude-regex are set",
			cmd:       fmt.Sprintf(`snapshot cloud-run %s --project p --region r --services alpha --exclude-regex "^b" %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: only one of --services, --exclude-regex is allowed\n",
		},
		{
			wantError: true,
			name:      "08 snapshot cloud-run fails if --services-regex and --exclude are set",
			cmd:       fmt.Sprintf(`snapshot cloud-run %s --project p --region r --services-regex "^a" --exclude beta %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: only one of --services-regex, --exclude is allowed\n",
		},
		{
			wantError: true,
			name:      "09 snapshot cloud-run fails if --services-regex and --exclude-regex are set",
			cmd:       fmt.Sprintf(`snapshot cloud-run %s --project p --region r --services-regex "^a" --exclude-regex "^b" %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: only one of --services-regex, --exclude-regex is allowed\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// runFilteredCmd executes the command and returns the combined output for
// substring assertions. Filter tests need to assert both presence (kept
// service appears) and absence (excluded service does not appear), so they
// cannot use the single-assertion cmdTestCase table.
func (suite *SnapshotCloudRunTestSuite) runFilteredCmd(filterArgs string) string {
	cmd := fmt.Sprintf(`snapshot cloud-run %s --project p --region r --dry-run %s %s`, suite.envName, filterArgs, suite.defaultKosliArguments)
	_, combined, _, _, err := executeCommandC(cmd)
	require.NoError(suite.T(), err, "command failed: %s", combined)
	return combined
}

func (suite *SnapshotCloudRunTestSuite) TestSnapshotCloudRunFilter_Services() {
	out := suite.runFilteredCmd("--services alpha")
	require.Contains(suite.T(), out, `"service_name": "alpha"`)
	require.NotContains(suite.T(), out, `"service_name": "beta"`)
}

func (suite *SnapshotCloudRunTestSuite) TestSnapshotCloudRunFilter_ServicesRegex() {
	out := suite.runFilteredCmd(`--services-regex "^al"`)
	require.Contains(suite.T(), out, `"service_name": "alpha"`)
	require.NotContains(suite.T(), out, `"service_name": "beta"`)
}

func (suite *SnapshotCloudRunTestSuite) TestSnapshotCloudRunFilter_Exclude() {
	out := suite.runFilteredCmd("--exclude alpha")
	require.NotContains(suite.T(), out, `"service_name": "alpha"`)
	require.Contains(suite.T(), out, `"service_name": "beta"`)
}

func (suite *SnapshotCloudRunTestSuite) TestSnapshotCloudRunFilter_ExcludeRegex() {
	out := suite.runFilteredCmd(`--exclude-regex "^al"`)
	require.NotContains(suite.T(), out, `"service_name": "alpha"`)
	require.Contains(suite.T(), out, `"service_name": "beta"`)
}

// TestSnapshotCloudRunCmd_HappyPathReportsToServer exercises the full
// CLI → local Kosli server roundtrip with the GCP client stubbed: the env is
// already created in SetupTest with type "cloud-run", and the command is
// expected to PUT the snapshot and emit the "[N] revisions were reported"
// success log mirroring ECS.
func (suite *SnapshotCloudRunTestSuite) TestSnapshotCloudRunCmd_HappyPathReportsToServer() {
	cmd := fmt.Sprintf(`snapshot cloud-run %s --project p --region r %s`, suite.envName, suite.defaultKosliArguments)
	_, combined, _, _, err := executeCommandC(cmd)

	require.NoError(suite.T(), err, "command failed: %s", combined)
	require.Contains(suite.T(), combined, fmt.Sprintf("[2] revisions were reported to environment %s", suite.envName))
}

// TestSnapshotCloudRunCmd_UnauthenticatedReturnsFriendlyError verifies that a
// gRPC Unauthenticated error from GCP surfaces as the actionable ADC message
// rather than a raw SDK string.
func (suite *SnapshotCloudRunTestSuite) TestSnapshotCloudRunCmd_UnauthenticatedReturnsFriendlyError() {
	newCloudRunClient = func(_ context.Context) (cloudRunLister, error) {
		return stubCloudRunLister{err: status.Error(codes.Unauthenticated, "token expired")}, nil
	}

	cmd := fmt.Sprintf(`snapshot cloud-run %s --project p --region r %s`, suite.envName, suite.defaultKosliArguments)
	_, combined, _, _, err := executeCommandC(cmd)

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "GCP authentication failed")
	require.Contains(suite.T(), combined, "metadata server")
}

func TestSnapshotCloudRunCommandTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotCloudRunTestSuite))
}
