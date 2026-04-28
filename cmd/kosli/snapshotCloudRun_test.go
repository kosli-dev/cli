package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kosli-dev/cli/internal/cloudrun"
	"github.com/stretchr/testify/suite"
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

func (suite *SnapshotCloudRunTestSuite) SetupTest() {
	suite.envName = "snapshot-cloud-run-env"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	newCloudRunClient = func(_ context.Context) (cloudRunLister, error) {
		return stubCloudRunLister{
			services: []cloudrun.Service{
				{
					Name: "hello-world",
					URI:  "https://hello-world.run.app",
					Revisions: []cloudrun.Revision{
						{
							Name:      "hello-world-rev1",
							Digests:   map[string]string{"gcr.io/x/hello@sha256:abc": "abc"},
							CreatedAt: time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC),
						},
					},
				},
			},
		}, nil
	}
}

func (suite *SnapshotCloudRunTestSuite) TearDownTest() {
	newCloudRunClient = origNewCloudRunClient
}

func (suite *SnapshotCloudRunTestSuite) TestSnapshotCloudRunCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "snapshot cloud-run fails if no args are provided",
			cmd:       fmt.Sprintf(`snapshot cloud-run --project p --region r %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "snapshot cloud-run fails if 2 args are provided",
			cmd:       fmt.Sprintf(`snapshot cloud-run %s xxx --project p --region r %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "snapshot cloud-run fails if --project is missing",
			cmd:       fmt.Sprintf(`snapshot cloud-run %s --region r %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"project\" not set\n",
		},
		{
			wantError: true,
			name:      "snapshot cloud-run fails if --region is missing",
			cmd:       fmt.Sprintf(`snapshot cloud-run %s --project p %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"region\" not set\n",
		},
		{
			name:        "snapshot cloud-run dry-runs the report URL and payload built from the GCP client",
			cmd:         fmt.Sprintf(`snapshot cloud-run %s --project proj-x --region europe-west1 %s`, suite.envName, suite.defaultKosliArguments),
			goldenRegex: `(?s)THIS IS A DRY-RUN.*report/cloud-run.*"revisionName": "hello-world-rev1".*"service_name": "hello-world".*"project": "proj-x".*"region": "europe-west1".*"gcr.io/x/hello@sha256:abc": "abc"`,
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestSnapshotCloudRunCommandTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotCloudRunTestSuite))
}
