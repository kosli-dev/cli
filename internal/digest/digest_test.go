package digest

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
			want: "f29c4d614fa3c1fa5e8b82239ad698febe7de2329b7fcc7b35e08e892bc3da85",
		},
		{
			name: "an empty dir has a digest.",
			args: args{
				dirName:    "test1",
				dirContent: []fileSystemEntry{},
			},
			want: "ab0ee213d0bc9b7f69411817874fdfe6550c640b5479e5111b90ccd566c1163b",
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
			want: "d32fbe18ef42d44c093f8c4e645cbe59ab0e8908462d2f94dac91a530064bd02",
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
			want: "6e8a8c47e0cf60365ca7de56b6e04c671d970e5e54c7b318143741047694edaa",
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
			want: "1fb9e9d620cd2e82a38c92603213ea13ca3b006bab75e71b48336d1e5d8b8901",
		},
		{
			name: "changing the dir name changes the digest.",
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
			want: "70f73a3ffa71818a39f8e16bd46f756a602a75f16dc4412c3cc90a15f5776d99",
		},
		{
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
			want: "7624535c2de214d13227c9ed83a3cd1359a1b30da8369942c0741e214c40f7e3",
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
			want: "865a17aa813c982c474fb61c9331c460b713636ce77f15e79834d3e98e10e47d",
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

			sha256, err := DirSha256(dirPath)
			require.NoErrorf(suite.T(), err, "error creating digest for test dir %s", dirPath)

			assert.Equal(suite.T(), t.want, sha256, fmt.Sprintf("TestDirSha256: %s , got: %v -- want: %v", t.name, sha256, t.want))
		})
	}
}

func (suite *DigestTestSuite) TestDirSha256Validation() {
	type args struct {
		name   string
		isFile bool
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
	} {
		suite.Run(t.name, func() {
			dirPath := filepath.Join(suite.tmpDir, t.args.name)

			if t.args.isFile {
				suite.createFileWithContent(dirPath, "")
			}

			_, err := DirSha256(dirPath)
			if t.errExpected {
				require.Errorf(suite.T(), err, fmt.Sprintf("TestDirSha256Validation: error was expected"))
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

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDigestTestSuite(t *testing.T) {
	suite.Run(t, new(DigestTestSuite))
}
