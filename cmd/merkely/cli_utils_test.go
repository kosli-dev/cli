package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type CliUtilsTestSuite struct {
	suite.Suite
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *CliUtilsTestSuite) TestWhichCI() {
	for _, t := range []struct {
		name    string
		envVars map[string]string
		want    string
	}{
		{
			name:    "Github actions is detected.",
			envVars: map[string]string{"GITHUB_RUN_NUMBER": "50"},
			want:    github,
		},
		{
			name:    "Bitbucket actions is detected.",
			envVars: map[string]string{"BITBUCKET_BUILD_NUMBER": "50"},
			want:    bitbucket,
		},
		{
			name:    "Teamcity actions is detected.",
			envVars: map[string]string{"TEAMCITY_VERSION": "50"},
			want:    teamcity,
		},
		{
			name:    "No env vars returns unknown",
			envVars: map[string]string{},
			want:    unknown,
		},
	} {
		suite.Run(t.name, func() {
			suite.setEnvVars(t.envVars)
			actual := WhichCI()
			// clean up
			suite.unsetEnvVars(t.envVars)
			assert.Equal(suite.T(), t.want, actual, fmt.Sprintf("TestWhichCI: %s , got: %v -- want: %v", t.name, actual, t.want))
		})
	}
}

func (suite *CliUtilsTestSuite) TestDefaultValue() {
	type args struct {
		ci      string
		flag    string
		envVars map[string]string
	}
	for _, t := range []struct {
		name string
		args args
		want string
	}{
		{
			name: "Lookup an existing default for Github.",
			args: args{
				ci:      github,
				flag:    "git-commit",
				envVars: map[string]string{"GITHUB_SHA": "some-sha"},
			},
			want: "some-sha",
		},
		{
			name: "Lookup an existing default for Bitbucket.",
			args: args{
				ci:   bitbucket,
				flag: "commit-url",
				envVars: map[string]string{
					"BITBUCKET_WORKSPACE": "example_space",
					"BITBUCKET_REPO_SLUG": "example_slug",
					"BITBUCKET_COMMIT":    "example_commit",
				},
			},
			want: "https://bitbucket.org/example_space/example_slug/commits/example_commit",
		},
		{
			name: "Lookup an existing default for Teamcity.",
			args: args{
				ci:   teamcity,
				flag: "git-commit",
				envVars: map[string]string{
					"BUILD_VCS_NUMBER": "example_commit",
				},
			},
			want: "example_commit",
		},
		{
			name: "Lookup a non-existing default for Github.",
			args: args{
				ci:      github,
				flag:    "non-existing",
				envVars: map[string]string{},
			},
			want: "",
		},
		{
			name: "Lookup a default for unknown ci.",
			args: args{
				ci:      unknown,
				flag:    "non-existing",
				envVars: map[string]string{},
			},
			want: "",
		},
	} {
		suite.Run(t.name, func() {
			suite.setEnvVars(t.args.envVars)
			actual := DefaultValue(t.args.ci, t.args.flag)
			// clean up
			suite.unsetEnvVars(t.args.envVars)
			assert.Equal(suite.T(), t.want, actual, fmt.Sprintf("TestDefaultValue: %s , got: %v -- want: %v", t.name, actual, t.want))
		})
	}
}

func (suite *CliUtilsTestSuite) TestRequireGlobalFlags() {
	type args struct {
		global *GlobalOpts
		fields []string
	}
	for _, t := range []struct {
		name        string
		args        args
		expectError bool
	}{
		{
			name: "Required fields are set.",
			args: args{
				global: &GlobalOpts{
					ApiToken: "secret",
					Owner:    "test",
				},
				fields: []string{"ApiToken", "Owner"},
			},
			expectError: false,
		},
		{
			name: "Required fields are not set.",
			args: args{
				global: &GlobalOpts{
					Owner: "test",
				},
				fields: []string{"ApiToken", "Owner"},
			},
			expectError: true,
		},
	} {
		suite.Run(t.name, func() {
			err := RequireGlobalFlags(t.args.global, t.args.fields)
			if t.expectError {
				require.Errorf(suite.T(), err, "TestRequireGlobalFlags: error was expected but got none.")
			} else {
				require.NoErrorf(suite.T(), err, "TestRequireGlobalFlags: got an error but was not expecting one:  %v", err)
			}
		})
	}
}

func (suite *CliUtilsTestSuite) TestGetFlagFromVarName() {
	for _, t := range []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "camel case var name is converted to a cmd flag name.",
			input: "ApiToken",
			want:  "--api-token",
		},
		{
			name:  "empty string returns empty string.",
			input: "",
			want:  "",
		},
		{
			name:  "one letter input returns --<letter>.",
			input: "A",
			want:  "--a",
		},
	} {
		suite.Run(t.name, func() {
			actual := GetFlagFromVarName(t.input)
			assert.Equal(suite.T(), t.want, actual, fmt.Sprintf("TestGetFlagFromVarName: %s , got: %v -- want: %v", t.name, actual, t.want))
		})
	}
}

func (suite *CliUtilsTestSuite) TestNoArgs() {
	type args struct {
		cmd     *cobra.Command
		cmdArgs []string
	}
	for _, t := range []struct {
		name        string
		args        args
		expectError bool
	}{
		{
			name: "providing no cmd args does not throw an error.",
			args: args{
				cmd:     &cobra.Command{},
				cmdArgs: []string{},
			},
			expectError: false,
		},
		{
			name: "providing cmd args throws an error.",
			args: args{
				cmd:     &cobra.Command{},
				cmdArgs: []string{"arg1"},
			},
			expectError: true,
		},
	} {
		suite.Run(t.name, func() {
			err := NoArgs(t.args.cmd, t.args.cmdArgs)
			if t.expectError {
				require.Errorf(suite.T(), err, "TestNoArgs: error was expected but got none.")
			} else {
				require.NoErrorf(suite.T(), err, "TestNoArgs: got an error but was not expecting one:  %v", err)
			}
		})
	}
}

func (suite *CliUtilsTestSuite) TestGetCIDefaultsTemplates() {
	text := GetCIDefaultsTemplates(supportedCIs, []string{"git-commit"})
	require.NotEmpty(suite.T(), text, "TestGetCIDefaultsTemplates: returned string should not be empty")
}

func (suite *CliUtilsTestSuite) TestGetSha256Digest() {
	type args struct {
		artifactType string
		artifactName string
	}
	for _, t := range []struct {
		name        string
		args        args
		expectError bool
		want        string
	}{
		{
			name: "not supported artifact type returns an error.",
			args: args{
				artifactType: "unknown",
				artifactName: "",
			},
			expectError: true,
		},
		{
			name: "non-existing dir returns an error.",
			args: args{
				artifactType: "dir",
				artifactName: "non-existing",
			},
			expectError: true,
		},
		{
			name: "non-existing file returns an error.",
			args: args{
				artifactType: "file",
				artifactName: "non-existing.txt",
			},
			expectError: true,
		},
		{
			name: "non-existing docker image returns an error.",
			args: args{
				artifactType: "docker",
				artifactName: "registry/non-existing",
			},
			expectError: true,
		},
	} {
		suite.Run(t.name, func() {
			fingerprint, err := GetSha256Digest(t.args.artifactType, t.args.artifactName)
			if t.expectError {
				require.Errorf(suite.T(), err, "TestGetSha256Digest: error was expected but got none.")
			} else {
				require.NoErrorf(suite.T(), err, "TestGetSha256Digest: got an error but was not expecting one:  %v", err)
				assert.Equal(suite.T(), t.want, fingerprint, fmt.Sprintf("TestGetSha256Digest: %s , got: %v -- want: %v", t.name, fingerprint, t.want))
			}
		})
	}
}

func (suite *CliUtilsTestSuite) TestLoadUserData() {
	type args struct {
		filename string
		content  string
		create   bool
	}
	for _, t := range []struct {
		name         string
		args         args
		expectError  bool
		wantedLength int
	}{
		{
			name: "a valid JSON file.",
			args: args{
				filename: "test1.json",
				content:  "{\"key\": \"value\"}",
				create:   true,
			},
			expectError:  false,
			wantedLength: 1,
		},
		{
			name: "a not valid JSON file.",
			args: args{
				filename: "test2.json",
				content:  "No json",
				create:   true,
			},
			expectError:  true,
			wantedLength: 0,
		},
		{
			name: "a non existing file returns an error.",
			args: args{
				filename: "test2.json",
				content:  "No json",
				create:   false,
			},
			expectError:  true,
			wantedLength: 0,
		},
	} {
		suite.Run(t.name, func() {
			tmpDir, err := ioutil.TempDir("", "testDir")
			require.NoError(suite.T(), err, "error creating a temporary test directory")
			defer os.RemoveAll(tmpDir)

			if t.args.create {
				testFile, err := os.Create(filepath.Join(tmpDir, t.args.filename))
				require.NoErrorf(suite.T(), err, "error creating test file %s", t.args.filename)

				_, err = testFile.Write([]byte(t.args.content))
				require.NoErrorf(suite.T(), err, "error writing content to test file %s", t.args.filename)
			}

			data, err := LoadUserData(filepath.Join(tmpDir, t.args.filename))
			if t.expectError {
				require.Errorf(suite.T(), err, "TestLoadUserData: error was expected but got none.")
			} else {
				require.NoErrorf(suite.T(), err, "TestLoadUserData: got an error but was not expecting one:  %v", err)
				require.Len(suite.T(), data, t.wantedLength, fmt.Sprintf("TestLoadUserData: %s , got: %v -- want: %v", t.name, len(data), t.wantedLength))
			}
		})
	}
}

func (suite *CliUtilsTestSuite) TestValidateArtifactArg() {
	for _, t := range []struct {
		name         string
		args         []string
		artifactType string
		inputSha256  string
		expectError  bool
	}{
		{
			name:         "two args are not allowed",
			args:         []string{"arg1", "arg2"},
			artifactType: "dir",
			expectError:  true,
		},
		{
			name:         "no args are not allowed",
			args:         []string{},
			artifactType: "dir",
			expectError:  true,
		},
		{
			name:         "empty args is not allowed",
			args:         []string{""},
			artifactType: "dir",
			expectError:  true,
		},
		{
			name:        "missing both artifact type and sha is not allowed",
			args:        []string{"arg1"},
			expectError: true,
		},
		{
			name:        "invalid sha256 is not allowed",
			args:        []string{"arg1"},
			inputSha256: "12345",
			expectError: true,
		},
		{
			name:         "happy case with artifact type",
			args:         []string{"arg1"},
			artifactType: "dir",
			expectError:  false,
		},
		{
			name:        "happy case with artifact sha",
			args:        []string{"arg1"},
			inputSha256: "8b4fd747df6882b897aa514af7b40571a7508cc78a8d48ae2c12f9f4bcb1598f",
			expectError: false,
		},
	} {
		suite.Run(t.name, func() {
			err := ValidateArtifactArg(t.args, t.artifactType, t.inputSha256)
			if t.expectError {
				require.Errorf(suite.T(), err, "error was expected but got none")
			} else {
				require.NoErrorf(suite.T(), err, "error was NOT expected but got %v", err)
			}
		})
	}
}

// setEnvVars sets env variables
func (suite *CliUtilsTestSuite) setEnvVars(envVars map[string]string) {
	for key, value := range envVars {
		err := os.Setenv(key, value)
		require.NoErrorf(suite.T(), err, "error setting env variable %s", key)
	}
}

// unsetEnvVars unsets env variables
func (suite *CliUtilsTestSuite) unsetEnvVars(envVars map[string]string) {
	for key := range envVars {
		err := os.Unsetenv(key)
		require.NoErrorf(suite.T(), err, "error unsetting env variable %s", key)
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCliUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(CliUtilsTestSuite))
}
