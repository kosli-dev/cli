package utils

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
			tmpDir, err := ioutil.TempDir("", "testDir")
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

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}
