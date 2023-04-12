package digest

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/kosli-dev/cli/internal/docker"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/utils"
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
	suite.tmpDir, err = os.MkdirTemp("", "testDir")
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

type fileEntry struct {
	name    string
	content string
}

type dirEntry struct {
	name  string
	files []fileEntry
	dirs  []dirEntry
}

func (suite *DigestTestSuite) TestDirSha256() {
	type fileSystemEntry struct {
		files []fileEntry
		dirs  []dirEntry
	}

	type args struct {
		dirName      string
		dirContent   []fileSystemEntry
		excludePaths []string
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
						files: []fileEntry{
							{
								name:    "file.extra",
								content: "this is known extra content",
							},
						},
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
						files: []fileEntry{
							{
								name:    "sample.txt",
								content: "some content.",
							},
						},
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
						files: []fileEntry{
							{
								name:    "sample.txt",
								content: "some content. And some more.",
							},
						},
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
						files: []fileEntry{
							{
								name:    "sample.yaml",
								content: "some content. And some more.",
							},
						},
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
						files: []fileEntry{
							{
								name:    "sample.yaml",
								content: "some content. And some more.",
							},
						},
					},
				},
			},
			want: "c38fbc1a99dad628142d0b7e2e05901362623d2b81e316d2cf650b08e93e0cef",
		},
		{
			name: "a dir with a nested dir has a digest.",
			args: args{
				dirName: "test5",
				dirContent: []fileSystemEntry{
					{
						files: []fileEntry{
							{
								name:    "sample.yaml",
								content: "some content. And some more.",
							},
						},
						dirs: []dirEntry{
							{
								name: "nested-dir",
								files: []fileEntry{
									{
										name:    "file1",
										content: "content1",
									},
									{
										name:    "file2",
										content: "content2",
									},
								},
							},
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
						files: []fileEntry{
							{
								name:    "sample.yaml",
								content: "some content. And some more.",
							},
						},
						dirs: []dirEntry{
							{
								name: "nested-dir2",
								files: []fileEntry{
									{
										name:    "file1",
										content: "content1",
									},
									{
										name:    "file2",
										content: "content2",
									},
								},
							},
						},
					},
				},
			},
			want: "db40d79b3a15b17ee9fcc2f49aa73736e0073de6b5a35c459268bb9a31e55139",
		},
		{
			name: "excluding dirs works with nested dir",
			args: args{
				dirName:      "exclusion1",
				excludePaths: []string{"logs"},
				dirContent: []fileSystemEntry{
					{
						files: []fileEntry{
							{
								name:    "sample.yaml",
								content: "some content. And some more.",
							},
						},
						dirs: []dirEntry{
							{
								name: "nested-dir",
								files: []fileEntry{
									{
										name:    "file1",
										content: "content1",
									},
									{
										name:    "file2",
										content: "content2",
									},
								},
							},
							{
								name: "logs",
								files: []fileEntry{
									{
										name:    "file1",
										content: "content1",
									},
								},
							},
						},
					},
				},
			},
			want: "5d3c17dae9e208bbb92ee04ff8342abf77cb0959764def4af3ccfe9a2109d4a7",
		},
		{
			name: "excluding dirs and files works with nested dir",
			args: args{
				dirName:      "exclusion2",
				excludePaths: []string{"logs", "nested-dir/file1"},
				dirContent: []fileSystemEntry{
					{
						files: []fileEntry{
							{
								name:    "sample.yaml",
								content: "some content. And some more.",
							},
						},
						dirs: []dirEntry{
							{
								name: "nested-dir",
								files: []fileEntry{
									{
										name:    "file1",
										content: "content1",
									},
									{
										name:    "file2",
										content: "content2",
									},
								},
							},
							{
								name: "logs",
								files: []fileEntry{
									{
										name:    "file1",
										content: "content1",
									},
								},
							},
						},
					},
				},
			},
			want: "2acbc9efc1f86f89086a9539244946839599b3639da7f4959744c20234cb4f40",
		},
		{
			name: "excluding dirs using glob pattern works",
			args: args{
				dirName:      "exclusion3",
				excludePaths: []string{"logs", "*/logs"},
				dirContent: []fileSystemEntry{
					{
						files: []fileEntry{
							{
								name:    "sample.yaml",
								content: "some content. And some more.",
							},
						},
						dirs: []dirEntry{
							{
								name: "nested-dir",
								files: []fileEntry{
									{
										name:    "file1",
										content: "content1",
									},
									{
										name:    "file2",
										content: "content2",
									},
								},
								dirs: []dirEntry{
									{
										name: "logs",
										files: []fileEntry{
											{
												name:    "log.txt",
												content: "this is a log",
											},
										},
									},
								},
							},
							{
								name: "logs",
								files: []fileEntry{
									{
										name:    "file1",
										content: "content1",
									},
								},
							},
						},
					},
				},
			},
			want: "5d3c17dae9e208bbb92ee04ff8342abf77cb0959764def4af3ccfe9a2109d4a7",
		},
	} {
		suite.Run(t.name, func() {
			topLevelPath := filepath.Join(suite.tmpDir, t.args.dirName)
			err := os.Mkdir(topLevelPath, 0777)
			require.NoErrorf(suite.T(), err, "error creating test dir %s", t.args.dirName)

			for _, entry := range t.args.dirContent {
				suite.createNestedDir(topLevelPath, entry.files, entry.dirs)
			}

			sha256, err := DirSha256(topLevelPath, t.args.excludePaths, logger.NewStandardLogger())
			require.NoErrorf(suite.T(), err, "error creating digest for test dir %s", topLevelPath)

			assert.Equal(suite.T(), t.want, sha256, fmt.Sprintf("TestDirSha256: %s , got: %v -- want: %v", t.name, sha256, t.want))
		})
	}
}

func (suite *DigestTestSuite) createNestedDir(path string, files []fileEntry, dirs []dirEntry) {
	for _, f := range files {
		filePath := filepath.Join(path, f.name)
		suite.createFileWithContent(filePath, f.content)
	}
	for _, d := range dirs {
		nestedPath := filepath.Join(path, d.name)
		err := os.Mkdir(nestedPath, 0777)
		require.NoErrorf(suite.T(), err, "error creating nested test dir %s", nestedPath)
		suite.createNestedDir(nestedPath, d.files, d.dirs)
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

			_, err := DirSha256(dirPath, []string{}, logger.NewStandardLogger())
			if t.errExpected {
				require.Errorf(suite.T(), err, "TestDirSha256Validation: error was expected")
			}

		})
	}
}

func (suite *DigestTestSuite) createFileWithContent(path, content string) {
	err := utils.CreateFileWithContent(path, content)
	require.NoErrorf(suite.T(), err, "error creating file %s", path)
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
			name:      "pulled image should get a digest",
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
				err := docker.PullDockerImage(t.imageName)
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

func (suite *DigestTestSuite) TestRemoteDockerImageSha256() {
	type want struct {
		sha256      string
		expectError bool
	}
	for _, t := range []struct {
		name           string
		imageName      string
		localImageName string
		localImageTag  string
		pullImage      bool
		want           want
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
			name:           "registry returns a digest for an existing image",
			imageName:      "library/alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
			localImageName: "local-registry/alpine",
			localImageTag:  "v1",
			pullImage:      true,
			want: want{
				expectError: false,
				sha256:      "e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
			},
		},
	} {
		suite.Run(t.name, func() {
			if t.pullImage {
				err := docker.PullDockerImage(t.imageName)
				require.NoErrorf(suite.T(), err, "TestRemoteDockerImageSha256: test image should be pullable")

				localImage := fmt.Sprintf("localhost:5001/%s:%s", t.localImageName, t.localImageTag)
				err = docker.TagDockerImage(t.imageName, localImage)
				require.NoErrorf(suite.T(), err, "TestRemoteDockerImageSha256: test image should be taggable")

				err = docker.PushDockerImage(localImage)
				require.NoErrorf(suite.T(), err, "TestRemoteDockerImageSha256: test image should be pushable")
			}
			actual, err := RemoteDockerImageSha256(t.localImageName, t.localImageTag, "http://localhost:5001/v2", "secret",
				logger.NewStandardLogger())
			if t.want.expectError {
				require.Errorf(suite.T(), err, "TestRemoteDockerImageSha256: error was expected")
			} else {
				require.NoErrorf(suite.T(), err, "TestRemoteDockerImageSha256: error was NOT expected")
				assert.Equal(suite.T(), t.want.sha256, actual, fmt.Sprintf("TestRemoteDockerImageSha256: want %s -- got %s", t.want.sha256, actual))
			}

		})
	}
}

func (suite *DigestTestSuite) TestExtractImageDigestFromRepoDigest() {
	type want struct {
		sha256      string
		expectError bool
	}
	for _, t := range []struct {
		name        string
		imageID     string
		repoDigests []string
		want        want
	}{
		{
			name:        "empty image ID should cause an error",
			imageID:     "",
			repoDigests: []string{"example@sha256:afcc7f1ac1b49db317a7196c902e61c6c3c4607d63599ee1a82d702d249a0ccb"},
			want: want{
				expectError: true,
			},
		},
		{
			name:        "empty repoDigests should cause an error",
			imageID:     "example",
			repoDigests: []string{},
			want: want{
				expectError: true,
			},
		},
		{
			name:        "if repoDigests has only item, the digest is returned from it",
			imageID:     "example",
			repoDigests: []string{"example@sha256:afcc7f1ac1b49db317a7196c902e61c6c3c4607d63599ee1a82d702d249a0ccb"},
			want: want{
				sha256: "afcc7f1ac1b49db317a7196c902e61c6c3c4607d63599ee1a82d702d249a0ccb",
			},
		},
		{
			name:    "if imageID is an ID (not a name), the returned digest is the first item in repoDigests",
			imageID: "12adea71a33bcce0925f5b2e951992cc2d8b69f4051122e93d5c35000e9b9e28",
			repoDigests: []string{
				"example@sha256:afcc7f1ac1b49db317a7196c902e61c6c3c4607d63599ee1a82d702d249a0ccb",
				"internal.registry.example.com:5000/example@sha256:b69959407d21e8a062e0416bf13405bb2b71ed7a84dde4158ebafacfa06f5578",
			},
			want: want{
				sha256: "afcc7f1ac1b49db317a7196c902e61c6c3c4607d63599ee1a82d702d249a0ccb",
			},
		},
		{
			name:    "if repoDigests has multiple items and image ID is a name, the matching digest is returned",
			imageID: "alpine",
			repoDigests: []string{
				"alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
				"localhost:5001/local-registry/alpine@sha256:afcc7f1ac1b49db317a7196c902e61c6c3c4607d63599ee1a82d702d249a0ccb",
			},
			want: want{
				sha256: "e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
			},
		},
		{
			name:    "for dockerhub images, the library prefix does is skipped from the image name",
			imageID: "library/alpine",
			repoDigests: []string{
				"alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
				"localhost:5001/local-registry/alpine@sha256:afcc7f1ac1b49db317a7196c902e61c6c3c4607d63599ee1a82d702d249a0ccb",
			},
			want: want{
				sha256: "e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
			},
		},
		{
			name:    "if the image ID is a name and it contains the sha256, the sha256 is skipped from the image name",
			imageID: "alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
			repoDigests: []string{
				"alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
				"localhost:5001/local-registry/alpine@sha256:afcc7f1ac1b49db317a7196c902e61c6c3c4607d63599ee1a82d702d249a0ccb",
			},
			want: want{
				sha256: "e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
			},
		},
		{
			name:    "if the image ID is a name and it contains the tag, the tag is skipped from the image name",
			imageID: "alpine:v1",
			repoDigests: []string{
				"alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
				"localhost:5001/local-registry/alpine@sha256:afcc7f1ac1b49db317a7196c902e61c6c3c4607d63599ee1a82d702d249a0ccb",
			},
			want: want{
				sha256: "e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
			},
		},
		{
			name:    "when the image name does not have a match in repoDigests, an error is returned",
			imageID: "example",
			repoDigests: []string{
				"alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5",
				"localhost:5001/local-registry/alpine@sha256:afcc7f1ac1b49db317a7196c902e61c6c3c4607d63599ee1a82d702d249a0ccb",
			},
			want: want{
				expectError: true,
			},
		},
	} {
		suite.Run(t.name, func() {
			actual, err := extractImageDigestFromRepoDigest(t.imageID, t.repoDigests)
			if t.want.expectError {
				require.Errorf(suite.T(), err, "TestExtractImageDigestFromRepoDigest: error was expected")
			} else {
				require.NoErrorf(suite.T(), err, "TestExtractImageDigestFromRepoDigest: error was NOT expected")
				assert.Equal(suite.T(), t.want.sha256, actual, fmt.Sprintf("TestExtractImageDigestFromRepoDigest: want %s -- got %s", t.want.sha256, actual))
			}
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDigestTestSuite(t *testing.T) {
	suite.Run(t, new(DigestTestSuite))
}
