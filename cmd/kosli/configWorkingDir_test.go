package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

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
	for _, name := range []string{"kosli.yml", "kosli.yaml", "kosli.json", "kosli.toml", "kosli.properties", "kosli.env"} {
		suite.Run(name, func() {
			defer func() { global = new(GlobalOpts) }()
			suite.stubHomeConfig()
			content := "host: https://attacker.example\n"
			switch filepath.Ext(name) {
			case ".json":
				content = `{"host": "https://attacker.example"}`
			case ".toml":
				content = `host = "https://attacker.example"`
			case ".properties", ".env":
				content = "host=https://attacker.example\n"
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
		// The settings the advisories were about, and the ones a hand-picked
		// key set was most likely to miss.
		{name: "http proxy only", content: "http-proxy: http://proxy:8080\n", wantWarn: true},
		{name: "kubeconfig only", content: "kubeconfig: ./some-kubeconfig.yml\n", wantWarn: true},
		{name: "flow only", content: "flow: some-flow\n", wantWarn: true},
		// trail and artifacts are CLI flags as well as template keys. As a
		// scalar they are config, so they must warn.
		{name: "trail as a scalar", content: "trail: my-trail\n", wantWarn: true},
		{name: "artifacts as a scalar", content: "artifacts: my-artifact\n", wantWarn: true},
		// A kosli.yml in a repository root is far more often a flow template,
		// which was never loaded as CLI config, so it must stay silent.
		{name: "flow template shape", content: "version: 1\ntrail:\n  attestations:\n    - name: pull-request\n", wantWarn: false},
		{name: "flow template without version", content: "trail:\n  artifacts:\n    - name: nginx\n", wantWarn: false},
		{name: "artifacts only template", content: "artifacts:\n  - name: nginx\n", wantWarn: false},
		{name: "unparsable file", content: "\tnot: [valid\n", wantWarn: false},
		{name: "empty file", content: "", wantWarn: false},
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

// TestHomeConfigIsStillLoaded pins the primary load path. Every other test in
// this suite stubs the default at a path that does not exist, so inverting the
// guard around the config read would leave the rest of the suite green.
func (suite *WorkingDirConfigTestSuite) TestHomeConfigIsStillLoaded() {
	path := filepath.Join(suite.T().TempDir(), defaultConfigFilename)
	suite.Require().NoError(os.WriteFile(path, []byte("host: https://home.example\n"), 0600))
	mockConfigGetter := new(MockConfigGetter)
	mockConfigGetter.Mock.On("defaultConfigFilePath").Return(path)
	defaultConfigFilePathFunc = mockConfigGetter.defaultConfigFilePath

	_, _, _, _, err := executeCommandC("version")

	suite.Require().NoError(err)
	suite.Equal("https://home.example", global.Host,
		"the home config file must still be loaded without the user naming it")
}

// TestOversizedWorkingDirConfigIsNotParsed pins that the warning does not hand
// an arbitrarily large repository-controlled file to a parser.
func (suite *WorkingDirConfigTestSuite) TestOversizedWorkingDirConfigIsNotParsed() {
	suite.stubHomeConfig()
	padding := strings.Repeat("# padding\n", 200000)
	suite.chdirWithConfig("kosli.yml", "org: some-org\n"+padding)

	_, _, _, stderr, err := executeCommandC("version")

	suite.Require().NoError(err)
	suite.NotContains(stderr, "no longer loaded automatically",
		"a file past the size ceiling must be skipped rather than parsed")
}

// TestWarnsAboutIgnoredDotEnvConfig covers the two extensions beyond YAML/JSON
// that viper can actually decode. A kosli.env in the working directory was a
// working redirect before this change, so it has to warn.
func (suite *WorkingDirConfigTestSuite) TestWarnsAboutIgnoredDotEnvConfig() {
	for _, name := range []string{"kosli.env", "kosli.dotenv"} {
		suite.Run(name, func() {
			defer func() { global = new(GlobalOpts) }()
			suite.stubHomeConfig()
			suite.chdirWithConfig(name, "host=https://attacker.example\n")

			_, _, _, stderr, err := executeCommandC("version")

			suite.Require().NoError(err)
			suite.Equal(defaultHost, global.Host)
			suite.Contains(stderr, name)
			suite.Contains(stderr, "no longer loaded automatically")
		})
	}
}

// TestUndecodableWorkingDirConfigIsSilent pins that a format viper lists but has
// no decoder for stays quiet. Before this change such a file made every command
// fail with "failed to parse config file", so no pipeline can have depended on
// it and there is nothing to warn about.
func (suite *WorkingDirConfigTestSuite) TestUndecodableWorkingDirConfigIsSilent() {
	suite.stubHomeConfig()
	suite.chdirWithConfig("kosli.properties", "org=some-org\n")

	_, _, _, stderr, err := executeCommandC("version")

	suite.Require().NoError(err, "an undecodable file must no longer fail the command")
	suite.NotContains(stderr, "no longer loaded automatically")
}

// TestNoWarningWhenHomeConfigExists pins that the warning is limited to the
// population that actually lost behaviour. The old default fell back to the
// working directory only when the home config file was absent, so a user who
// has one never loaded the working-directory file, and telling them to pass
// --config-file would replace their home config rather than restore anything.
func (suite *WorkingDirConfigTestSuite) TestNoWarningWhenHomeConfigExists() {
	path := filepath.Join(suite.T().TempDir(), defaultConfigFilename)
	suite.Require().NoError(os.WriteFile(path, []byte("host: https://home.example\n"), 0600))
	mockConfigGetter := new(MockConfigGetter)
	mockConfigGetter.Mock.On("defaultConfigFilePath").Return(path)
	defaultConfigFilePathFunc = mockConfigGetter.defaultConfigFilePath
	suite.chdirWithConfig("kosli.yml", "org: some-org\n")

	_, _, _, stderr, err := executeCommandC("version")

	suite.Require().NoError(err)
	suite.Equal("https://home.example", global.Host)
	suite.NotContains(stderr, "no longer loaded automatically",
		"a user with a home config file never loaded the working-directory file")
}

// TestNonRegularWorkingDirConfigIsNotParsed pins that the size ceiling cannot
// be walked around with a symlink. os.Stat follows one, and a checkout can ship
// kosli.yml -> /dev/zero, which reports IsDir false and Size 0 while an
// unbounded read waits behind it. The run is bounded so that a regression fails
// here instead of hanging the package.
func (suite *WorkingDirConfigTestSuite) TestNonRegularWorkingDirConfigIsNotParsed() {
	if runtime.GOOS == "windows" {
		suite.T().Skip("/dev/zero and os.Symlink are POSIX-only")
	}
	suite.stubHomeConfig()
	dir := suite.T().TempDir()
	suite.Require().NoError(os.Symlink(os.DevNull, filepath.Join(dir, "kosli.json")))
	suite.Require().NoError(os.Symlink("/dev/zero", filepath.Join(dir, "kosli.yml")))
	suite.T().Chdir(dir)

	type result struct {
		stderr string
		err    error
	}
	done := make(chan result, 1)
	go func() {
		_, _, _, stderr, err := executeCommandC("version")
		done <- result{stderr, err}
	}()

	select {
	case got := <-done:
		suite.Require().NoError(got.err)
		suite.NotContains(got.stderr, "no longer loaded automatically")
	case <-time.After(30 * time.Second):
		// FailNow, not Fail: the goroutine is still inside the unbounded read,
		// and letting the test continue would restore the working directory
		// from under a live command.
		suite.FailNow("reading a non-regular config file did not terminate")
	}
}

// TestNoWarningWhenHomeConfigIsNotYaml pins that the gate follows the config
// name the read above uses, not one filename. A home config is loaded from
// ~/.kosli.json just as happily as ~/.kosli.yml, and warning that user would
// tell them to replace a config file that was loaded moments earlier.
func (suite *WorkingDirConfigTestSuite) TestNoWarningWhenHomeConfigIsNotYaml() {
	home := suite.T().TempDir()
	suite.Require().NoError(os.WriteFile(filepath.Join(home, ".kosli.json"),
		[]byte(`{"host": "https://home.example"}`), 0600))
	mockConfigGetter := new(MockConfigGetter)
	mockConfigGetter.Mock.On("defaultConfigFilePath").Return(filepath.Join(home, defaultConfigFilename))
	defaultConfigFilePathFunc = mockConfigGetter.defaultConfigFilePath
	suite.chdirWithConfig("kosli.yml", "org: some-org\n")

	_, _, _, stderr, err := executeCommandC("version")

	suite.Require().NoError(err)
	suite.Equal("https://home.example", global.Host,
		"a home config file is loaded by config name, so .json counts")
	suite.NotContains(stderr, "no longer loaded automatically",
		"this user's home config was loaded, so nothing was lost")
}

// TestWarnsWhenAnotherCommandShadowsTheConfigFileFlag pins that the warning
// follows the Kosli config file flag, not whatever flag of that name the running
// command happens to declare. snapshot k8s registers its own --config-file for
// namespace selectors, so looking the flag up on the command suppressed the
// warning on the very command #6779 was reported against.
func (suite *WorkingDirConfigTestSuite) TestWarnsWhenAnotherCommandShadowsTheConfigFileFlag() {
	suite.stubHomeConfig()
	dir := suite.T().TempDir()
	suite.Require().NoError(os.WriteFile(filepath.Join(dir, "kosli.yml"),
		[]byte("org: some-org\nhost: https://attacker.example\n"), 0600))
	suite.Require().NoError(os.WriteFile(filepath.Join(dir, "k8s-envs.yml"),
		[]byte("environments:\n  - name: prod-env\n    namespaces: [default]\n"), 0600))
	suite.T().Chdir(dir)

	// The warning is emitted in PersistentPreRunE, which does not stop the run,
	// and --api-token DRY_RUN suppresses only the Kosli request, not the cluster
	// read. A kubeconfig that cannot exist stops it before any cluster is
	// reached, on a developer machine with a current context as well as in CI.
	_, _, _, stderr, err := executeCommandC(
		"snapshot k8s --config-file k8s-envs.yml --kubeconfig " +
			filepath.Join(dir, "no-such-kubeconfig") + " --api-token DRY_RUN --org some-org")

	suite.Require().Error(err, "the run must stop at the kubeconfig, never reaching a cluster")
	suite.Contains(stderr, "no longer loaded automatically",
		"a command's own --config-file must not be mistaken for the Kosli config file")
	suite.Contains(stderr, "This command declares its own --config-file",
		"on this command --config-file, -c and KOSLI_CONFIG_FILE all name something else, so none may be offered as the fix")
	suite.NotContains(stderr, "pass --config-file kosli.yml")
	suite.Equal(defaultHost, global.Host)
}

// TestOnlyTheFileViperWouldHaveLoadedIsReported pins that the loop stops where
// viper stopped. viper took the first existing name in SupportedExts order, so
// with a kosli.json template beside a real kosli.yml it loaded the template and
// never read the yml. Skipping ahead to the yml would claim it was loaded when
// it never was, and the remedy would resolve back to the template, since
// --config-file strips the extension and searches the name again.
func (suite *WorkingDirConfigTestSuite) TestOnlyTheFileViperWouldHaveLoadedIsReported() {
	suite.stubHomeConfig()
	dir := suite.T().TempDir()
	suite.Require().NoError(os.WriteFile(filepath.Join(dir, "kosli.json"),
		[]byte(`{"trail": {"artifacts": [{"name": "nginx"}]}}`), 0600))
	suite.Require().NoError(os.WriteFile(filepath.Join(dir, "kosli.yml"),
		[]byte("org: some-org\n"), 0600))
	suite.T().Chdir(dir)

	_, _, _, stderr, err := executeCommandC("version")

	suite.Require().NoError(err)
	suite.NotContains(stderr, "no longer loaded automatically",
		"the file viper loaded was the template, and a template is not a broken pipeline")
	suite.NotContains(stderr, "kosli.yml",
		"kosli.yml was never the loaded file, and --config-file kosli.yml would load the template anyway")
}

func TestWorkingDirConfigTestSuite(t *testing.T) {
	suite.Run(t, new(WorkingDirConfigTestSuite))
}
