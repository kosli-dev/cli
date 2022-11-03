package main

import (
	"fmt"
	"github.com/kosli-dev/cli/cmd/kosli/test_support"
	"testing"

	"github.com/kosli-dev/cli/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EnvironmentReportDockerTestSuite struct {
	suite.Suite
	CreatedContainerIDs []string
}

func (suite *EnvironmentReportDockerTestSuite) SetupSuite() {
	err := utils.PullDockerImage(test_support.ImageName)
	require.NoError(suite.T(), err, "pulling the docker image should pass")
}

func (suite *EnvironmentReportDockerTestSuite) TearDownSuite() {
	for _, id := range suite.CreatedContainerIDs {
		err := utils.RemoveDockerContainer(id)
		require.NoError(suite.T(), err, fmt.Sprintf("RemoveDockerContainer: %s", id))
	}
}

func (suite *EnvironmentReportDockerTestSuite) TestCreateDockerArtifactsData() {
	for _, t := range []struct {
		name           string
		imageName      string
		expectedSha256 string
	}{
		{
			name:           "DockerArtifactsData contains the right image digest",
			imageName:      "library/alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
			expectedSha256: "e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
		},
	} {
		suite.Run(t.name, func() {
			err := utils.PullDockerImage(t.imageName)
			require.NoError(suite.T(), err, fmt.Sprintf("PullDockerImage: %s", t.imageName))

			containerID, err := utils.RunDockerContainer(t.imageName)

			require.NoError(suite.T(), err, fmt.Sprintf("RunDockerContainer for %s", t.imageName))
			suite.CreatedContainerIDs = append(suite.CreatedContainerIDs, containerID)

			// wait for container
			assert.Contains(suite.T(), suite.containerDigests(), t.expectedSha256)
		})
	}
}

func (suite *EnvironmentReportDockerTestSuite) containerDigests() []string {
	data, err := CreateDockerArtifactsData()
	require.NoError(suite.T(), err, "TestCreateDockerArtifactsData: CreateDockerArtifactsData")

	actualDigests := []string{}
	for _, item := range data {
		for _, digest := range item.Digests {
			actualDigests = append(actualDigests, digest)
		}
	}
	return actualDigests
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEnvironmentReportDockerTestSuite(t *testing.T) {
	suite.Run(t, new(EnvironmentReportDockerTestSuite))
}
