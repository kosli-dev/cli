package server

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

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
	suite.tmpDir, err = ioutil.TempDir("", "testDir")
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
			name: "can get a artifact data for a single path",
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
				{"directory-name": "ccfac53009268f5db73ae31ef062c832e20d23ffa66d6b4ee9923dd4fac8676c"},
			},
		},
		{
			name: "can get a artifact data for two paths",
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
				{"directory-name2": "4fadb5e568c94bd6adebc8bbead8492df170f4245725d047bb731655099317a8"},
				{"directory-name3": "da9f25ae376572c038ecd9aea2b7d6a5c7ac133cdb18bb99e8a9d6c39c6866b7"},
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

			serverData, err := CreateServerArtifactsData(paths)
			require.NoErrorf(suite.T(), err, "error creating server artifact data: %v", err)

			digestsList := []map[string]string{}

			for i, data := range serverData {
				digestsList = append(digestsList, data.Digests)
				assert.NotEqual(suite.T(), int64(0), data.CreationTimestamp, fmt.Sprintf("TestCreateServerArtifactsData: %s , got: %v, should not be 0, at index: %d", t.name, data.CreationTimestamp, i))
			}
			assert.ElementsMatch(suite.T(), t.want, digestsList, fmt.Sprintf("TestCreateServerArtifactsData: %s , got: %v -- want: %v", t.name, digestsList, t.want))

		})
	}
}

func (suite *ServerTestSuite) TestCreateServerArtifactsDataInvalid() {

	paths := []string{"a/b/c"}

	_, err := CreateServerArtifactsData(paths)
	require.Errorf(suite.T(), err, "error was expected")
}

func (suite *ServerTestSuite) createFileWithContent(path, content string) {
	file, err := os.Create(path)
	require.NoErrorf(suite.T(), err, "error creating test file %s", path)
	_, err = file.Write([]byte(content))
	require.NoErrorf(suite.T(), err, "error adding content to test file %s", path)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDigestTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
