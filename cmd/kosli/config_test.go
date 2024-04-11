package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/zalando/go-keyring"
)

// MockConfigGetter is a mock implementation of the ConfigGetter interface
type MockConfigGetter struct {
	mock.Mock
}

// defaultConfigFilePath is a method that satisfies the ConfigGetter interface
func (m *MockConfigGetter) defaultConfigFilePath() string {
	args := m.Called()
	return args.String(0)
}

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ConfigCommandTestSuite struct {
	suite.Suite
	tmpConfigFilePath string
}

func (suite *ConfigCommandTestSuite) SetupTest() {
	dir, err := os.MkdirTemp("", "tmp-config-file")
	require.NoError(suite.T(), err)
	suite.tmpConfigFilePath = filepath.Join(dir, defaultConfigFilename)
}

func (suite *ConfigCommandTestSuite) TearDownTest() {
	err := os.RemoveAll(suite.tmpConfigFilePath)
	require.NoError(suite.T(), err)
	defaultConfigFilePathFunc = (&RealConfigGetter{}).defaultConfigFilePath
	global = new(GlobalOpts)
}

func (suite *ConfigCommandTestSuite) TestConfigCmd() {
	// Create a new instance of the mock
	mockConfigGetter := new(MockConfigGetter)
	// Set up expectations
	mockConfigGetter.On("defaultConfigFilePath").Return(suite.tmpConfigFilePath)
	// Replace the original function with the mock
	defaultConfigFilePathFunc = mockConfigGetter.defaultConfigFilePath

	// mock keyring credentials store
	keyring.MockInit()

	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "cannot use --config-file with config command",
			cmd:       "config --config-file xyz",
			golden:    "Error: cannot use --config-file with config command\n",
		},
		{
			name:   "can configure org",
			cmd:    "config --org foo",
			golden: fmt.Sprintf("default config file [%s] updated successfully.\n", defaultConfigFilePathFunc()),
		},
		{
			name:   "can configure host",
			cmd:    "config --host https://foo.com",
			golden: fmt.Sprintf("default config file [%s] updated successfully.\n", defaultConfigFilePathFunc()),
		},
		{
			name:   "can configure max api retries",
			cmd:    "config --max-api-retries 1",
			golden: fmt.Sprintf("default config file [%s] updated successfully.\n", defaultConfigFilePathFunc()),
		},
		{
			name:   "can configure api token",
			cmd:    "config --api-token top-secret",
			golden: fmt.Sprintf("default config file [%s] updated successfully.\n", defaultConfigFilePathFunc()),
		},
		{
			name:   "can configure a non-global flag",
			cmd:    "config --set flow=bar",
			golden: fmt.Sprintf("default config file [%s] updated successfully.\n", defaultConfigFilePathFunc()),
		},
		{
			name:   "can unset a configured flag",
			cmd:    "config --unset flow",
			golden: fmt.Sprintf("default config file [%s] updated successfully.\n", defaultConfigFilePathFunc()),
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestConfigCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigCommandTestSuite))
}
