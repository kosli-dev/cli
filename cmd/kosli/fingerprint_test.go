package main

import (
	"testing"

	"github.com/kosli-dev/cli/internal/docker"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type FingerprintTestSuite struct {
	suite.Suite
	imageName string
}

func (suite *FingerprintTestSuite) SetupSuite() {
	suite.imageName = "library/alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5"
	err := docker.PullDockerImage(suite.imageName)
	require.NoError(suite.T(), err)
}

func (suite *FingerprintTestSuite) TestFingerprintCmd() {
	tests := []cmdTestCase{
		{
			name:   "file fingerprint",
			cmd:    "fingerprint --artifact-type file testdata/file1",
			golden: "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9\n",
		},
		{
			name:   "dir fingerprint",
			cmd:    "fingerprint --artifact-type dir testdata/folder1",
			golden: "c43808cb04c6e66c4c6fc1f972dd67c3b9b71c81e0a0c78730da3699922d17be\n",
		},
		{
			name:   "dir fingerprint with exclude",
			cmd:    "fingerprint --artifact-type dir testdata/folder1 -x folder2",
			golden: "773fd3300860454a2b065c5912c03008adb11e6a6dcf7c1c64c094ceab8f430a\n",
		},
		{
			name:   "dir fingerprint with ignore file",
			cmd:    "fingerprint --artifact-type dir testdata/folder1-with-ignore",
			golden: "038897ea5334462098d65125380d58a493671fb3b8bdbbee1e75ec8bd4a65c23\n",
		},
		{
			name:   "dir fingerprint with symbolic links 1",
			cmd:    "fingerprint --artifact-type dir testdata/folder-with-symlinks-1",
			golden: "1c0740265b38509fb9ce7babf00d7eb57e7e08b59dad8be66897c1ebb5b36409\n",
		},
		{
			// The two folders contains the same files and folder, but a/b/c/point-to-dir points
			// different directories
			name:   "dir fingerprint with symbolic links 2",
			cmd:    "fingerprint --artifact-type dir testdata/folder-with-symlinks-2",
			golden: "2a5fe76bc616a97b2ff6f30f46380f1230b98a58df8aec6e96c1fb0e03b41fd9\n",
		},
		{
			name:      "fails if type is directory but the argument is not a dir",
			cmd:       "fingerprint --artifact-type dir testdata/file1",
			wantError: true,
		},
		{
			name:   "docker fingerprint works when the image is available",
			cmd:    "fingerprint --artifact-type docker alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
			golden: "e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5\n",
		},
		{
			name:      "docker fingerprint fails when the image is NOT available",
			cmd:       "fingerprint --artifact-type docker nginx-not-existing",
			wantError: true,
		},
		{
			name:      "non-existing file fingerprint",
			cmd:       "fingerprint --artifact-type file not-existing.txt",
			wantError: true,
		},
		{
			name:      "non-existing dir fingerprint",
			cmd:       "fingerprint --artifact-type unknown testdata",
			wantError: true,
		},
		{
			name:      "fails if artifact type is not supported",
			cmd:       "fingerprint --artifact-type file not-existing.txt",
			wantError: true,
		},
		{
			name:      "missing required --artifact-type flag",
			cmd:       "fingerprint testdata/file1",
			wantError: true,
		},
		{
			name:      "missing required registry flags",
			cmd:       "fingerprint --artifact-type docker --registry-provider dockerhub merkely/change",
			wantError: true,
		},
		{
			name:      "setting registry flags with non-docker artifact-type causes an error",
			cmd:       "fingerprint --artifact-type file --registry-provider dockerhub --registry-username user --registry-password pass merkely/change",
			wantError: true,
		},
	}
	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFingerprintTestSuite(t *testing.T) {
	suite.Run(t, new(FingerprintTestSuite))
}
