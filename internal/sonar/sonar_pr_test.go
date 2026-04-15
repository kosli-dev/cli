package sonar

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func newTestLogger() *log.Logger {
	return log.NewLogger(io.Discard, io.Discard, false)
}

// SonarPRTestSuite tests that PR analyses can be retrieved from SonarCloud.
// It uses a real PR scan (PR 359 on cyber-dojo_differ) to reproduce the bug
// where GetProjectAnalysisFromAnalysisID fails because project_analyses/search
// does not return PR analyses.
//
// Requires KOSLI_SONAR_API_TOKEN with access to the cyber-dojo org.
type SonarPRTestSuite struct {
	suite.Suite
	tokenHeader string
	httpClient  *http.Client
}

func (suite *SonarPRTestSuite) SetupSuite() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_SONAR_API_TOKEN"})
	suite.httpClient = &http.Client{}
	suite.tokenHeader = fmt.Sprintf("Bearer %s", os.Getenv("KOSLI_SONAR_API_TOKEN"))
}

// TestGetSonarResults_PRScan tests that GetSonarResults succeeds for a PR scan.
// The CE task for PR 359 on cyber-dojo_differ (task ID: AZ2Qge89T7Y829rQbv87)
// returns analysisId "c089abaa-cf80-4d26-93eb-75898234ea02" and pullRequest "359".
// Currently fails because project_analyses/search does not return PR analyses
// on SonarCloud, and the code has no alternative path for PR scans.
func (suite *SonarPRTestSuite) TestGetSonarResults_PRScan() {
	t := suite.T()

	sc := NewSonarConfig(
		os.Getenv("KOSLI_SONAR_API_TOKEN"),
		"../../cmd/kosli/testdata/sonar/sonarcloud/.scannerwork-pr359",
		"", // ceTaskUrl will be read from report-task.txt
		"", // projectKey
		"", // serverURL
		"", // revision
		30, // maxWait
	)

	results, err := sc.GetSonarResults(newTestLogger())
	require.NoError(t, err)
	require.NotNil(t, results)
	require.NotEmpty(t, results.Revision)
	require.NotEmpty(t, results.AnalaysedAt)
	require.Equal(t, "OK", results.QualityGate.Status)
}

func TestSonarPRTestSuite(t *testing.T) {
	suite.Run(t, new(SonarPRTestSuite))
}
