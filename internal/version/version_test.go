package version

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type VersionTestSuite struct {
	suite.Suite
}

// reset the variables before each test
func (suite *VersionTestSuite) SetupTest() {
	version = "main"
	metadata = ""
	gitCommit = ""
	gitTreeState = ""
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *VersionTestSuite) TestGetVersion() {
	type args struct {
		metadata string
		version  string
	}
	for _, t := range []struct {
		name string
		args args
		want string
	}{
		{
			name: "version is main when metadata is empty.",
			args: args{
				metadata: "",
			},
			want: "main",
		},
		{
			name: "version is suffixed with metadat when metadata is not empty.",
			args: args{
				metadata: "bla",
			},
			want: "main+bla",
		},
		{
			name: "default version is overwritten when provided and there is metadata.",
			args: args{
				metadata: "rc",
				version:  "v1.2.3",
			},
			want: "v1.2.3+rc",
		},
		{
			name: "default version is overwritten when provided.",
			args: args{
				version: "v1.2.3",
			},
			want: "v1.2.3",
		},
	} {
		suite.Run(t.name, func() {
			metadata = t.args.metadata
			if t.args.version != "" {
				version = t.args.version
			}

			actual := GetVersion()
			assert.Equal(suite.T(), t.want, actual, fmt.Sprintf("TestGetVersion: %s , got: %v -- want: %v", t.name, actual, t.want))
		})
	}
}

func (suite *VersionTestSuite) TestGet() {
	metadata = "unreleased"
	version = "v1.2.3-test"
	gitCommit = "1234fakesha1"
	gitTreeState = "dirty"

	expected := BuildInfo{
		Version:      version + "+" + metadata,
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		GoVersion:    runtime.Version(),
	}
	actual := Get()
	assert.Equal(suite.T(), expected, actual, fmt.Sprintf("build info should match, got: %v -- expected: %v", actual, expected))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestVersionTestSuite(t *testing.T) {
	suite.Run(t, new(VersionTestSuite))
}
