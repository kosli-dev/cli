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

type fileSystemEntry struct {
	files      map[string]string          // file names in a dir and their content
	nestedDirs map[string]fileSystemEntry // nested dirs in a dir
}

func (suite *ServerTestSuite) TestCreateServerArtifactsData() {

	for _, t := range []struct {
		name         string
		fileSystem   map[string]fileSystemEntry
		paths        []string
		excludePaths []string
		want         []map[string]string
	}{
		{
			name: "can get artifact data for a single path",
			fileSystem: map[string]fileSystemEntry{
				"directory-name": {
					files: map[string]string{
						"sample.txt": "some content.",
					},
					nestedDirs: make(map[string]fileSystemEntry),
				},
			},
			paths: []string{"directory-name"},
			want: []map[string]string{
				{"directory-name": "388ab80164bbd9d96f132b046b8d09354f68b79a3668da7b507625cd1230dddf"},
			},
		},
		{
			name: "can get artifact data for two paths",
			fileSystem: map[string]fileSystemEntry{
				"directory-name2": {
					files: map[string]string{
						"sample.txt": "some content.",
					},
					nestedDirs: make(map[string]fileSystemEntry),
				},
				"directory-name3": {
					files: map[string]string{
						"sample-2.txt": "some more content.",
					},
					nestedDirs: make(map[string]fileSystemEntry),
				},
			},
			paths: []string{"directory-name2", "directory-name3"},
			want: []map[string]string{
				{"directory-name2": "388ab80164bbd9d96f132b046b8d09354f68b79a3668da7b507625cd1230dddf"},
				{"directory-name3": "3025bae51416f4348cbeaf3d2f7394a21d637792c850fb63d6c5242f073bc9b3"},
			},
		},
		{
			name: "can get digest of a directory containing a file with a space in its name",
			fileSystem: map[string]fileSystemEntry{
				"directory-name4": {
					files: map[string]string{
						"SPV STS Prod.pem": "content",
					},
					nestedDirs: make(map[string]fileSystemEntry),
				},
			},
			paths: []string{"directory-name4"},
			want: []map[string]string{
				{"directory-name4": "0cccd704109ea889ed18739f9e0ed610b7993512a23998a6491886a9de77845d"},
			},
		},
		{
			name: "can get data with excludePaths that excludes a file",
			fileSystem: map[string]fileSystemEntry{
				"top-dir": {
					files: map[string]string{
						"file1.txt": "some content.",
						"app.log":   "some logs here.",
					},
					nestedDirs: make(map[string]fileSystemEntry),
				},
			},
			paths:        []string{"top-dir"},
			excludePaths: []string{"*.log"},
			want: []map[string]string{
				{"top-dir": "6b8174228c507be8c8d8482c2516b9c4775e401810db649b0cab40e75104f3a0"},
			},
		},
		{
			name: "can get data with paths containing a Glob pattern",
			fileSystem: map[string]fileSystemEntry{
				"glob-dir1": { // should be included
					files: map[string]string{},
					nestedDirs: map[string]fileSystemEntry{
						"nested-glob-1": {
							files: map[string]string{
								"file1": "new content",
							},
							nestedDirs: map[string]fileSystemEntry{},
						},
					},
				},
				"glob-dir2": { // should be included
					files: map[string]string{
						"file2.exe": "this is an executable.",
						"app.log":   "some logs here.", // should be excluded
					},
					nestedDirs: make(map[string]fileSystemEntry),
				},
				"non-glob-dir": {
					files: map[string]string{
						"file1.custom": "some content.", // should be included as a separate artifact
						"app.log":      "some logs here.",
					},
					nestedDirs: make(map[string]fileSystemEntry),
				},
			},
			paths:        []string{"glob-dir*", "*/*.custom"},
			excludePaths: []string{"*.log"},
			want: []map[string]string{
				{
					"glob-dir1": "3efd00f4aa9088a0c07629b308bb77e338fc229ab032711f8202cb76f6d2e8e9",
				},
				{
					"glob-dir2": "94839bc10d01803d9ebbacd5ab9edc470176fe233ea1f13861a54f604066c47d",
				},
				{
					"non-glob-dir/file1.custom": "593cfe761544e1363f7594b403c222a4d93cc3f15246ba88d1efc3cb8e817cc5",
				},
			},
		},
	} {
		suite.Run(t.name, func() {
			suite.setupTestFileSystem(suite.tmpDir, t.fileSystem)
			if t.excludePaths == nil {
				t.excludePaths = []string{}
			}

			for i, path := range t.paths {
				t.paths[i] = filepath.Join(suite.tmpDir, path)
			}

			serverData, err := CreateServerArtifactsData(t.paths, t.excludePaths, logger.NewStandardLogger())
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

func (suite *ServerTestSuite) setupTestFileSystem(tmpDir string, fs map[string]fileSystemEntry) {
	for entryName, fsEntry := range fs {
		dirPath := filepath.Join(tmpDir, entryName)
		err := os.Mkdir(dirPath, 0777)
		require.NoErrorf(suite.T(), err, "error creating test dir %s", entryName)

		for fileName, fileContent := range fsEntry.files {
			suite.createFileWithContent(filepath.Join(dirPath, fileName), fileContent)
		}
		suite.setupTestFileSystem(dirPath, fsEntry.nestedDirs)
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

func (suite *ServerTestSuite) TestCreatePathsArtifactsData() {

	for _, t := range []struct {
		name      string
		pathsSpec *PathsSpec
		want      []map[string]string
		wantError bool
	}{
		{
			name: "can get artifact data for a dir",
			pathsSpec: &PathsSpec{
				Version: 1,
				Artifacts: map[string]ArtifactPathSpec{
					"differ": {
						Path:   "testdata/folder1",
						Ignore: []string{"empty"},
					},
				},
			},
			want: []map[string]string{
				{"differ": "51687c011a07c59b6ae9e774e6dc8b5b85343c1a0cfad2b5a0c3613744d19d2b"},
			},
		},
		{
			name: "can get artifact data for a file",
			pathsSpec: &PathsSpec{
				Version: 1,
				Artifacts: map[string]ArtifactPathSpec{
					"differ": {
						Path:   "testdata/folder1/full.txt",
						Ignore: []string{"empty"},
					},
				},
			},
			want: []map[string]string{
				{"differ": "ff9c0fc39bdcbd5770c67fb1bf49d10f1815fc028edf1a6d83ddb75b64ae85be"},
			},
		},
	} {
		suite.Run(t.name, func() {
			serverData, err := CreatePathsArtifactsData(t.pathsSpec, logger.NewStandardLogger())
			require.Equal(suite.T(), t.wantError, err != nil, err)

			digestsList := []map[string]string{}

			for i, data := range serverData {
				digestsList = append(digestsList, data.Digests)
				assert.NotEqual(suite.T(), int64(0), data.CreationTimestamp, fmt.Sprintf("TestCreatePathsArtifactsData: %s , got: %v, should not be 0, at index: %d", t.name, data.CreationTimestamp, i))
			}

			for artifactName, digest := range t.want {
				require.Equal(suite.T(), digest, digestsList[artifactName])
			}
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
