package docker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DockerTestSuite struct {
	suite.Suite
	testImageName string
}

func (suite *DockerTestSuite) SetupSuite() {
	suite.testImageName = "library/alpine:latest"
}

func (suite *DockerTestSuite) SetupTest() {
	err := PullDockerImage(suite.testImageName)
	require.NoError(suite.T(), err)
}

func (suite *DockerTestSuite) TestNewDockerClientNegotiatesAPIVersion() {
	cli, err := newDockerClient()
	suite.Require().NoError(err)

	// Ping triggers version negotiation. If WithAPIVersionNegotiation() is
	// missing and the SDK default exceeds the daemon's max API version,
	// this call will fail with "client version X is too new".
	ping, err := cli.Ping(context.Background())
	suite.Require().NoError(err,
		"Docker client should be able to ping the daemon; "+
			"if this fails with 'client version X is too new', "+
			"WithAPIVersionNegotiation() may be missing from newDockerClient()")

	// After negotiation the client version must not exceed the server's max.
	suite.Assert().NotEmpty(ping.APIVersion,
		"server should report its API version in the ping response")
	suite.Assert().LessOrEqual(cli.ClientVersion(), ping.APIVersion,
		"negotiated client API version should be <= server max API version")
}

func (suite *DockerTestSuite) TestPullDockerImage() {
	for _, t := range []struct {
		name      string
		imageName string
		wantErr   bool
	}{
		{
			name:      "pulling an existing image works",
			imageName: suite.testImageName,
		},
		{
			name:      "pulling a non-existing image fails",
			imageName: "non-existing:latest",
			wantErr:   true,
		},
	} {
		suite.Run(t.name, func() {
			err := PullDockerImage(t.imageName)
			if t.wantErr {
				require.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)
			}
		})
	}
}

func (suite *DockerTestSuite) TestPushDockerImage() {
	for _, t := range []struct {
		name      string
		imageName string
		wantErr   bool
	}{
		{
			name:      "pushing an existing image works",
			imageName: "localhost:5001/alpine:latest",
		},
		{
			name:      "pushing a non-existing image fails",
			imageName: "non-existing:latest",
			wantErr:   true,
		},
	} {
		suite.Run(t.name, func() {
			if !t.wantErr {
				err := TagDockerImage(suite.testImageName, t.imageName)
				require.NoError(suite.T(), err)
			}

			err := PushDockerImage(t.imageName)
			if t.wantErr {
				require.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)
			}
		})
	}
}

func (suite *DockerTestSuite) TestTagDockerImage() {
	for _, t := range []struct {
		name      string
		imageName string
		wantErr   bool
	}{
		{
			name:      "tagging an existing image works",
			imageName: suite.testImageName,
		},
		{
			name:      "tagging a non-existing image fails",
			imageName: "non-existing:latest",
			wantErr:   true,
		},
	} {
		suite.Run(t.name, func() {
			err := TagDockerImage(t.imageName, "new-tag:latest")
			if t.wantErr {
				require.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)
			}
		})
	}
}

func (suite *DockerTestSuite) TestRemoveDockerImage() {
	for _, t := range []struct {
		name      string
		imageName string
		wantErr   bool
	}{
		{
			name:      "removing an existing image works",
			imageName: suite.testImageName,
		},
		{
			name:      "removing a non-existing image fails",
			imageName: "non-existing:latest",
			wantErr:   true,
		},
	} {
		suite.Run(t.name, func() {
			err := RemoveDockerImage(t.imageName)
			if t.wantErr {
				require.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)
			}
		})
	}
}

func (suite *DockerTestSuite) TestRunAndRemoveDockerContainer() {
	containerID, err := RunDockerContainer(suite.testImageName)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), containerID)

	err = RemoveDockerContainer(containerID)
	require.NoError(suite.T(), err)
}

func TestDockerTestSuite(t *testing.T) {
	suite.Run(t, new(DockerTestSuite))
}
