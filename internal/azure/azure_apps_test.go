package azure

import (
	"os"
	"testing"

	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// All methods that begin with "Test" are run as tests within a
// suite.
type AzureAppsTestSuite struct {
	suite.Suite
	staticCreds   AzureStaticCredentials
	defaultClient *AzureClient
}

func (suite *AzureAppsTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.Suite.T(), []string{
		"INTEGRATION_TEST_AZURE_CLIENT_SECRET",
		"INTEGRATION_TEST_AZURE_CLIENT_ID",
	})
	suite.staticCreds = AzureStaticCredentials{
		TenantId:          "e52b5fba-43c2-4eaf-91c1-579dc6fae771",
		ClientId:          os.Getenv("INTEGRATION_TEST_AZURE_CLIENT_ID"),
		ClientSecret:      os.Getenv("INTEGRATION_TEST_AZURE_CLIENT_SECRET"),
		SubscriptionId:    "96cdee58-1fa8-419d-a65a-7233b3465632",
		ResourceGroupName: "EnvironmentReportingExperiment",
	}
	var err error
	suite.defaultClient, err = suite.staticCreds.NewAzureClient()
	require.NoError(suite.Suite.T(), err)
}

// func (suite *AzureAppsTestSuite) TestCanDownloadAppZip() {
// 	token, err := suite.defaultClient.getBearerToken()
// 	require.NoError(suite.Suite.T(), err)
// 	require.NotEmpty(suite.Suite.T(), token)

// 	appName := "kosli-dev-WaveApp"
// 	tmpDir, err := os.MkdirTemp("", "*")
// 	require.NoError(suite.Suite.T(), err)
// 	defer os.RemoveAll(tmpDir)
// 	dest := filepath.Join(tmpDir, appName+".zip")
// 	err = downloadAppPackage(appName, token, dest)
// 	require.NoError(suite.Suite.T(), err)

// 	// check download file exists
// 	_, err = os.Stat(dest)
// 	require.False(suite.Suite.T(), os.IsNotExist(err))

// 	// check downloaded file is a valid zip
// 	r, err := zip.OpenReader(dest)
// 	require.NoError(suite.Suite.T(), err)
// 	defer r.Close()

// }

// func (suite *AzureAppsTestSuite) TestFingerprintZipService() {
// 	appName := "kosli-dev-WaveApp"
// 	appKind := "app"
// 	appData, err := suite.defaultClient.fingerprintZipService(&armappservice.Site{Name: &appName, Kind: &appKind}, logger.NewStandardLogger())
// 	require.NoError(suite.Suite.T(), err)

// 	digest := appData.Digests[appName]
// 	sha256Regex := regexp.MustCompile(`^[a-f0-9]{64}$`)
// 	require.True(suite.Suite.T(), sha256Regex.MatchString(digest))

// 	require.EqualValues(suite.Suite.T(), appData, AppData{AppName: appName,
// 		AppKind:       appKind,
// 		DigestsSource: "kosli-cli",
// 		StartedAt:     0,
// 		Digests: map[string]string{
// 			appName: digest,
// 		},
// 	})

// 	unknownAppName := "unknown"
// 	_, err = suite.defaultClient.fingerprintZipService(&armappservice.Site{Name: &unknownAppName}, logger.NewStandardLogger())
// 	require.Error(suite.Suite.T(), err)
// }

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAzureAppsTestSuite(t *testing.T) {
	suite.Run(t, new(AzureAppsTestSuite))
}
