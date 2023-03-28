package utils

import (
	"fmt"
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
type UtilsTestSuite struct {
	suite.Suite
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *UtilsTestSuite) TestContains() {
	type args struct {
		list []string
		item string
	}
	for _, t := range []struct {
		name string
		args args
		want bool
	}{
		{
			name: "item is not found when the list is empty.",
			args: args{
				list: []string{},
				item: "foo",
			},
			want: false,
		},
		{
			name: "item is found when the list contains it.",
			args: args{
				list: []string{"foo", "bar"},
				item: "foo",
			},
			want: true,
		},
		{
			name: "item is not found when the list does not contain it.",
			args: args{
				list: []string{"foo", "bar"},
				item: "example",
			},
			want: false,
		},
	} {
		suite.Run(t.name, func() {
			actual := Contains(t.args.list, t.args.item)
			assert.Equal(suite.T(), t.want, actual, fmt.Sprintf("TestContains: %s , got: %v -- want: %v", t.name, actual, t.want))
		})
	}
}

func (suite *UtilsTestSuite) TestIsJSON() {
	for _, t := range []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "empty JSON returns true.",
			input: "{}",
			want:  true,
		},
		{
			name:  "valid JSON returns true.",
			input: "{\"a\": 50}",
			want:  true,
		},
		{
			name:  "invalid JSON returns false.",
			input: "",
			want:  false,
		},
	} {
		suite.Run(t.name, func() {
			actual := IsJSON(t.input)
			assert.Equal(suite.T(), t.want, actual, fmt.Sprintf("TestIsJSON: %s , got: %v -- want: %v", t.name, actual, t.want))
		})
	}
}

func (suite *UtilsTestSuite) TestLoadFileContent() {
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
			name: "an empty file is loaded as an empty string.",
			args: args{
				filename: "test1",
				content:  "",
				create:   true,
			},
			expectError: false,
		},
		{
			name: "a file with content is loaded correctly.",
			args: args{
				filename: "test2",
				content:  "",
				create:   true,
			},
			expectError: false,
		},
		{
			name: "a non-existing file returns an error.",
			args: args{
				filename: "test3",
				content:  "",
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

			actual, err := LoadFileContent(filepath.Join(tmpDir, t.args.filename))
			if t.expectError {
				require.Errorf(suite.T(), err, "loading content for test file %s IS expected to return an error", t.args.filename)
			} else {
				require.NoErrorf(suite.T(), err, "loading content for test file %s is NOT expected to return an error", t.args.filename)
			}

			assert.Equal(suite.T(), t.args.content, actual, fmt.Sprintf("TestLoadFileContent: %s , got: %v -- want: %v", t.name, actual, t.args.content))
		})
	}
}

func (suite *UtilsTestSuite) TestCreateFile() {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "test.txt")
	f, err := CreateFile(path)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), f)
	require.FileExists(suite.T(), path)
}

func (suite *UtilsTestSuite) TestIsFileIsDir() {
	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(suite.T(), err)
	defer os.RemoveAll(tmpDir)

	ok, err := IsDir(tmpDir)
	require.NoError(suite.T(), err)
	require.True(suite.T(), ok)

	path := filepath.Join(tmpDir, "test.txt")
	_, err = CreateFile(path)
	require.NoError(suite.T(), err)
	require.FileExists(suite.T(), path)

	ok, err = IsFile(path)
	require.NoError(suite.T(), err)
	require.True(suite.T(), ok)

	ok, err = IsDir(path)
	require.NoError(suite.T(), err)
	require.False(suite.T(), ok)

	nonExistingPath := filepath.Join(tmpDir, "non-existing.txt")
	ok, err = IsFile(nonExistingPath)
	require.Error(suite.T(), err)
	require.False(suite.T(), ok)

	nonExistingDir := "non-existing"
	ok, err = IsDir(nonExistingDir)
	require.Error(suite.T(), err)
	require.False(suite.T(), ok)
}

func (suite *UtilsTestSuite) TestTar() {
	for _, t := range []struct {
		name            string
		srcType         string
		shouldCreateSrc bool
		tarFileName     string
		wantError       bool
	}{
		{
			name:            "can tar a file",
			srcType:         "file",
			shouldCreateSrc: true,
			tarFileName:     "file.tgz",
		},
		{
			name:            "can tar a dir",
			srcType:         "dir",
			shouldCreateSrc: true,
			tarFileName:     "dir.tgz",
		},
		{
			name:            "fails when src does not exist",
			shouldCreateSrc: false,
			wantError:       true,
		},
	} {
		suite.Run(t.name, func() {
			path := "non-existing"
			if t.shouldCreateSrc {
				tmpDir, err := os.MkdirTemp("", "")
				require.NoError(suite.T(), err)
				suite.createFileWithContent(filepath.Join(tmpDir, "newfile.txt"), "Hello World!")
				defer os.RemoveAll(tmpDir)
				path = tmpDir
				if t.srcType == "file" {
					path = filepath.Join(tmpDir, "some-file.txt")
					suite.createFileWithContent(path, "Hello World!")
				}
			}

			tarPath, err := Tar(path, t.tarFileName)
			require.True(suite.T(), t.wantError == (err != nil))
			if !t.wantError {
				defer os.RemoveAll(tarPath)
				require.Equal(suite.T(), t.tarFileName, filepath.Base(tarPath))
			}
		})
	}
}

func (suite *UtilsTestSuite) createFileWithContent(path, content string) {
	err := CreateFileWithContent(path, content)
	require.NoErrorf(suite.T(), err, "error creating file %s", path)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}
