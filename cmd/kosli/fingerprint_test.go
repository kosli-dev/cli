package main

import (
	"testing"

	"github.com/kosli-dev/cli/cmd/kosli/test_support"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type FingerprintTestSuite struct {
	suite.Suite
}

func (suite *FingerprintTestSuite) SetupSuite() {
	test_support.PullExampleImage(suite.T())
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
			cmd:    "fingerprint --artifact-type dir testdata",
			golden: "a0b019f292a7b00b24390e0e1f405b03c0e7cc2ac9748481fd8e7bfd9263c74a\n",
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
