package main

import (
	"fmt"
	"testing"

	"github.com/kosli-dev/cli/internal/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type SnapshotDockerTestSuite struct {
	suite.Suite
	imageName           string
	createdContainerIDs []string
}

func (suite *SnapshotDockerTestSuite) SetupSuite() {
	suite.imageName = "library/alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5"
	err := docker.PullDockerImage(suite.imageName)
	require.NoError(suite.T(), err)
}

func (suite *SnapshotDockerTestSuite) TearDownSuite() {
	for _, id := range suite.createdContainerIDs {
		err := docker.RemoveDockerContainer(id)
		require.NoError(suite.T(), err, fmt.Sprintf("RemoveDockerContainer: %s", id))
	}
}

func (suite *SnapshotDockerTestSuite) TestCreateDockerArtifactsData() {
	for _, t := range []struct {
		name           string
		imageName      string
		expectedSha256 string
	}{
		{
			name:           "DockerArtifactsData contains the right image digest",
			imageName:      suite.imageName,
			expectedSha256: "e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
		},
	} {
		suite.Run(t.name, func() {
			suite.withRunningContainer(t.imageName)

			assert.Contains(suite.T(), suite.containerDigests(), t.expectedSha256)
		})
	}
}

func (suite *SnapshotDockerTestSuite) withRunningContainer(imageName string) {
	containerID, err := docker.RunDockerContainer(imageName)
	require.NoError(suite.T(), err, fmt.Sprintf("RunDockerContainer for %s", imageName))
	suite.createdContainerIDs = append(suite.createdContainerIDs, containerID)
}

func (suite *SnapshotDockerTestSuite) containerDigests() []string {
	data, err := CreateDockerArtifactsData()
	require.NoError(suite.T(), err, "CreateDockerArtifactsData")

	var actualDigests []string
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
	suite.Run(t, new(SnapshotDockerTestSuite))
}
