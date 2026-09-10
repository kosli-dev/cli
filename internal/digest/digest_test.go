package digest

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kosli-dev/cli/internal/docker"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
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

			sha256, err := FileSha256(filepath.Join(suite.tmpDir, t.args.filename), logger.NewStandardLogger())
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
		{
			name: ".kosli_ignore is included in the fingerprint",
			args: args{
				dirName: "exclusion4",
				dirContent: []fileSystemEntry{
					{
						files: []fileEntry{
							{
								name:    "sample.yaml",
								content: "some content. And some more.",
							},
							{
								name:    ".kosli_ignore",
								content: "logs\n*/logs",
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
			want: "4048004ce6f91e8ce27c35e5e857948a46750567cf7c873ee32b120f4318363c",
		},
		{
			name: "excluding dirs using .kosli_ignore in root works",
			args: args{
				dirName: "exclusion5",
				dirContent: []fileSystemEntry{
					{
						files: []fileEntry{
							{
								name:    "sample.yaml",
								content: "some content. And some more.",
							},
							{
								name:    ".kosli_ignore",
								content: "logs\n*/logs",
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
			want: "4048004ce6f91e8ce27c35e5e857948a46750567cf7c873ee32b120f4318363c",
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

// Differential rather than golden: what matters is that the two trees differ.
func (suite *DigestTestSuite) TestDirSha256IgnoreFileCannotHideItself() {
	baseline := map[string]string{
		"app/index.js":    "console.log(1)",
		"app/lib/util.js": "exports.x = 1",
	}
	for i, t := range []struct {
		name         string
		ignore       string
		excludePaths []string
		wantEqual    bool
	}{
		{
			name:   "a self-excluding ignore file cannot hide an added file",
			ignore: ".kosli_ignore\napp/backdoor.js",
		},
		{
			name:   "a glob matching the ignore file cannot hide it",
			ignore: "*ignore*\napp/backdoor.js",
		},
		{
			name:   "a dotted glob matching the ignore file cannot hide it",
			ignore: ".kosli*\napp/backdoor.js",
		},
		{
			// filepathx concatenates the pieces of a ** pattern, so this resolves to
			// "dir//.kosli_ignore", which the walk never emits.
			name:   "a ** glob matching the ignore file cannot hide it",
			ignore: "**/.kosli_ignore\napp/backdoor.js",
		},
		{
			// Bare ** excludes every path, so this differs from the baseline whether
			// or not the ignore file is protected. It documents; the three cases above
			// pin.
			name:   "a bare ** cannot hide the ignore file",
			ignore: "**\napp/backdoor.js",
		},
		{
			// The entries still come from the tree, so a file listed there is hidden.
			// Pinned as the migration path, not as a safety property.
			name:         "--exclude of the ignore file keeps applying the entries it lists",
			ignore:       "app/backdoor.js",
			excludePaths: []string{".kosli_ignore"},
			wantEqual:    true,
		},
	} {
		suite.Run(t.name, func() {
			approved := suite.createDirWithFiles(fmt.Sprintf("approved%d", i), baseline)
			deployed := suite.createDirWithFiles(fmt.Sprintf("deployed%d", i), baseline)
			suite.createFileWithContent(filepath.Join(deployed, "app/backdoor.js"), "BACKDOOR")
			suite.createFileWithContent(filepath.Join(deployed, ".kosli_ignore"), t.ignore)

			approvedSha, err := DirSha256(approved, t.excludePaths, logger.NewStandardLogger())
			require.NoError(suite.T(), err)
			deployedSha, err := DirSha256(deployed, t.excludePaths, logger.NewStandardLogger())
			require.NoError(suite.T(), err)

			if t.wantEqual {
				assert.Equal(suite.T(), approvedSha, deployedSha)
			} else {
				assert.NotEqual(suite.T(), approvedSha, deployedSha,
					"the added file is invisible to the fingerprint")
			}
		})
	}
}

// Pinned at the function because DirSha256 cannot reach it: WalkDir reads the same
// directory and surfaces a persistent failure itself.
func (suite *DigestTestSuite) TestIgnoreFilePathInTreeFailsClosedOnAnUnreadableDir() {
	dir := suite.createDirWithFiles("unreadable", map[string]string{".kosli_ignore": ".kosli_ignore"})
	// Write and search but not read: Lstat resolves while ReadDir fails.
	require.NoError(suite.T(), os.Chmod(dir, 0300))
	defer func() { _ = os.Chmod(dir, 0700) }()

	if _, err := os.ReadDir(dir); err == nil {
		suite.T().Skip("directory is readable regardless of mode, probably running as root")
	}

	path, err := ignoreFilePathInTree(dir)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "", path)
}

func (suite *DigestTestSuite) TestDirSha256LogsWhenTheIgnoreFileIsKept() {
	dir := suite.createDirWithFiles("kept", map[string]string{"app/index.js": "console.log(1)"})
	suite.createFileWithContent(filepath.Join(dir, ".kosli_ignore"), ".kosli_ignore")

	var out bytes.Buffer
	_, err := DirSha256(dir, nil, logger.NewLogger(io.Discard, &out, true))
	require.NoError(suite.T(), err)

	assert.Contains(suite.T(), out.String(), "an exclusion list cannot exclude itself")
}

// The silent legs carry this test: the warning must speak only when the exclusion
// reaches the walk.
func (suite *DigestTestSuite) TestDirSha256WarnsOnlyWhenAFlagExclusionWeakensTheFingerprint() {
	const warning = "is excluded by a flag"
	for i, t := range []struct {
		name         string
		ignore       string
		excludePaths []string
		wantWarning  bool
	}{
		{
			name:         "excluding an ignore file that carries rules",
			ignore:       "logs",
			excludePaths: []string{".kosli_ignore"},
			wantWarning:  true,
		},
		{
			name:         "excluding an empty ignore file still weakens the fingerprint",
			ignore:       "",
			excludePaths: []string{".kosli_ignore"},
			wantWarning:  true,
		},
		{
			name:         "excluding a comment-only ignore file still weakens the fingerprint",
			ignore:       "# nothing to see here",
			excludePaths: []string{".kosli_ignore"},
			wantWarning:  true,
		},
		{
			// Silent because nothing is excluded at all: filepathx resolves this to
			// "dir//.kosli_ignore", which the walk never emits. Give ** patterns a
			// path the walk emits and this case moves to wantWarning: true.
			name:         "a ** spelling that never excludes anything stays silent",
			ignore:       "logs",
			excludePaths: []string{"**/.kosli_ignore"},
			wantWarning:  false,
		},
		{
			name:   "no flag exclusion at all",
			ignore: "logs",
		},
		{
			name:         "excluding an unrelated path",
			ignore:       "logs",
			excludePaths: []string{"app"},
		},
	} {
		suite.Run(t.name, func() {
			dir := suite.createDirWithFiles(fmt.Sprintf("warn%d", i), map[string]string{
				"app/index.js": "console.log(1)",
				"logs/a.log":   "noise",
			})
			suite.createFileWithContent(filepath.Join(dir, ".kosli_ignore"), t.ignore)

			var warnings bytes.Buffer
			_, err := DirSha256(dir, t.excludePaths, logger.NewLogger(io.Discard, &warnings, false))
			require.NoError(suite.T(), err)

			if t.wantWarning {
				assert.Contains(suite.T(), warnings.String(), warning)
			} else {
				assert.NotContains(suite.T(), warnings.String(), warning)
			}
		})
	}
}

func (suite *DigestTestSuite) TestIgnoreFilePathInTree() {
	for i, t := range []struct {
		name       string
		ignoreName string
	}{
		{name: "the usual lower-case name", ignoreName: ".kosli_ignore"},
		{name: "an upper-case name on a case-insensitive filesystem", ignoreName: ".KOSLI_IGNORE"},
		{name: "a mixed-case name on a case-insensitive filesystem", ignoreName: ".Kosli_Ignore"},
	} {
		suite.Run(t.name, func() {
			// The directory name must not embed ignoreName: on a case-insensitive
			// filesystem the three cases would collapse into one directory.
			dir := suite.createDirWithFiles(fmt.Sprintf("tree%d", i), map[string]string{"app.js": "app"})
			suite.createFileWithContent(filepath.Join(dir, t.ignoreName), "logs")

			if _, err := os.Stat(filepath.Join(dir, ".kosli_ignore")); err != nil {
				suite.T().Skipf("filesystem is case-sensitive, so %s is not the ignore file", t.ignoreName)
			}

			got, err := ignoreFilePathInTree(dir)
			require.NoError(suite.T(), err)
			assert.Equal(suite.T(), filepath.Join(dir, t.ignoreName), got)
		})
	}
}

// The only test that fails under a folded-first mutation.
func (suite *DigestTestSuite) TestIgnoreFilePathInTreeWithBothSpellings() {
	dir := suite.createDirWithFiles("both", map[string]string{"app.js": "app"})
	suite.createFileWithContent(filepath.Join(dir, ".KOSLI_IGNORE"), "")
	suite.createFileWithContent(filepath.Join(dir, ".kosli_ignore"), "logs")
	suite.requireCaseSensitiveDir(dir)

	got, err := ignoreFilePathInTree(dir)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), filepath.Join(dir, ".kosli_ignore"), got)
}

// Documents rather than pins: a folded-first mutation locates the empty decoy, so
// no rules are read and the trees differ regardless.
func (suite *DigestTestSuite) TestDirSha256IgnoreFileCannotHideItselfBesideACaseVariant() {
	baseline := map[string]string{"app/index.js": "console.log(1)"}
	approved := suite.createDirWithFiles("approved-decoy", baseline)
	suite.createFileWithContent(filepath.Join(approved, ".KOSLI_IGNORE"), "")
	suite.createFileWithContent(filepath.Join(approved, ".kosli_ignore"), ".kosli_ignore")
	suite.requireCaseSensitiveDir(approved)

	deployed := suite.createDirWithFiles("deployed-decoy", baseline)
	suite.createFileWithContent(filepath.Join(deployed, ".KOSLI_IGNORE"), "")
	suite.createFileWithContent(filepath.Join(deployed, ".kosli_ignore"), ".kosli_ignore\napp/backdoor.js")
	suite.createFileWithContent(filepath.Join(deployed, "app/backdoor.js"), "BACKDOOR")

	approvedSha, err := DirSha256(approved, nil, logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	deployedSha, err := DirSha256(deployed, nil, logger.NewStandardLogger())
	require.NoError(suite.T(), err)

	assert.NotEqual(suite.T(), approvedSha, deployedSha,
		"the added file is invisible to the fingerprint")
}

func (suite *DigestTestSuite) requireCaseSensitiveDir(dir string) {
	entries, err := os.ReadDir(dir)
	require.NoError(suite.T(), err)
	found := 0
	for _, entry := range entries {
		if strings.EqualFold(entry.Name(), ".kosli_ignore") {
			found++
		}
	}
	if found < 2 {
		suite.T().Skip("filesystem is case-insensitive, so the two spellings are one file")
	}
}

func (suite *DigestTestSuite) TestIgnoreFilePathInTreeIgnoresADirectory() {
	dir := suite.createDirWithFiles("dir-named-ignore", map[string]string{"app.js": "app"})
	require.NoError(suite.T(), os.Mkdir(filepath.Join(dir, ".kosli_ignore"), 0777))
	suite.createFileWithContent(filepath.Join(dir, ".kosli_ignore", "inside.txt"), "secret")

	path, err := ignoreFilePathInTree(dir)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "", path)

	withDir, err := DirSha256(dir, []string{".kosli_ignore"}, logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	withoutDir, err := DirSha256(
		suite.createDirWithFiles("no-dir-named-ignore", map[string]string{"app.js": "app"}),
		nil, logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), withoutDir, withDir)
}

func (suite *DigestTestSuite) TestIgnoreFilePathInTreeWithoutIgnoreFile() {
	dir := suite.createDirWithFiles("no-ignore", map[string]string{"app.js": "app"})
	got, err := ignoreFilePathInTree(dir)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "", got)
}

// An alias is a distinct walk entry, so excluding it leaves the ignore file in the
// digest.
func (suite *DigestTestSuite) TestDirSha256IgnoreFileAliasesStayExcludable() {
	baseline := map[string]string{"app/index.js": "console.log(1)"}
	for i, t := range []struct {
		name string
		link func(ignorePath, aliasPath string) error
	}{
		{name: "a hard link of the ignore file can be excluded", link: os.Link},
		{
			name: "a symlink to the ignore file can be excluded",
			link: func(ignorePath, aliasPath string) error { return os.Symlink(ignorePath, aliasPath) },
		},
	} {
		suite.Run(t.name, func() {
			withAlias := suite.createDirWithFiles(fmt.Sprintf("with-alias%d", i), baseline)
			suite.createFileWithContent(filepath.Join(withAlias, ".kosli_ignore"), "alias")
			require.NoError(suite.T(), t.link(
				filepath.Join(withAlias, ".kosli_ignore"),
				filepath.Join(withAlias, "alias"),
			))

			noAlias := suite.createDirWithFiles(fmt.Sprintf("no-alias%d", i), baseline)
			suite.createFileWithContent(filepath.Join(noAlias, ".kosli_ignore"), "alias")

			withAliasSha, err := DirSha256(withAlias, nil, logger.NewStandardLogger())
			require.NoError(suite.T(), err)
			noAliasSha, err := DirSha256(noAlias, nil, logger.NewStandardLogger())
			require.NoError(suite.T(), err)

			assert.Equal(suite.T(), noAliasSha, withAliasSha,
				"the alias is excluded, leaving the same digest as a tree without it")
		})
	}
}

// Only the root ignore file is read as an exclusion list.
func (suite *DigestTestSuite) TestDirSha256NestedIgnoreFileIsExcludable() {
	baseline := map[string]string{
		"app.js":          "app",
		"vendor/lib/v.js": "v",
		".kosli_ignore":   "vendor/**/.kosli_ignore",
	}
	withNested := suite.createDirWithFiles("with-nested", baseline)
	suite.createFileWithContent(filepath.Join(withNested, "vendor/lib/.kosli_ignore"), "nested-rules")
	withoutNested := suite.createDirWithFiles("without-nested", baseline)

	withNestedSha, err := DirSha256(withNested, nil, logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	withoutNestedSha, err := DirSha256(withoutNested, nil, logger.NewStandardLogger())
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), withoutNestedSha, withNestedSha,
		"a nested .kosli_ignore is an ordinary file and the root list may exclude it")
}

// A case-insensitive filesystem lets the ignore file be named in one case and refer
// to itself in another.
func (suite *DigestTestSuite) TestDirSha256IgnoreFileCannotHideItselfCaseInsensitiveFS() {
	for i, t := range []struct {
		name       string
		ignoreName string
		ignore     string
	}{
		{
			name:       "an upper-case ignore file naming itself cannot hide an added file",
			ignoreName: ".KOSLI_IGNORE",
			ignore:     ".KOSLI_IGNORE\napp/backdoor.js",
		},
		{
			name:       "a mixed-case ignore file naming itself cannot hide an added file",
			ignoreName: ".Kosli_Ignore",
			ignore:     ".Kosli_Ignore\napp/backdoor.js",
		},
		{
			name:       "an upper-case glob cannot hide an upper-case ignore file",
			ignoreName: ".KOSLI_IGNORE",
			ignore:     ".KOSLI*\napp/backdoor.js",
		},
	} {
		suite.Run(t.name, func() {
			baseline := map[string]string{
				"app/index.js":    "console.log(1)",
				"app/lib/util.js": "exports.x = 1",
			}
			approved := suite.createDirWithFiles(fmt.Sprintf("approved%d", i), baseline)
			deployed := suite.createDirWithFiles(fmt.Sprintf("deployed%d", i), baseline)
			suite.createFileWithContent(filepath.Join(deployed, "app/backdoor.js"), "BACKDOOR")
			suite.createFileWithContent(filepath.Join(deployed, t.ignoreName), t.ignore)

			if _, err := os.Stat(filepath.Join(deployed, ".kosli_ignore")); err != nil {
				suite.T().Skipf("filesystem is case-sensitive, so %s is not read as an ignore file", t.ignoreName)
			}

			approvedSha, err := DirSha256(approved, nil, logger.NewStandardLogger())
			require.NoError(suite.T(), err)
			deployedSha, err := DirSha256(deployed, nil, logger.NewStandardLogger())
			require.NoError(suite.T(), err)

			assert.NotEqual(suite.T(), approvedSha, deployedSha,
				"the added file is invisible to the fingerprint")
		})
	}
}

func (suite *DigestTestSuite) createDirWithFiles(name string, files map[string]string) string {
	dirPath := filepath.Join(suite.tmpDir, name)
	err := os.MkdirAll(dirPath, 0777)
	require.NoErrorf(suite.T(), err, "error creating test dir %s", dirPath)
	for path, content := range files {
		suite.createFileWithContent(filepath.Join(dirPath, path), content)
	}
	return dirPath
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

// TestValidateDigestErrorMessage guards the exact error message. Command tests in
// cmd/kosli assert this text (including the regex pattern) in their golden output.
func (suite *DigestTestSuite) TestValidateDigestErrorMessage() {
	err := ValidateDigest("xxxx")
	require.EqualError(suite.T(),
		err,
		"xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$")
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
			client, clientErr := requests.NewKosliClient("", 1, false, logger.NewStandardLogger())
			require.NoErrorf(suite.T(), clientErr, "TestRemoteDockerImageSha256: client construction must not fail")
			actual, err := RemoteDockerImageSha256(client, t.localImageName, t.localImageTag, "http://localhost:5001/v2", "secret")
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

// A stopped scan must not pass off the entries read so far as the file's rules.
func (suite *DigestTestSuite) TestExcludePathsFromFileErrorsOnAStoppedScan() {
	dir := suite.createDirWithFiles("stopped-scan", map[string]string{})
	path := filepath.Join(dir, ".kosli_ignore")
	suite.createFileWithContent(path, "#"+strings.Repeat("z", bufio.MaxScanTokenSize)+"\nlogs\n")

	paths, err := excludePathsFromFile(path)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), paths)
}

func (suite *DigestTestSuite) TestGetExcludePathsFromIgnoreFile() {
	type want struct {
		expectError  bool
		excludePaths []string
	}
	const MISSING_FILE_NAME = "NO_CREATE.ignore"
	for _, t := range []struct {
		name           string
		ignoreFileName string
		content        string
		want           want
	}{
		{
			name:           "missing file should return []",
			ignoreFileName: MISSING_FILE_NAME,
			want: want{
				expectError:  false,
				excludePaths: []string{},
			},
		},
		{
			name:           "empty file should return []",
			ignoreFileName: "empty.ignore",
			content:        ``,
			want: want{
				expectError:  false,
				excludePaths: []string{},
			},
		},
		{
			name:           "a single line ignore file",
			ignoreFileName: "empty.ignore",
			content:        `logs`,
			want: want{
				expectError:  false,
				excludePaths: []string{"logs"},
			},
		},
		{
			name:           "a multi line ignore file",
			ignoreFileName: "multi.ignore",
			content: `logs
*/logs`,
			want: want{
				expectError:  false,
				excludePaths: []string{"logs", "*/logs"},
			},
		},
		{
			name:           "an ignore file with blank lines",
			ignoreFileName: "multi-with-blank.ignore",
			content: `logs
    
	
*/logs
`,
			want: want{
				expectError:  false,
				excludePaths: []string{"logs", "*/logs"},
			},
		},
		{
			name:           "a commented ignore file",
			ignoreFileName: "multi.ignore",
			content: `logs
# a line comment
*/logs # an end of line comment`,
			want: want{
				expectError:  false,
				excludePaths: []string{"logs", "*/logs"},
			},
		},
	} {
		suite.Run(t.name, func() {
			assert.False(suite.T(), t.ignoreFileName == "", "ignoreFileName cannot be empty string")
			ignoreFilePath := filepath.Join(suite.tmpDir, t.ignoreFileName)
			if t.ignoreFileName != MISSING_FILE_NAME {
				testFile, err := os.Create(ignoreFilePath)
				require.NoErrorf(suite.T(), err, "error creating test file %s: %s", t.ignoreFileName, err)

				_, err = testFile.Write([]byte(t.content))
				require.NoErrorf(suite.T(), err, "error writing content to test file %s: %s", t.ignoreFileName, err)
			}

			actual, err := excludePathsFromFile(ignoreFilePath)
			if t.want.expectError {
				require.Errorf(suite.T(), err, "TestGetExcludePathsFromIgnoreFile: error was expected: %s", err)
			} else {
				require.NoErrorf(suite.T(), err, "TestGetExcludePathsFromIgnoreFile: error was NOT expected: %s", err)
				assert.Equal(suite.T(), t.want.excludePaths, actual, fmt.Sprintf("TestGetExcludePathsFromIgnoreFile: want %s -- got %s", t.want.excludePaths, actual))
			}
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDigestTestSuite(t *testing.T) {
	suite.Run(t, new(DigestTestSuite))
}

func BenchmarkValidateDigest(b *testing.B) {
	sha256 := "db40d79b3a15b17ee9fcc2f49aa73736e0073de6b5a35c459268bb9a31e55139"
	for b.Loop() {
		if err := ValidateDigest(sha256); err != nil {
			b.Fatal(err)
		}
	}
}
