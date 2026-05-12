package main

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	"github.com/kosli-dev/cli/internal/digest"
	"github.com/kosli-dev/cli/internal/docker"
	log "github.com/kosli-dev/cli/internal/logger"
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

func TestDockerArtifactsFromContainers(t *testing.T) {
	t.Run("returns digest for a single container when inspect succeeds", func(t *testing.T) {
		containers := []container.Summary{
			{Names: []string{"/c1"}, Image: "alpine", Created: 100},
		}
		getDigest := func(imageID string) (string, error) {
			return "deadbeef", nil
		}

		result, err := dockerArtifactsFromContainers(containers, getDigest, newTestLogger())

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "deadbeef", result[0].Digests["alpine"])
		assert.Equal(t, int64(100), result[0].CreationTimestamp)
	})

	t.Run("skips container when inspect returns errdefs.NotFound, returns the rest", func(t *testing.T) {
		containers := []container.Summary{
			{Names: []string{"/good"}, Image: "alpine", Created: 100},
			{Names: []string{"/orphan"}, Image: "ghost", Created: 200},
		}
		getDigest := func(imageID string) (string, error) {
			if imageID == "ghost" {
				return "", fmt.Errorf("No such image: sha256:abc: %w", cerrdefs.ErrNotFound)
			}
			return "deadbeef", nil
		}
		logBuf := &bytes.Buffer{}

		result, err := dockerArtifactsFromContainers(containers, getDigest, newTestLoggerWithErr(logBuf))

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, "deadbeef", result[0].Digests["alpine"])
		assert.Contains(t, logBuf.String(), "orphan")
	})

	t.Run("skips container when inspect returns ErrRepoDigestUnavailable (preserved)", func(t *testing.T) {
		containers := []container.Summary{
			{Names: []string{"/good"}, Image: "alpine", Created: 100},
			{Names: []string{"/local-only"}, Image: "scratch-built", Created: 200},
		}
		getDigest := func(imageID string) (string, error) {
			if imageID == "scratch-built" {
				return "", digest.ErrRepoDigestUnavailable
			}
			return "deadbeef", nil
		}
		logBuf := &bytes.Buffer{}

		result, err := dockerArtifactsFromContainers(containers, getDigest, newTestLoggerWithInfo(logBuf))

		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Contains(t, logBuf.String(), "local-only")
	})

	t.Run("returns error for any other inspect error (preserved)", func(t *testing.T) {
		containers := []container.Summary{
			{Names: []string{"/c1"}, Image: "alpine", Created: 100},
		}
		boom := errors.New("daemon connection broken")
		getDigest := func(imageID string) (string, error) {
			return "", boom
		}

		result, err := dockerArtifactsFromContainers(containers, getDigest, newTestLogger())

		require.ErrorIs(t, err, boom)
		assert.Empty(t, result)
	})

	t.Run("mixed: skips middle NotFound, returns first and third in order", func(t *testing.T) {
		containers := []container.Summary{
			{Names: []string{"/first"}, Image: "alpine", Created: 100},
			{Names: []string{"/middle"}, Image: "ghost", Created: 200},
			{Names: []string{"/third"}, Image: "busybox", Created: 300},
		}
		getDigest := func(imageID string) (string, error) {
			switch imageID {
			case "alpine":
				return "aaaa", nil
			case "busybox":
				return "bbbb", nil
			default:
				return "", fmt.Errorf("No such image: %w", cerrdefs.ErrNotFound)
			}
		}

		result, err := dockerArtifactsFromContainers(containers, getDigest, newTestLogger())

		require.NoError(t, err)
		require.Len(t, result, 2)
		assert.Equal(t, "aaaa", result[0].Digests["alpine"])
		assert.Equal(t, int64(100), result[0].CreationTimestamp)
		assert.Equal(t, "bbbb", result[1].Digests["busybox"])
		assert.Equal(t, int64(300), result[1].CreationTimestamp)
	})
}

func newTestLogger() *log.Logger {
	return log.NewLogger(&bytes.Buffer{}, &bytes.Buffer{}, false)
}

func newTestLoggerWithErr(errBuf *bytes.Buffer) *log.Logger {
	return log.NewLogger(&bytes.Buffer{}, errBuf, false)
}

func newTestLoggerWithInfo(infoBuf *bytes.Buffer) *log.Logger {
	return log.NewLogger(infoBuf, &bytes.Buffer{}, false)
}
