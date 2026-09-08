package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

// WorkingDirConfigTestSuite covers the config file the CLI must NOT load: one
// sitting in the current working directory that the user never named. Loading
// it lets the contents of a checkout set host, http-proxy or kubeconfig for a
// command run with a real API token (kosli-dev/server#6778, #6779).
type WorkingDirConfigTestSuite struct {
	suite.Suite
}

func (suite *WorkingDirConfigTestSuite) TearDownTest() {
	defaultConfigFilePathFunc = (&RealConfigGetter{}).defaultConfigFilePath
	global = new(GlobalOpts)
}

// stubHomeConfig points the default config path at a file that does not exist,
// which is the state that used to trigger the working-directory fallback.
func (suite *WorkingDirConfigTestSuite) stubHomeConfig() string {
	path := filepath.Join(suite.T().TempDir(), defaultConfigFilename)
	mockConfigGetter := new(MockConfigGetter)
	mockConfigGetter.Mock.On("defaultConfigFilePath").Return(path)
	defaultConfigFilePathFunc = mockConfigGetter.defaultConfigFilePath
	return path
}

// chdirWithConfig writes content to a named config file in a temp directory and
// makes that directory the working directory for the test.
func (suite *WorkingDirConfigTestSuite) chdirWithConfig(name, content string) {
	dir := suite.T().TempDir()
	suite.Require().NoError(os.WriteFile(filepath.Join(dir, name), []byte(content), 0600))
	suite.T().Chdir(dir)
}

func (suite *WorkingDirConfigTestSuite) TestDefaultIsHomePathWhenHomeConfigIsAbsent() {
	path := suite.stubHomeConfig()

	_, _, _, _, err := executeCommandC("version")

	suite.Require().NoError(err)
	suite.Equal(path, global.ConfigFile,
		"the default config file must be the home path, not a bare name that resolves to the working directory")
}

func (suite *WorkingDirConfigTestSuite) TestWorkingDirConfigIsNotLoaded() {
	for _, name := range []string{"kosli.yml", "kosli.yaml", "kosli.json", "kosli.toml"} {
		suite.Run(name, func() {
			defer func() { global = new(GlobalOpts) }()
			suite.stubHomeConfig()
			content := "host: https://attacker.example\n"
			switch filepath.Ext(name) {
			case ".json":
				content = `{"host": "https://attacker.example"}`
			case ".toml":
				content = `host = "https://attacker.example"`
			}
			suite.chdirWithConfig(name, content)

			_, _, _, _, err := executeCommandC("version")

			suite.Require().NoError(err)
			suite.Equal(defaultHost, global.Host,
				"a config file in the working directory must not set the host")
		})
	}
}

func (suite *WorkingDirConfigTestSuite) TestExplicitConfigFileFlagStillLoadsWorkingDirConfig() {
	suite.stubHomeConfig()
	suite.chdirWithConfig("kosli.yml", "host: https://named.example\n")

	_, _, _, stderr, err := executeCommandC("version --config-file kosli.yml")

	suite.Require().NoError(err)
	suite.Equal("https://named.example", global.Host,
		"a config file the user names must still be loaded")
	suite.NotContains(stderr, "no longer loaded automatically",
		"naming the file is the supported way to load it, so there is nothing to warn about")
}

func (suite *WorkingDirConfigTestSuite) TestConfigFileEnvVarStillLoadsWorkingDirConfig() {
	suite.stubHomeConfig()
	suite.chdirWithConfig("kosli.yml", "host: https://named.example\n")
	suite.T().Setenv("KOSLI_CONFIG_FILE", "kosli.yml")

	_, _, _, stderr, err := executeCommandC("version")

	suite.Require().NoError(err)
	suite.Equal("https://named.example", global.Host,
		"KOSLI_CONFIG_FILE must still be able to name a working-directory file")
	suite.NotContains(stderr, "no longer loaded automatically")
}

// TestEmptyDefaultConfigPathLoadsNothing covers the case where no home
// directory can be resolved, so defaultConfigFilePath returns no path at all.
// The working directory must not become the fallback search location. The empty
// path is injected because homedir.Dir cannot be made to fail portably: on
// macOS it falls back to dscl even with HOME unset.
func (suite *WorkingDirConfigTestSuite) TestEmptyDefaultConfigPathLoadsNothing() {
	mockConfigGetter := new(MockConfigGetter)
	mockConfigGetter.Mock.On("defaultConfigFilePath").Return("")
	defaultConfigFilePathFunc = mockConfigGetter.defaultConfigFilePath
	suite.chdirWithConfig("kosli.yml", "host: https://attacker.example\n")

	_, _, _, _, err := executeCommandC("version")

	suite.Require().NoError(err)
	suite.Equal(defaultHost, global.Host,
		"with no default config path the working directory must not be searched instead")
}

func (suite *WorkingDirConfigTestSuite) TestWarnsAboutIgnoredWorkingDirConfig() {
	cases := []struct {
		name     string
		content  string
		wantWarn bool
	}{
		{name: "host", content: "host: https://attacker.example\n", wantWarn: true},
		{name: "org", content: "org: some-org\n", wantWarn: true},
		{name: "api token", content: "api-token: abc123\n", wantWarn: true},
		{name: "documented uppercase keys", content: "ORG: some-org\nAPI-TOKEN: abc123\n", wantWarn: true},
		// A kosli.yml in a repository root is far more often a flow template,
		// which was never loaded as CLI config, so it must stay silent.
		{name: "flow template shape", content: "trail:\n  artifacts:\n    - name: nginx\n", wantWarn: false},
		{name: "unparsable file", content: "\tnot: [valid\n", wantWarn: false},
	}
	for _, tc := range cases {
		suite.Run(tc.name, func() {
			defer func() { global = new(GlobalOpts) }()
			suite.stubHomeConfig()
			suite.chdirWithConfig("kosli.yml", tc.content)

			_, _, _, stderr, err := executeCommandC("version")

			suite.Require().NoError(err)
			if tc.wantWarn {
				suite.Contains(stderr, "kosli.yml")
				suite.Contains(stderr, "no longer loaded automatically")
				suite.Contains(stderr, "--config-file kosli.yml",
					"the warning must name the fix, not just the problem")
			} else {
				suite.NotContains(stderr, "no longer loaded automatically")
			}
		})
	}
}

func TestWorkingDirConfigTestSuite(t *testing.T) {
	suite.Run(t, new(WorkingDirConfigTestSuite))
}

// TestConfigCommandFailsWithoutHomeDirectory pins the other side of an empty
// default config path: `kosli config` must say so rather than silently writing
// a config file into the current working directory.
func (suite *WorkingDirConfigTestSuite) TestConfigCommandFailsWithoutHomeDirectory() {
	mockConfigGetter := new(MockConfigGetter)
	mockConfigGetter.Mock.On("defaultConfigFilePath").Return("")
	defaultConfigFilePathFunc = mockConfigGetter.defaultConfigFilePath
	suite.T().Chdir(suite.T().TempDir())

	_, _, _, _, err := executeCommandC("config --org some-org")

	suite.Require().Error(err)
	suite.Contains(err.Error(), "Could not determine your home directory")
	_, statErr := os.Stat(defaultConfigFilename)
	suite.Require().Error(statErr, "no config file may be written into the working directory")
}
