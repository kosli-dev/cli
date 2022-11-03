package main

import (
	"fmt"
	"strings"
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

func (suite *EnvironmentReportDockerTestSuite) TearDownSuite() {
	for _, id := range suite.CreatedContainerIDs {
		err := utils.RemoveDockerContainer(id)
		require.NoError(suite.T(), err, "removing the docker container should pass")
	}
}

func (suite *EnvironmentReportDockerTestSuite) TestCreateDockerArtifactsData() {
	type want struct {
		sha256      string
		expectError bool
	}
	for _, t := range []struct {
		name      string
		imageName string
		want      want
	}{
		{
			name:      "DockerArtifactsData contains the right image digest",
			imageName: "library/alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
			want: want{
				sha256: "e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
			},
		},
	} {
		suite.Run(t.name, func() {
			containerID, err := utils.RunDockerContainer(t.imageName)
			require.NoError(suite.T(), err, "TestCreateDockerArtifactsData: running a container should pass")
			suite.CreatedContainerIDs = append(suite.CreatedContainerIDs, containerID)

			data, err := CreateDockerArtifactsData()
			if t.want.expectError {
				require.Error(suite.T(), err, "TestCreateDockerArtifactsData: error expected but did not happen")
			} else {
				require.NoError(suite.T(), err, "TestCreateDockerArtifactsData: error happened but it is not expected")
			}

			for _, item := range data {
				name, _, _ := strings.Cut(t.imageName, "@sha256:")
				if v, ok := item.Digests[name]; ok {
					assert.Equal(suite.T(), t.want.sha256, v, fmt.Sprintf("TestCreateDockerArtifactsData: want %s -- got %s", t.want.sha256, v))
				}
			}
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEnvironmentReportDockerTestSuite(t *testing.T) {
	suite.Run(t, new(EnvironmentReportDockerTestSuite))
}
