package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type SnapshotPathsTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	envName               string
}

func (suite *SnapshotPathsTestSuite) SetupSuite() {
	suite.envName = "snapshot-paths-env"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateEnv(global.Org, suite.envName, "server", suite.T())
}

func (suite *SnapshotPathsTestSuite) TestSnapshotPathsCmd() {
	tests := []cmdTestCase{
		{
			wantError:   true,
			name:        "fails when paths spec file does not exist",
			cmd:         fmt.Sprintf(`snapshot paths --paths-file testdata/paths-files/does-not-exist.yml %s %s`, suite.envName, suite.defaultKosliArguments),
			goldenRegex: "Error: failed to parse path spec file \\[testdata\\/paths-files\\/does-not-exist\\.yml\\] : Config File \"does-not-exist\" Not Found in \"\\[.*\\/cli\\/cmd\\/kosli\\/testdata\\/paths-files\\]\"\n",
		},
		{
			wantError: true,
			name:      "fails when paths spec file is invalid (fails to unmarshal)",
			cmd:       fmt.Sprintf(`snapshot paths --paths-file testdata/paths-files/invalid-pathsfile.yml %s %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: failed to unmarshal path spec file [testdata/paths-files/invalid-pathsfile.yml] : 1 error(s) decoding:\n\n* '' has invalid keys: foo, versionnn\n",
		},
		{
			wantError: true,
			name:      "fails when paths spec file is invalid (fails to validate)",
			cmd:       fmt.Sprintf(`snapshot paths --paths-file testdata/paths-files/invalid-values-pathsfile.yml %s %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: path spec file [testdata/paths-files/invalid-values-pathsfile.yml] is invalid: Key: 'PathsSpec.Version' Error:Field validation for 'Version' failed on the 'oneof' tag\n",
		},
		{
			wantError: true,
			name:      "fails when --paths-file and --path are provided",
			cmd:       fmt.Sprintf(`snapshot paths --paths-file testdata/paths-files/valid-pathsfile.yml --path foo %s %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: only one of --paths-file, --path is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when --paths-file and --exclude are provided",
			cmd:       fmt.Sprintf(`snapshot paths --paths-file testdata/paths-files/valid-pathsfile.yml --exclude foo %s %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: only one of --paths-file, --exclude is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when --paths-file and --name are provided",
			cmd:       fmt.Sprintf(`snapshot paths --paths-file testdata/paths-files/valid-pathsfile.yml --name foo %s %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: only one of --paths-file, --name is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when neither --paths-file nor --path are provided",
			cmd:       fmt.Sprintf(`snapshot paths %s %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: at least one of --paths-file, --path is required\n",
		},
		{
			name:   "can report artifact data with YAML path spec file",
			cmd:    fmt.Sprintf(`snapshot paths --paths-file testdata/paths-files/valid-pathsfile.yml %s %s`, suite.envName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("[1] artifacts were reported to environment %s\n", suite.envName),
		},
		{
			name:   "can report artifact data with JSON path spec file",
			cmd:    fmt.Sprintf(`snapshot paths --paths-file testdata/paths-files/valid-pathsfile.json %s %s`, suite.envName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("[1] artifacts were reported to environment %s\n", suite.envName),
		},
		{
			name:   "can report artifact data with TOML path spec file",
			cmd:    fmt.Sprintf(`snapshot paths --paths-file testdata/paths-files/valid-pathsfile.toml %s %s`, suite.envName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("[1] artifacts were reported to environment %s\n", suite.envName),
		},
		{
			wantError: true,
			name:      "--name is required when --path is used",
			cmd:       fmt.Sprintf(`snapshot paths --path testdata/file1 %s %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: flag --name is required when flag --path is set\n",
		},
		{
			name:   "can report artifact data with --path and --name",
			cmd:    fmt.Sprintf(`snapshot paths --path testdata/file1 --name foo %s %s`, suite.envName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("[1] artifacts were reported to environment %s\n", suite.envName),
		},
		{
			name:   "can report artifact data with --path and --exclude",
			cmd:    fmt.Sprintf(`snapshot paths --path testdata/server --name foo --exclude app.app,"**/logs.txt" %s %s`, suite.envName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("[1] artifacts were reported to environment %s\n", suite.envName),
		},
	}

	runTestCmd(suite.T(), tests)

}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSnapshotPathsTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotPathsTestSuite))
}
