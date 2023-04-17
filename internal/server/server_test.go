package server

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ServerTestSuite struct {
	suite.Suite
	tmpDir string
}

// create a new tmpDir before each test
func (suite *ServerTestSuite) SetupTest() {
	var err error
	suite.tmpDir, err = os.MkdirTemp("", "testDir")
	require.NoError(suite.T(), err, "error creating a temporary test directory")
}

// clean up tmpDir after each test
func (suite *ServerTestSuite) AfterTest() {
	err := os.RemoveAll(suite.tmpDir)
	require.NoErrorf(suite.T(), err, "error cleaning up the temporary test directory %s", suite.tmpDir)
}

func (suite *ServerTestSuite) TestCreateServerArtifactsData() {
	type fileSystemEntry struct {
		name     string
		content  string            // file content (if entry is a file)
		children map[string]string // dir files (if entry is dir)
	}

	for _, t := range []struct {
		name       string
		fileSystem map[string][]fileSystemEntry
		want       []map[string]string
	}{
		{
			name: "can get artifact data for a single path",
			fileSystem: map[string][]fileSystemEntry{
				"directory-name": {
					{
						name:     "sample.txt",
						content:  "some content.",
						children: make(map[string]string),
					},
				},
			},

			want: []map[string]string{
				{"directory-name": "388ab80164bbd9d96f132b046b8d09354f68b79a3668da7b507625cd1230dddf"},
			},
		},
		{
			name: "can get artifact data for two paths",
			fileSystem: map[string][]fileSystemEntry{
				"directory-name2": {
					{
						name:     "sample.txt",
						content:  "some content.",
						children: make(map[string]string),
					},
				},
				"directory-name3": {
					{
						name:     "sample-2.txt",
						content:  "some more content.",
						children: make(map[string]string),
					},
				},
			},

			want: []map[string]string{
				{"directory-name2": "388ab80164bbd9d96f132b046b8d09354f68b79a3668da7b507625cd1230dddf"},
				{"directory-name3": "3025bae51416f4348cbeaf3d2f7394a21d637792c850fb63d6c5242f073bc9b3"},
			},
		},
		{
			name: "can get digest of a directory containing a file with a space in its name",
			fileSystem: map[string][]fileSystemEntry{
				"directory-name4": {
					{
						name:     "SPV STS Prod.pem",
						content:  "content",
						children: make(map[string]string),
					},
				},
			},

			want: []map[string]string{
				{"directory-name4": "0cccd704109ea889ed18739f9e0ed610b7993512a23998a6491886a9de77845d"},
			},
		},
	} {
		suite.Run(t.name, func() {
			paths := []string{}
			for dirName, dirContent := range t.fileSystem {
				dirPath := filepath.Join(suite.tmpDir, dirName)
				err := os.Mkdir(dirPath, 0777)
				require.NoErrorf(suite.T(), err, "error creating test dir %s", dirName)
				paths = append(paths, dirPath)

				for _, entry := range dirContent {
					path := filepath.Join(suite.tmpDir, dirName, entry.name)
					if len(entry.children) == 0 { // file
						suite.createFileWithContent(path, entry.content)
					} else { // dir
						err := os.Mkdir(path, 0777)
						require.NoErrorf(suite.T(), err, "error creating test dir %s", path)
						for name, data := range entry.children {
							filePath := filepath.Join(path, name)
							suite.createFileWithContent(filePath, data)
						}
					}
				}
			}

			serverData, err := CreateServerArtifactsData(paths, []string{}, logger.NewStandardLogger())
			require.NoErrorf(suite.T(), err, "error creating server artifact data: %v", err)

			digestsList := []map[string]string{}

			for i, data := range serverData {
				digestsList = append(digestsList, data.Digests)
				assert.NotEqual(suite.T(), int64(0), data.CreationTimestamp, fmt.Sprintf("TestCreateServerArtifactsData: %s , got: %v, should not be 0, at index: %d", t.name, data.CreationTimestamp, i))
			}

			expected := []map[string]string{}
			for _, m := range t.want {
				tmpMap := make(map[string]string)
				for k, v := range m {
					tmpMap[filepath.Join(suite.tmpDir, k)] = v
				}
				expected = append(expected, tmpMap)
			}
			assert.ElementsMatch(suite.T(), expected, digestsList, fmt.Sprintf("TestCreateServerArtifactsData: %s , got: %v -- want: %v", t.name, digestsList, expected))

		})
	}
}

func (suite *ServerTestSuite) TestCreateServerArtifactsDataWithFiles() {
	type args struct {
		name    string
		content string
		create  bool
	}

	for _, t := range []struct {
		name        string
		args        args
		expectError bool
		want        []map[string]string
	}{
		{
			name: "can get artifact data for a single file",
			args: args{
				name:    "sample.txt",
				content: "some content.",
				create:  true,
			},
			expectError: false,
			want: []map[string]string{
				{"sample.txt": "593cfe761544e1363f7594b403c222a4d93cc3f15246ba88d1efc3cb8e817cc5"},
			},
		},
		{
			name: "attempting to get artifact data for a non-existing file returns error",
			args: args{
				name:    "sample2.txt",
				content: "some content.",
				create:  false,
			},
			expectError: true,
			want:        []map[string]string{},
		},
	} {
		suite.Run(t.name, func() {
			paths := []string{}
			path := filepath.Join(suite.tmpDir, t.args.name)
			paths = append(paths, path)
			if t.args.create {
				suite.createFileWithContent(path, t.args.content)
			}

			serverData, err := CreateServerArtifactsData(paths, []string{}, logger.NewStandardLogger())
			if t.expectError {
				require.Errorf(suite.T(), err, "was expecting error during creating server artifact data but got none")
			} else {
				require.NoErrorf(suite.T(), err, "error creating server artifact data was NOT expected: %v", err)

				digestsList := []map[string]string{}

				for i, data := range serverData {
					digestsList = append(digestsList, data.Digests)
					assert.NotEqual(suite.T(), int64(0), data.CreationTimestamp, fmt.Sprintf("TestCreateServerArtifactsDataWithFiles: %s , got: %v, should not be 0, at index: %d", t.name, data.CreationTimestamp, i))
				}

				expected := []map[string]string{}
				for _, m := range t.want {
					tmpMap := make(map[string]string)
					for k, v := range m {
						tmpMap[filepath.Join(suite.tmpDir, k)] = v
					}
					expected = append(expected, tmpMap)
				}
				assert.ElementsMatch(suite.T(), expected, digestsList, fmt.Sprintf("TestCreateServerArtifactsDataWithFiles: %s , got: %v -- want: %v", t.name, digestsList, expected))
			}

		})
	}
}

func (suite *ServerTestSuite) TestCreateServerArtifactsDataInvalid() {

	paths := []string{"a/b/c"}

	_, err := CreateServerArtifactsData(paths, []string{}, logger.NewStandardLogger())
	require.Errorf(suite.T(), err, "error was expected")
}

func (suite *ServerTestSuite) TestGetPathLastModifiedTimestamp() {
	ts, err := getPathLastModifiedTimestamp("server.go")
	require.NoError(suite.T(), err)
	require.Greater(suite.T(), ts, int64(0))

	ts, err = getPathLastModifiedTimestamp("non-existing")
	require.Error(suite.T(), err)
	require.Equal(suite.T(), ts, int64(0))
}

func (suite *ServerTestSuite) createFileWithContent(path, content string) {
	err := utils.CreateFileWithContent(path, content)
	require.NoErrorf(suite.T(), err, "error creating file %s", path)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
