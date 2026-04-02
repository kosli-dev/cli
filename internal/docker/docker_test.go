package docker

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DockerTestSuite struct {
	suite.Suite
	testImageName string
}

func (suite *DockerTestSuite) SetupSuite() {
	suite.testImageName = "library/alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5"
}

func (suite *DockerTestSuite) SetupTest() {
	err := PullDockerImage(suite.testImageName)
	require.NoError(suite.T(), err)
}

func (suite *DockerTestSuite) TestPullDockerImage() {
	tests := []struct {
		name      string
		imageName string
		wantErr   bool
	}{
		{
			name:      "pulling a real image works",
			imageName: suite.testImageName,
		},
		{
			name:      "pulling a fictional image fails",
			imageName: "library/non-existing",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := PullDockerImage(tt.imageName)
			if tt.wantErr {
				require.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)
			}
		})
	}
}

func (suite *DockerTestSuite) TestPushDockerImage() {
	tests := []struct {
		name       string
		imageName  string
		tagImageAs string
		wantErr    bool
	}{
		{
			name:       "pushing a real image works",
			imageName:  suite.testImageName,
			tagImageAs: "localhost:5001/alpine:v1",
		},
		{
			name:      "pushing to a repo without permission to push fails",
			imageName: suite.testImageName,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.tagImageAs != "" {
				err := TagDockerImage(tt.imageName, tt.tagImageAs)
				require.NoError(suite.T(), err)
			} else {
				tt.tagImageAs = tt.imageName
			}

			err := PushDockerImage(tt.tagImageAs)
			if tt.wantErr {
				require.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)
			}
		})
	}
}

func (suite *DockerTestSuite) TestTagDockerImage() {
	err := TagDockerImage(suite.testImageName, "new-tag")
	require.NoError(suite.T(), err)
}

func (suite *DockerTestSuite) TestRemoveDockerImage() {
	err := RemoveDockerImage(suite.testImageName)
	require.NoError(suite.T(), err)
	err = RemoveDockerImage("non-existing-image")
	require.Error(suite.T(), err)
}

func (suite *DockerTestSuite) TestRunDockerContainer() {
	id, err := RunDockerContainer(suite.testImageName)
	require.NoError(suite.T(), err)
	require.NotEmpty(suite.T(), id)
	err = RemoveDockerContainer(id)
	require.NoError(suite.T(), err)

	id, err = RunDockerContainer("not-existing-image")
	require.Error(suite.T(), err)
	require.Empty(suite.T(), id)
}

func TestDockerTestSuite(t *testing.T) {
	suite.Run(t, new(DockerTestSuite))
}
