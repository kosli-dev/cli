package main

import (
	"bytes"
	"testing"

	"github.com/kosli-dev/cli/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type FingerprintTestSuite struct {
	suite.Suite
}

const imageName = "library/alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5"

func (suite *FingerprintTestSuite) SetupSuite() {
	err := utils.PullDockerImage(imageName)
	require.NoError(suite.T(), err, "pulling the docker image should pass")
}

func (suite *FingerprintTestSuite) TearDownSuite() {
	err := utils.RemoveDockerImage(imageName)
	require.NoError(suite.T(), err, "removing the docker image should pass")
}

func (suite *FingerprintTestSuite) TestRun() {
	for _, t := range []struct {
		name           string
		opts           fingerprintOptions
		args           []string
		expectedSha256 string
		errorExpected  bool
	}{
		{
			name:           "Fingerprinting gives the correct sha256 of a file",
			opts:           fingerprintOptions{artifactType: "file"},
			args:           []string{"testdata/file1"},
			expectedSha256: "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9",
		},
		{
			name:           "Fingerprinting gives the correct sha256 of a directory",
			opts:           fingerprintOptions{artifactType: "dir"},
			args:           []string{"testdata"},
			expectedSha256: "a0b019f292a7b00b24390e0e1f405b03c0e7cc2ac9748481fd8e7bfd9263c74a",
		},
		{
			name:          "Fingerprinting fails if type is directory but the argument is not a dir",
			opts:          fingerprintOptions{artifactType: "dir"},
			args:          []string{"testdata/file1"},
			errorExpected: true,
		},
		{
			name:          "Fingerprinting fails if a directory does not exist",
			opts:          fingerprintOptions{artifactType: "dir"},
			args:          []string{"non-existing"},
			errorExpected: true,
		},
		{
			name:          "Fingerprinting fails if a file does not exist",
			opts:          fingerprintOptions{artifactType: "file"},
			args:          []string{"non-existing"},
			errorExpected: true,
		},
		{
			name:          "Fingerprinting fails if artifact type is not supported",
			opts:          fingerprintOptions{artifactType: "unknown"},
			args:          []string{"testdata"},
			errorExpected: true,
		},
		{
			name:           "Fingerprinting an available docker image works",
			opts:           fingerprintOptions{artifactType: "docker"},
			args:           []string{"library/alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5"},
			expectedSha256: "e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
		},
		{
			name:          "Fingerprinting a non-available docker image fails",
			opts:          fingerprintOptions{artifactType: "docker"},
			args:          []string{"library/merkely"},
			errorExpected: true,
		},
	} {
		suite.Run(t.name, func() {
			var out bytes.Buffer
			err := t.opts.run(t.args, &out)
			if t.errorExpected {
				require.Error(suite.T(), err, "Expected errors but got none")
			} else {
				assert.Equalf(suite.T(), t.expectedSha256, out.String(), "TestCmdRun: want %s, got %s", t.expectedSha256, out.String())
			}
		})
	}
}

func (suite *FingerprintTestSuite) TestFingerprintCmd() {
	tests := []cmdTestCase{
		{
			name:   "file fingerprint",
			cmd:    "fingerprint --artifact-type file testdata/file1",
			golden: "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9",
		},
		{
			name:   "dir fingerprint",
			cmd:    "fingerprint --artifact-type dir testdata",
			golden: "a0b019f292a7b00b24390e0e1f405b03c0e7cc2ac9748481fd8e7bfd9263c74a",
		},
		{
			name:   "docker fingerprint",
			cmd:    "fingerprint --artifact-type docker alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
			golden: "e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
		},
		{
			name:      "non-existing file fingerprint",
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
			name:      "setting registry flags with non-docker artifact-type casues an error",
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
