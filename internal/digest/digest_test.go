package digest

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type DigestTestSuite struct {
	suite.Suite
	tmpDir string
}

// create a new tmpDir before each test
func (suite *DigestTestSuite) SetupTest() {
	var err error
	suite.tmpDir, err = ioutil.TempDir("", "testDir")
	require.NoError(suite.T(), err, "error creating a temporary test directory")
}

// clean up tmpDir after each test
func (suite *DigestTestSuite) AfterTest() {
	err := os.RemoveAll(suite.tmpDir)
	require.NoErrorf(suite.T(), err, "error cleaning up the temporary test directory %s", suite.tmpDir)
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *DigestTestSuite) TestFileSha256() {
	type args struct {
		filename string
		content  string
	}
	for _, t := range []struct {
		name string
		args args
		want string
	}{
		{
			name: "an empty file has a digest.",
			args: args{
				filename: "test1",
				content:  "",
			},
			want: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name: "a non-empty file has a digest.",
			args: args{
				filename: "test2",
				content:  "this is non empty",
			},
			want: "1256d6510a6606ad61a4f6104243a291b18383b456d50205eba893b51e1807bc",
		},
		{
			name: "a slightly different file has a different digest.",
			args: args{
				filename: "test3",
				content:  "this is non empty.",
			},
			want: "a50afcf37a327e0715b3148c2625bc28b3e4dcdf32b6cf78c8b8fa3ac1ebfe47",
		},
		{
			name: "a different file name with same content has the same digest.",
			args: args{
				filename: "test4",
				content:  "this is non empty.",
			},
			want: "a50afcf37a327e0715b3148c2625bc28b3e4dcdf32b6cf78c8b8fa3ac1ebfe47",
		},
	} {
		suite.Run(t.name, func() {
			testFile, err := os.Create(filepath.Join(suite.tmpDir, t.args.filename))
			require.NoErrorf(suite.T(), err, "error creating test file %s", t.args.filename)

			_, err = testFile.Write([]byte(t.args.content))
			require.NoErrorf(suite.T(), err, "error writing content to test file %s", t.args.filename)

			sha256, err := FileSha256(filepath.Join(suite.tmpDir, t.args.filename))
			require.NoErrorf(suite.T(), err, "error creating digest for test file %s", t.args.filename)

			assert.Equal(suite.T(), t.want, sha256, fmt.Sprintf("TestFileSha256: %s , got: %v -- want: %v", t.name, sha256, t.want))
		})
	}
}

func (suite *DigestTestSuite) TestDirSha256() {
	type fileSystemEntry struct {
		name     string
		content  string            // file content (if entry is a file)
		children map[string]string // dir files (if entry is dir)
	}
	type args struct {
		dirName    string
		dirContent []fileSystemEntry
	}
	for _, t := range []struct {
		name string
		args args
		want string
	}{
		{
			name: "change client (python) counterpart.",
			args: args{
				dirName: "test_dir_with_one_file_with_known_content",
				dirContent: []fileSystemEntry{
					{
						name:     "file.extra",
						content:  "this is known extra content",
						children: make(map[string]string),
					},
				},
			},
			want: "c71f5baef8cce289c9b7c971cf219e21b787a025af68ad6539b82634fe62819e",
		},
		{
			name: "an empty dir has a digest.",
			args: args{
				dirName:    "test1",
				dirContent: []fileSystemEntry{},
			},
			want: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name: "a non-empty dir has a digest.",
			args: args{
				dirName: "test2",
				dirContent: []fileSystemEntry{
					{
						name:     "sample.txt",
						content:  "some content.",
						children: make(map[string]string),
					},
				},
			},
			want: "388ab80164bbd9d96f132b046b8d09354f68b79a3668da7b507625cd1230dddf",
		},
		{
			name: "changing a file content changes the digest.",
			args: args{
				dirName: "test3",
				dirContent: []fileSystemEntry{
					{
						name:     "sample.txt",
						content:  "some content. And some more.",
						children: make(map[string]string),
					},
				},
			},
			want: "5b0e14a923d7239b0a23750a6bbfc837f71e684b8bdc2909d5ff6d90e59449c1",
		},
		{
			name: "changing a file name changes the digest.",
			args: args{
				dirName: "test4",
				dirContent: []fileSystemEntry{
					{
						name:     "sample.yaml",
						content:  "some content. And some more.",
						children: make(map[string]string),
					},
				},
			},
			want: "c38fbc1a99dad628142d0b7e2e05901362623d2b81e316d2cf650b08e93e0cef",
		},
		{
			name: "changing the root dir name doesn't change the digest.",
			args: args{
				dirName: "test44",
				dirContent: []fileSystemEntry{
					{
						name:     "sample.yaml",
						content:  "some content. And some more.",
						children: make(map[string]string),
					},
				},
			},
			want: "c38fbc1a99dad628142d0b7e2e05901362623d2b81e316d2cf650b08e93e0cef",
		},
		{
			// this test case is replicated in change for consistency
			name: "a dir with a nested dir has a digest.",
			args: args{
				dirName: "test5",
				dirContent: []fileSystemEntry{
					{
						name:     "sample.yaml",
						content:  "some content. And some more.",
						children: make(map[string]string),
					},
					{
						name: "nested-dir",
						children: map[string]string{
							"file1": "content1",
							"file2": "content2",
						},
					},
				},
			},
			want: "5d3c17dae9e208bbb92ee04ff8342abf77cb0959764def4af3ccfe9a2109d4a7",
		},
		{
			name: "a dir with a nested dir with a different name has a different digest.",
			args: args{
				dirName: "test6",
				dirContent: []fileSystemEntry{
					{
						name:     "sample.yaml",
						content:  "some content. And some more.",
						children: make(map[string]string),
					},
					{
						name: "nested-dir2",
						children: map[string]string{
							"file1": "content1",
							"file2": "content2",
						},
					},
				},
			},
			want: "db40d79b3a15b17ee9fcc2f49aa73736e0073de6b5a35c459268bb9a31e55139",
		},
	} {
		suite.Run(t.name, func() {
			dirPath := filepath.Join(suite.tmpDir, t.args.dirName)
			err := os.Mkdir(dirPath, 0777)
			require.NoErrorf(suite.T(), err, "error creating test dir %s", t.args.dirName)

			for _, entry := range t.args.dirContent {
				path := filepath.Join(suite.tmpDir, t.args.dirName, entry.name)
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

			verbose := false
			sha256, err := DirSha256(dirPath, verbose)
			require.NoErrorf(suite.T(), err, "error creating digest for test dir %s", dirPath)

			assert.Equal(suite.T(), t.want, sha256, fmt.Sprintf("TestDirSha256: %s , got: %v -- want: %v", t.name, sha256, t.want))
		})
	}
}

func (suite *DigestTestSuite) TestDirSha256Validation() {
	type args struct {
		name       string
		isFile     bool
		isAbsolute bool
	}
	for _, t := range []struct {
		name        string
		args        args
		errExpected bool
	}{
		{
			name: "non existing path raises an error",
			args: args{
				name: "test1",
			},
			errExpected: true,
		},
		{
			name: "path is a file path but not a directory",
			args: args{
				name:   "test2",
				isFile: true,
			},
			errExpected: true,
		},
		{
			name: "path is an empty string",
			args: args{
				name:       "",
				isAbsolute: true,
			},
			errExpected: true,
		},
	} {
		suite.Run(t.name, func() {
			dirPath := filepath.Join(suite.tmpDir, t.args.name)
			if t.args.isAbsolute {
				dirPath = t.args.name
			}

			if t.args.isFile {
				suite.createFileWithContent(dirPath, "")
			}

			verbose := false
			_, err := DirSha256(dirPath, verbose)
			if t.errExpected {
				require.Errorf(suite.T(), err, "TestDirSha256Validation: error was expected")
			}

		})
	}
}

func (suite *DigestTestSuite) createFileWithContent(path, content string) {
	file, err := os.Create(path)
	require.NoErrorf(suite.T(), err, "error creating test file %s", path)
	_, err = file.Write([]byte(content))
	require.NoErrorf(suite.T(), err, "error adding content to test file %s", path)
}

func (suite *DigestTestSuite) TestValidateDigest() {
	for _, t := range []struct {
		name        string
		sha256      string
		expectError bool
	}{
		{
			name:        "a valid sha256",
			sha256:      "db40d79b3a15b17ee9fcc2f49aa73736e0073de6b5a35c459268bb9a31e55139",
			expectError: false,
		},
		{
			name:        "a sha256 with characters outside [a-f0-9] is invalid",
			sha256:      "xyz0d79b3a15b17ee9fcc2f49aa73736e0073de6b5a35c459268bb9a31e55139",
			expectError: true,
		},
		{
			name:        "a sha256 with less than 64 characters is invalid",
			sha256:      "db40d79b3a15b17ee9fcc2f49aa73736e0073de6b5a3",
			expectError: true,
		},
		{
			name:        "a sha256 with more than 64 characters is invalid",
			sha256:      "db40d79b3a15b17ee9fcc2f49aa73736e0073de6b5a35c459268bb9a31e55139sd23",
			expectError: true,
		},
	} {
		suite.Run(t.name, func() {
			err := ValidateDigest(t.sha256)
			if t.expectError {
				require.Errorf(suite.T(), err, "TestValidateDigest: error was expected")
			} else {
				require.NoErrorf(suite.T(), err, "TestValidateDigest: error was NOT expected")
			}

		})
	}
}

func (suite *DigestTestSuite) TestDockerImageSha256() {
	type want struct {
		sha256      string
		expectError bool
	}
	for _, t := range []struct {
		name      string
		imageName string
		pullImage bool
		want      want
	}{
		{
			name:      "empty image name should cause an error",
			imageName: "",
			pullImage: false,
			want: want{
				expectError: true,
			},
		},
		{
			name:      "non existing image should cause an error",
			imageName: "imaginery/non-existing",
			pullImage: false,
			want: want{
				expectError: true,
			},
		},
		{
			name:      "pulled image should gets a digest",
			imageName: "library/alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
			pullImage: true,
			want: want{
				expectError: false,
				sha256:      "e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
			},
		},
	} {
		suite.Run(t.name, func() {
			if t.pullImage {
				err := pullDockerImage(t.imageName)
				require.NoErrorf(suite.T(), err, "TestDockerImageSha256: test image should be pullable")
			}
			actual, err := DockerImageSha256(t.imageName)
			if t.want.expectError {
				require.Errorf(suite.T(), err, "TestDockerImageSha256: error was expected")
			} else {
				require.NoErrorf(suite.T(), err, "TestDockerImageSha256: error was NOT expected")
				assert.Equal(suite.T(), t.want.sha256, actual, fmt.Sprintf("TestDockerImageSha256: want %s -- got %s", t.want.sha256, actual))
			}

		})
	}
}

// pullDockerImage pulls a docker image or returns an error
func pullDockerImage(imageName string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	rc, err := cli.ImagePull(context.Background(), imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer rc.Close()
	io.Copy(os.Stdout, rc)

	return nil
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDigestTestSuite(t *testing.T) {
	suite.Run(t, new(DigestTestSuite))
}
