package digest

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/utils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type VirtualDirTestSuite struct {
	suite.Suite
	tmpDir string
}

func (suite *VirtualDirTestSuite) SetupTest() {
	suite.tmpDir = suite.T().TempDir()
}

// TestVirtualDirSha256MatchesDirSha256 is the test that matters: for each tree,
// materialise it on disk, fingerprint it with DirSha256, then fingerprint the
// same (path, content sha256) pairs with VirtualDirSha256 and require the two
// to be identical. Anything VirtualDirSha256 gets wrong about walk order, name
// hashing or nesting shows up here as a mismatch.
func (suite *VirtualDirTestSuite) TestVirtualDirSha256MatchesDirSha256() {
	for _, t := range []struct {
		name  string
		files map[string]string // path relative to the tree root -> content
	}{
		{
			name:  "a single file at the root",
			files: map[string]string{"README.md": "# readme\n"},
		},
		{
			name:  "two files at the root",
			files: map[string]string{"README.md": "# readme\n", "notes.txt": "notes\n"},
		},
		{
			name: "nested directories",
			files: map[string]string{
				"README.md":                  "# readme\n",
				"dummy/dummy_2/template.yml": "key: value\n",
				"dummy/other.txt":            "other\n",
			},
		},
		{
			// '.' (0x2E) sorts before '/' (0x2F), so a flat sort of the keys
			// gives a.txt, a/z -- while WalkDir gives a, a/z, a.txt. Sorting
			// the key list instead of the tree fails exactly here.
			name: "a dir name sorting between two file names",
			files: map[string]string{
				"a.txt": "a\n",
				"a/z":   "z\n",
				"b.txt": "b\n",
			},
		},
		{
			name:  "a deep single-child chain",
			files: map[string]string{"a/b/c/d/e/f.txt": "deep\n"},
		},
		{
			name: "dot-prefixed and unicode names",
			files: map[string]string{
				".hidden":      "hidden\n",
				"ünïcode.txt":  "unicode\n",
				"dir/.keep":    "",
				"dir/naïve.md": "naive\n",
			},
		},
		{
			name:  "an empty file",
			files: map[string]string{"empty.txt": "", "other.txt": "x\n"},
		},
		{
			name: "many files across several levels",
			files: map[string]string{
				"a.txt": "a\n", "b.txt": "b\n", "c/d.txt": "d\n", "c/e.txt": "e\n",
				"c/f/g.txt": "g\n", "c/f/h.txt": "h\n", "i/j.txt": "j\n",
			},
		},
	} {
		suite.Run(t.name, func() {
			root := suite.T().TempDir()
			virtualFiles := make([]VirtualFile, 0, len(t.files))
			for path, content := range t.files {
				suite.createFile(filepath.Join(root, filepath.FromSlash(path)), content)
				virtualFiles = append(virtualFiles, VirtualFile{
					Path:   path,
					Sha256: sha256OfString(content),
				})
			}

			want, err := DirSha256(root, []string{}, logger.NewStandardLogger())
			require.NoError(suite.T(), err)

			got, err := VirtualDirSha256(virtualFiles, logger.NewStandardLogger())
			require.NoError(suite.T(), err)

			require.Equal(suite.T(), want, got,
				"VirtualDirSha256 should equal DirSha256 of the same tree")
		})
	}
}

// TestVirtualDirSha256IgnoresInputOrder pins that the result depends on the tree,
// not on the order S3 happened to list the objects in.
func (suite *VirtualDirTestSuite) TestVirtualDirSha256IgnoresInputOrder() {
	files := []VirtualFile{
		{Path: "c/f/g.txt", Sha256: sha256OfString("g")},
		{Path: "a.txt", Sha256: sha256OfString("a")},
		{Path: "c/d.txt", Sha256: sha256OfString("d")},
		{Path: "b.txt", Sha256: sha256OfString("b")},
	}
	reversed := make([]VirtualFile, len(files))
	for i, f := range files {
		reversed[len(files)-1-i] = f
	}

	first, err := VirtualDirSha256(files, logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	second, err := VirtualDirSha256(reversed, logger.NewStandardLogger())
	require.NoError(suite.T(), err)

	require.Equal(suite.T(), first, second)
}

func (suite *VirtualDirTestSuite) TestVirtualDirSha256Errors() {
	validSha := sha256OfString("x")
	for _, t := range []struct {
		name       string
		files      []VirtualFile
		wantErrMsg string
	}{
		{
			name:       "no files",
			files:      []VirtualFile{},
			wantErrMsg: "no files",
		},
		{
			name: "a duplicate path",
			files: []VirtualFile{
				{Path: "a.txt", Sha256: validSha},
				{Path: "a.txt", Sha256: validSha},
			},
			wantErrMsg: "duplicate path",
		},
		{
			name: "a path used as both file and directory",
			files: []VirtualFile{
				{Path: "a", Sha256: validSha},
				{Path: "a/b", Sha256: validSha},
			},
			wantErrMsg: "both a file and a directory",
		},
		{
			name: "a path used as both directory and file",
			files: []VirtualFile{
				{Path: "a/b", Sha256: validSha},
				{Path: "a", Sha256: validSha},
			},
			wantErrMsg: "both a file and a directory",
		},
		{
			name:       "an empty path segment",
			files:      []VirtualFile{{Path: "a//b", Sha256: validSha}},
			wantErrMsg: "not a clean relative path",
		},
		{
			name:       "a parent directory segment",
			files:      []VirtualFile{{Path: "../evil", Sha256: validSha}},
			wantErrMsg: "not a clean relative path",
		},
		{
			name:       "a current directory segment",
			files:      []VirtualFile{{Path: "a/./b", Sha256: validSha}},
			wantErrMsg: "not a clean relative path",
		},
		{
			name:       "a leading slash",
			files:      []VirtualFile{{Path: "/a.txt", Sha256: validSha}},
			wantErrMsg: "not a clean relative path",
		},
		{
			name:       "a trailing slash",
			files:      []VirtualFile{{Path: "a/", Sha256: validSha}},
			wantErrMsg: "not a clean relative path",
		},
		{
			name:       "an empty path",
			files:      []VirtualFile{{Path: "", Sha256: validSha}},
			wantErrMsg: "not a clean relative path",
		},
		{
			name:       "an invalid sha256",
			files:      []VirtualFile{{Path: "a.txt", Sha256: "not-a-digest"}},
			wantErrMsg: "not a valid SHA256 fingerprint",
		},
		{
			name:       "an uppercase sha256",
			files:      []VirtualFile{{Path: "a.txt", Sha256: strings.ToUpper(validSha)}},
			wantErrMsg: "not a valid SHA256 fingerprint",
		},
	} {
		suite.Run(t.name, func() {
			_, err := VirtualDirSha256(t.files, logger.NewStandardLogger())
			require.Error(suite.T(), err)
			require.Contains(suite.T(), err.Error(), t.wantErrMsg)
		})
	}
}

func (suite *VirtualDirTestSuite) TestSingleVirtualFile() {
	validSha := sha256OfString("x")
	for _, t := range []struct {
		name     string
		files    []VirtualFile
		wantOK   bool
		wantBase string
	}{
		{
			name:     "one file at the root",
			files:    []VirtualFile{{Path: "README.md", Sha256: validSha}},
			wantOK:   true,
			wantBase: "README.md",
		},
		{
			name:     "one file nested under prefixes keeps its base name",
			files:    []VirtualFile{{Path: "dummy/dummy_2/template.yml", Sha256: validSha}},
			wantOK:   true,
			wantBase: "template.yml",
		},
		{
			name: "two files is not a single file",
			files: []VirtualFile{
				{Path: "a.txt", Sha256: validSha},
				{Path: "b.txt", Sha256: validSha},
			},
			wantOK: false,
		},
		{
			name:   "no files is not a single file",
			files:  []VirtualFile{},
			wantOK: false,
		},
	} {
		suite.Run(t.name, func() {
			file, ok := SingleVirtualFile(t.files)
			require.Equal(suite.T(), t.wantOK, ok)
			if t.wantOK {
				require.Equal(suite.T(), t.wantBase, file.Name())
			}
		})
	}
}

// TestSingleFileMatchesFileSha256 pins the equivalence the aws package relies on:
// a one-object snapshot is fingerprinted as that file's content digest, exactly
// as content mode does via containsSingleFile + FileSha256.
func (suite *VirtualDirTestSuite) TestSingleFileMatchesFileSha256() {
	content := "the only object\n"
	path := filepath.Join(suite.tmpDir, "only.txt")
	suite.createFile(path, content)

	want, err := FileSha256(path, logger.NewStandardLogger())
	require.NoError(suite.T(), err)

	file, ok := SingleVirtualFile([]VirtualFile{{Path: "nested/only.txt", Sha256: sha256OfString(content)}})
	require.True(suite.T(), ok)
	require.Equal(suite.T(), want, file.Sha256)
}

// createFile writes content to path, creating parent directories as needed.
func (suite *VirtualDirTestSuite) createFile(path, content string) {
	suite.T().Helper()
	require.NoError(suite.T(), utils.CreateFileWithContent(path, content))
}

func TestVirtualDirTestSuite(t *testing.T) {
	suite.Run(t, new(VirtualDirTestSuite))
}
