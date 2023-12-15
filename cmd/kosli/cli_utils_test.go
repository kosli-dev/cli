package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	log "github.com/kosli-dev/cli/internal/logger"
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
		ci               string
		flag             string
		envVars          map[string]string
		unsetTestsEnvVar bool
	}
	for _, t := range []struct {
		name string
		args args
		want string
	}{
		{
			name: "Lookup an existing default for Github.",
			args: args{
				ci:               github,
				flag:             "git-commit",
				envVars:          map[string]string{"GITHUB_SHA": "some-sha"},
				unsetTestsEnvVar: true,
			},
			want: "some-sha",
		},
		{
			name: "Lookup default org for Github.",
			args: args{
				ci:               github,
				flag:             "org",
				envVars:          map[string]string{"GITHUB_REPOSITORY_OWNER": "cyber-dojo"},
				unsetTestsEnvVar: true,
			},
			want: "cyber-dojo",
		},
		{
			name: "Lookup default repository for Github.",
			args: args{
				ci:               github,
				flag:             "repository",
				envVars:          map[string]string{"GITHUB_REPOSITORY": "cyber-dojo/dashboard"},
				unsetTestsEnvVar: true,
			},
			want: "cyber-dojo/dashboard",
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
				unsetTestsEnvVar: true,
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
				unsetTestsEnvVar: true,
			},
			want: "example_commit",
		},
		{
			name: "Lookup a non-existing default for Github.",
			args: args{
				ci:               github,
				flag:             "non-existing",
				envVars:          map[string]string{},
				unsetTestsEnvVar: true,
			},
			want: "",
		},
		{
			name: "Lookup a default for unknown ci.",
			args: args{
				ci:               unknown,
				flag:             "non-existing",
				envVars:          map[string]string{},
				unsetTestsEnvVar: true,
			},
			want: "",
		},
		{
			name: "Lookup a default when DOCS env var is set returns empty string.",
			args: args{
				ci:               github,
				flag:             "git-commit",
				envVars:          map[string]string{"DOCS": "True", "GITHUB_SHA": "some-sha"},
				unsetTestsEnvVar: true,
			},
			want: "",
		},
		{
			name: "Lookup a default when KOSLI_TESTS env var is set returns empty string.",
			args: args{
				ci:      github,
				flag:    "git-commit",
				envVars: map[string]string{"KOSLI_TESTS": "True", "GITHUB_SHA": "some-sha"},
			},
			want: "",
		},
		{
			name: "Lookup commit-url for CircleCI with a repo from bitbucket returns correct url (with 'commits')",
			args: args{
				ci:               circleci,
				flag:             "commit-url",
				unsetTestsEnvVar: true,
				envVars:          map[string]string{"CIRCLE_REPOSITORY_URL": "git@bitbucket.org:ewelinawilkosz/cli-test.git", "CIRCLE_SHA1": "2492011ef04a9da09d35be706cf6a4c5bc6f1e69"},
			},
			want: "https://bitbucket.org/ewelinawilkosz/cli-test/commits/2492011ef04a9da09d35be706cf6a4c5bc6f1e69",
		},
		{
			name: "Lookup commit-url for CircleCI with a repo that is not from bitbucket returns correct url (with 'commits')",
			args: args{
				ci:               circleci,
				flag:             "commit-url",
				unsetTestsEnvVar: true,
				envVars:          map[string]string{"CIRCLE_REPOSITORY_URL": "git@github.com:cyber-dojo/kosli-environment-reporter.git", "CIRCLE_SHA1": "84d80cd07ef86c1a5afbe69af491e5b3836a3f42"},
			},
			want: "https://github.com/cyber-dojo/kosli-environment-reporter/commit/84d80cd07ef86c1a5afbe69af491e5b3836a3f42",
		},
	} {
		suite.Run(t.name, func() {
			value, testMode := os.LookupEnv("KOSLI_TESTS")
			if t.args.unsetTestsEnvVar && testMode {
				err := os.Unsetenv("KOSLI_TESTS")
				require.NoError(suite.T(), err, "should have unset KOSLI_TESTS env var without error")
			}
			suite.setEnvVars(t.args.envVars)
			actual := DefaultValue(t.args.ci, t.args.flag)
			// clean up any env vars we set from the test case
			suite.unsetEnvVars(t.args.envVars)
			// recover KOSLI_TESTS env variable to its original state before the test
			if testMode {
				os.Setenv("KOSLI_TESTS", value)
			}
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
					Org:      "test",
				},
				fields: []string{"ApiToken", "Org"},
			},
			expectError: false,
		},
		{
			name: "Required fields are not set.",
			args: args{
				global: &GlobalOpts{
					Org: "test",
				},
				fields: []string{"ApiToken", "Org"},
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

func (suite *CliUtilsTestSuite) TestGetCIDefaultsTemplates() {
	text := GetCIDefaultsTemplates(supportedCIs, []string{"git-commit"})
	require.NotEmpty(suite.T(), text, "TestGetCIDefaultsTemplates: returned string should not be empty")
}

func (suite *CliUtilsTestSuite) TestGetSha256Digest() {
	type args struct {
		fingerprintOptions *fingerprintOptions
		artifactName       string
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
				fingerprintOptions: &fingerprintOptions{
					artifactType: "unknown",
				},
				artifactName: "",
			},
			expectError: true,
		},
		{
			name: "non-existing dir returns an error.",
			args: args{
				fingerprintOptions: &fingerprintOptions{
					artifactType: "dir",
				},
				artifactName: "non-existing",
			},
			expectError: true,
		},
		{
			name: "non-existing file returns an error.",
			args: args{
				fingerprintOptions: &fingerprintOptions{
					artifactType: "file",
				},
				artifactName: "non-existing.txt",
			},
			expectError: true,
		},
		{
			name: "non-existing docker image returns an error.",
			args: args{
				fingerprintOptions: &fingerprintOptions{
					artifactType: "docker",
				},
				artifactName: "registry/non-existing",
			},
			expectError: true,
		},
		{
			name: "getting digest from docker registry fails when credentials are invalid",
			args: args{
				fingerprintOptions: &fingerprintOptions{
					artifactType:     "docker",
					registryProvider: "dockerhub",
					registryUsername: "user",
					registryPassword: "pass",
				},
				artifactName: "merkely/change",
			},
			expectError: true,
		},
		{
			name: "getting digest from docker registry fails when provider is not supported",
			args: args{
				fingerprintOptions: &fingerprintOptions{
					artifactType:     "docker",
					registryProvider: "unknown",
					registryUsername: "user",
					registryPassword: "pass",
				},
				artifactName: "merkely/change",
			},
			expectError: true,
		},
	} {
		suite.Run(t.name, func() {
			fingerprint, err := GetSha256Digest(t.args.artifactName, t.args.fingerprintOptions,
				log.NewStandardLogger())
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
		name        string
		args        args
		expectError bool
	}{
		{
			name: "a valid JSON file with an object.",
			args: args{
				filename: "test1.json",
				content:  "{\"key\": \"value\"}",
				create:   true,
			},
			expectError: false,
		},
		{
			name: "a valid JSON file with a list.",
			args: args{
				filename: "test_list.json",
				content:  "[{\"key\": \"value\"}]",
				create:   true,
			},
			expectError: false,
		},
		{
			name: "a not valid JSON file.",
			args: args{
				filename: "test2.json",
				content:  "No json",
				create:   true,
			},
			expectError: true,
		},
		{
			name: "a non existing file returns an error.",
			args: args{
				filename: "test2.json",
				content:  "No json",
				create:   false,
			},
			expectError: true,
		},
	} {
		suite.Run(t.name, func() {
			tmpDir, err := os.MkdirTemp("", "testDir")
			require.NoError(suite.T(), err, "error creating a temporary test directory")
			defer os.RemoveAll(tmpDir)

			if t.args.create {
				testFile, err := os.Create(filepath.Join(tmpDir, t.args.filename))
				require.NoErrorf(suite.T(), err, "error creating test file %s", t.args.filename)

				_, err = testFile.Write([]byte(t.args.content))
				require.NoErrorf(suite.T(), err, "error writing content to test file %s", t.args.filename)
			}

			_, err = LoadJsonData(filepath.Join(tmpDir, t.args.filename))
			if t.expectError {
				require.Errorf(suite.T(), err, "TestLoadUserData: error was expected but got none.")
			} else {
				require.NoErrorf(suite.T(), err, "TestLoadUserData: got an error but was not expecting one:  %v", err)
			}
		})
	}
}

func (suite *CliUtilsTestSuite) TestValidateArtifactArg() {
	for _, t := range []struct {
		name                      string
		args                      []string
		artifactType              string
		inputSha256               string
		alwaysRequireArtifactName bool
		expectError               bool
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
		{
			name:                      "throws an error when sha256 is provided and arg(filename) is not and it is expected",
			args:                      []string{""},
			inputSha256:               "8b4fd747df6882b897aa514af7b40571a7508cc78a8d48ae2c12f9f4bcb1598f",
			expectError:               true,
			alwaysRequireArtifactName: true,
		},
		{
			name:                      "does not throw an error when sha256 is provided and arg(filename) is not and it is NOT expected",
			args:                      []string{""},
			inputSha256:               "8b4fd747df6882b897aa514af7b40571a7508cc78a8d48ae2c12f9f4bcb1598f",
			expectError:               false,
			alwaysRequireArtifactName: false,
		},
		{
			name:                      "arguments with leading space cause a more detailed error message",
			args:                      []string{"/tmp", " "},
			inputSha256:               "8b4fd747df6882b897aa514af7b40571a7508cc78a8d48ae2c12f9f4bcb1598f",
			expectError:               true,
			alwaysRequireArtifactName: true,
		},
	} {
		suite.Run(t.name, func() {
			err := ValidateArtifactArg(t.args, t.artifactType, t.inputSha256, t.alwaysRequireArtifactName)
			if t.expectError {
				require.Errorf(suite.T(), err, "error was expected but got none")
			} else {
				require.NoErrorf(suite.T(), err, "error was NOT expected but got %v", err)
			}
		})
	}
}

func (suite *CliUtilsTestSuite) TestGetRegistryEndpointForProvider() {
	for _, t := range []struct {
		name        string
		provider    string
		want        *registryProviderEndpoints
		expectError bool
	}{
		{
			name:     "github provider returns expected endpoints",
			provider: "github",
			want: &registryProviderEndpoints{
				mainApi: "https://ghcr.io/v2",
				authApi: "https://ghcr.io",
				service: "ghcr.io",
			},
		},
		{
			name:     "dockerhub provider returns expected endpoints",
			provider: "dockerhub",
			want: &registryProviderEndpoints{
				mainApi: "https://registry-1.docker.io/v2",
				authApi: "https://auth.docker.io",
				service: "registry.docker.io",
			},
		},
	} {
		suite.Run(t.name, func() {
			endpoints, err := getRegistryEndpointForProvider(t.provider)
			if t.expectError {
				require.Errorf(suite.T(), err, "error was expected but got none")
			} else {
				require.NoErrorf(suite.T(), err, "error was NOT expected but got %v", err)
				require.Equalf(suite.T(), t.want, endpoints,
					"TestGetRegistryEndpointForProvider: got %v -- want %v", t.want, endpoints)
			}
		})
	}
}

func (suite *CliUtilsTestSuite) TestValidateRegistryFlags() {
	for _, t := range []struct {
		name        string
		options     *fingerprintOptions
		expectError bool
	}{
		{
			name: "registry flags are valid",
			options: &fingerprintOptions{
				artifactType:     "docker",
				registryProvider: "dockerhub",
				registryUsername: "user",
				registryPassword: "pass",
			},
		},
		{
			name: "non-docker type with registry flags set casues an error",
			options: &fingerprintOptions{
				artifactType:     "file",
				registryProvider: "dockerhub",
				registryUsername: "user",
				registryPassword: "pass",
			},
			expectError: true,
		},
		{
			name: "missing username causes an error",
			options: &fingerprintOptions{
				artifactType:     "docker",
				registryProvider: "dockerhub",
				registryPassword: "pass",
			},
			expectError: true,
		},
		{
			name: "missing password causes an error",
			options: &fingerprintOptions{
				artifactType:     "docker",
				registryProvider: "dockerhub",
				registryUsername: "user",
			},
			expectError: true,
		},
		{
			name: "missing provider causes an error 1",
			options: &fingerprintOptions{
				artifactType:     "docker",
				registryUsername: "user",
				registryPassword: "pass",
			},
			expectError: true,
		},
		{
			name: "missing provider causes an error 2",
			options: &fingerprintOptions{
				artifactType:     "docker",
				registryUsername: "user",
			},
			expectError: true,
		},
		{
			name: "missing provider causes an error 3",
			options: &fingerprintOptions{
				artifactType:     "docker",
				registryPassword: "pass",
			},
			expectError: true,
		},
	} {
		suite.Run(t.name, func() {
			err := ValidateRegistryFlags(&cobra.Command{}, t.options)
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

func (suite *CliUtilsTestSuite) TestMuXRequiredFlags() {
	tests := []struct {
		name       string
		flagNames  []string
		atLeastOne bool
		wantErr    bool
		setFlags   int
	}{
		{
			name:      "you can't set both mutually exclusive flags",
			flagNames: []string{"flag1", "flag2"},
			wantErr:   true,
			setFlags:  2,
		},
		{
			name:       "you can set one of TWO mutually exclusive flags when one is strictly required",
			flagNames:  []string{"flag1", "flag2", "flag3"},
			setFlags:   1,
			atLeastOne: true,
		},
		{
			name:       "it is okay to not set any of TWO mutually exclusive flags when atLeastOne is false",
			flagNames:  []string{"flag1", "flag2"},
			setFlags:   0,
			atLeastOne: false,
		},
		{
			name:      "you can set one of THREE mutually exclusive flags",
			flagNames: []string{"flag1", "flag2", "flag3"},
			setFlags:  1,
		},
		{
			name:      "you can't set all mutually exclusive flags",
			flagNames: []string{"flag1", "flag2", "flag3"},
			wantErr:   true,
			setFlags:  3,
		},
		{
			name:      "you can't set more than one of mutually exclusive flags",
			flagNames: []string{"flag1", "flag2", "flag3"},
			wantErr:   true,
			setFlags:  2,
		},
		{
			name:       "an error is returned when none of the mutually exclusive flags are set",
			flagNames:  []string{"flag1", "flag2"},
			wantErr:    true,
			setFlags:   0,
			atLeastOne: true,
		},
	}
	for _, t := range tests {
		suite.Run(t.name, func() {
			cmd := &cobra.Command{}
			var var1, var2, var3 string
			cmd.Flags().StringVar(&var1, "flag1", "", "")
			cmd.Flags().StringVar(&var2, "flag2", "", "")
			cmd.Flags().StringVar(&var3, "flag3", "", "")

			for i := 0; i < t.setFlags; i++ {
				cmd.Flags().Lookup(t.flagNames[i]).Changed = true
			}
			err := MuXRequiredFlags(cmd, t.flagNames, t.atLeastOne)
			if t.wantErr {
				require.Error(suite.T(), err)
			} else {
				require.NoError(suite.T(), err)
			}
		})
	}
}

func (suite *CliUtilsTestSuite) TestFormattedTimestamp() {
	tests := []struct {
		name      string
		timestamp interface{}
		short     bool
		expected  string
		wantErr   bool
	}{
		{
			name:      "can format int64 timestamp",
			timestamp: int64(1679652243),
			short:     true,
			expected:  "Fri, 24 Mar 2023 10:04:03 UTC",
		},
		{
			name:      "can format float64 timestamp",
			timestamp: float64(1679652243),
			short:     true,
			expected:  "Fri, 24 Mar 2023 10:04:03 UTC",
		},
		{
			name:      "can format string timestamp",
			timestamp: "1679652243",
			short:     true,
			expected:  "Fri, 24 Mar 2023 10:04:03 UTC",
		},
		{
			name:      "invalid string format for timestamp fails",
			timestamp: "not-a-timestamp",
			short:     true,
			wantErr:   true,
		},
		{
			name:      "nil value for timestamp returns N/A",
			timestamp: nil,
			short:     true,
			expected:  "N/A",
		},
		{
			name:      "unsupported format returns an error",
			timestamp: true,
			short:     true,
			wantErr:   true,
		},
	}
	for _, t := range tests {
		suite.Run(t.name, func() {
			os.Setenv("KOSLI_TESTS_FORMATTED_TIMESTAMP", "True")
			defer os.Unsetenv("KOSLI_TESTS_FORMATTED_TIMESTAMP")
			ts, err := formattedTimestamp(t.timestamp, t.short)
			require.True(suite.T(), t.wantErr == (err != nil))
			require.Equal(suite.T(), t.expected, ts)
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCliUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(CliUtilsTestSuite))
}

func (suite *CliUtilsTestSuite) TestHandleExpressions() {
	tests := []struct {
		name       string
		expression string
		wantName   string
		wantId     int
		wantErr    bool
	}{
		{
			name:       "valid expression without special characters works",
			expression: "hadron",
			wantName:   "hadron",
			wantId:     -1,
		},
		{
			name:       "valid expression with # works",
			expression: "hadron#12",
			wantName:   "hadron",
			wantId:     12,
		},
		{
			name:       "valid expression with ~ works 1",
			expression: "hadron~1",
			wantName:   "hadron",
			wantId:     -2,
		},
		{
			name:       "valid expression with ~ works 2",
			expression: "hadron~2",
			wantName:   "hadron",
			wantId:     -3,
		},
		{
			name:       "invalid expression causes an error",
			expression: "hadron#abc",
			wantErr:    true,
		},
		{
			name:       "missing flow name with # causes an error",
			expression: "#12",
			wantErr:    true,
		},
		{
			name:       "missing flow name with ~ causes an error",
			expression: "~12",
			wantErr:    true,
		},
		{
			name:       "invalid expression containing ~ and # causes an error",
			expression: "hadron#2~12",
			wantErr:    true,
		},
		{
			name:       "invalid expression containing multiple #s causes an error",
			expression: "hadron#2#3",
			wantErr:    true,
		},
		{
			name:       "invalid expression containing multiple ~s causes an error",
			expression: "hadron~2~1",
			wantErr:    true,
		},
	}
	for _, t := range tests {
		suite.Run(t.name, func() {
			name, id, err := handleExpressions(t.expression)
			require.True(suite.T(), err != nil == t.wantErr)
			require.Equal(suite.T(), t.wantName, name)
			require.Equal(suite.T(), t.wantId, id)
		})
	}
}

func (suite *CliUtilsTestSuite) TestHandleSnapshotExpressions() {
	tests := []struct {
		name         string
		expression   string
		wantName     string
		wantFragment string
		wantErr      bool
	}{
		{
			name:         "valid expression without special characters works",
			expression:   "hadron",
			wantName:     "hadron",
			wantFragment: "-1",
		},
		{
			name:         "valid expression with # works",
			expression:   "hadron#12",
			wantName:     "hadron",
			wantFragment: "%2312",
		},
		{
			name:         "valid expression with ~ works 1",
			expression:   "hadron~1",
			wantName:     "hadron",
			wantFragment: "~1",
		},
		{
			name:         "valid expression with ~ works 2",
			expression:   "hadron~2",
			wantName:     "hadron",
			wantFragment: "~2",
		},
		{
			name:         "valid expression with @ works",
			expression:   "hadron@{2023-07-04T11:04:02}",
			wantName:     "hadron",
			wantFragment: "@%7B2023-07-04T11:04:02%7D",
		},
		{
			name:         "invalid expression still parsed and sent to server to handle",
			expression:   "hadron#abc",
			wantName:     "hadron",
			wantFragment: "%23abc",
		},
		{
			name:       "missing environment name with # causes an error",
			expression: "#12",
			wantErr:    true,
		},
		{
			name:       "missing environment name with ~ causes an error",
			expression: "~12",
			wantErr:    true,
		},
		{
			name:       "missing environment name with @ causes an error",
			expression: "@12",
			wantErr:    true,
		},
	}
	for _, t := range tests {
		suite.Run(t.name, func() {
			name, id, err := handleSnapshotExpressions(t.expression)
			require.True(suite.T(), err != nil == t.wantErr)
			require.Equal(suite.T(), t.wantName, name)
			require.Equal(suite.T(), t.wantFragment, id)
		})
	}
}

func (suite *CliUtilsTestSuite) TestHandleArtifactExpression() {
	tests := []struct {
		name       string
		expression string
		wantName   string
		wantId     string
		wantSep    string
		wantErr    bool
	}{
		{
			name:       "expressions without fingerprint/commit sha are invalid",
			expression: "hadron",
			wantErr:    true,
		},
		{
			name:       "valid expression with @ works",
			expression: "hadron@bd0de77b3b982927eab0bdfcc82ff2cf3dc023e0e6a5375aad7f185baa28bd30",
			wantName:   "hadron",
			wantId:     "bd0de77b3b982927eab0bdfcc82ff2cf3dc023e0e6a5375aad7f185baa28bd30",
			wantSep:    "@",
		},
		{
			name:       "valid expression with : works",
			expression: "hadron:5146ebd",
			wantName:   "hadron",
			wantId:     "5146ebd",
			wantSep:    ":",
		},
		{
			name:       "missing flow name with @ causes an error",
			expression: "@12",
			wantErr:    true,
		},
		{
			name:       "missing flow name with : causes an error",
			expression: ":12",
			wantErr:    true,
		},
		{
			name:       "invalid expression containing @ and : causes an error",
			expression: "hadron@xxx:yyy",
			wantErr:    true,
		},
		{
			name:       "invalid expression containing multiple @s does not cause an error", // fails on the server
			expression: "hadron@xxx@yyy",
			wantName:   "hadron",
			wantId:     "xxx@yyy",
			wantSep:    "@",
		},
		{
			name:       "invalid expression containing multiple :s causes an error", // fails on the server
			expression: "hadron:xxx:yyy",
			wantName:   "hadron",
			wantId:     "xxx:yyy",
			wantSep:    ":",
		},
	}
	for _, t := range tests {
		suite.Run(t.name, func() {
			name, id, sep, err := handleArtifactExpression(t.expression)
			require.True(suite.T(), err != nil == t.wantErr)
			require.Equal(suite.T(), t.wantName, name)
			require.Equal(suite.T(), t.wantId, id)
			require.Equal(suite.T(), t.wantSep, sep)
		})
	}
}
