package digest

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/utils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type VirtualIgnoreTestSuite struct {
	suite.Suite
}

// ignoreTestTree is shaped to reach every glob behaviour DirSha256 has on disk:
// a directory and a file sharing a stem, the same name at several depths, a
// nested .kosli_ignore, and a directory whose content can be excluded while the
// directory itself stays.
var ignoreTestTree = map[string]string{
	"app.js":                   "app",
	"app.log":                  "log at the root",
	"notes.txt":                "notes",
	"logs/file1":               "c1",
	"logs/deep/file2":          "c2",
	"nested-dir/file1":         "n1",
	"nested-dir/logs/log.txt":  "nl",
	"vendor/lib/v.js":          "v",
	"vendor/lib/.kosli_ignore": "nested-rules",
	"a/x":                      "ax",
	"a/b/x":                    "abx",
	"a/b/c/x":                  "abcx",
}

// TestMatchesDirSha256 is the test that matters. For each rule set, the tree
// plus a root .kosli_ignore holding the rules is materialised on disk and
// fingerprinted with DirSha256; the same files and rules go through
// VirtualDirSha256 and must give the identical digest. Rows marked hasEffect
// also require the rules to have changed the digest, so a row cannot pass
// because both sides ignored the rules.
func (suite *VirtualIgnoreTestSuite) TestMatchesDirSha256() {
	for _, t := range []struct {
		name      string
		ignore    string
		hasEffect bool
	}{
		{name: "no rules", ignore: ""},
		{name: "only comments and blanks", ignore: "# nothing\n\n   \n"},
		{name: "a directory by name", ignore: "logs", hasEffect: true},
		{name: "a file by name", ignore: "app.js", hasEffect: true},
		{name: "a directory at depth one", ignore: "*/logs", hasEffect: true},
		{name: "the content of a directory but not the directory", ignore: "logs/*", hasEffect: true},
		{name: "a suffix at the root", ignore: "*.log", hasEffect: true},
		{name: "a suffix at any depth", ignore: "**/*.log", hasEffect: true},
		{name: "a literal name at any depth", ignore: "**/x", hasEffect: true},
		{name: "a literal name at any depth under a prefix", ignore: "a/**/x", hasEffect: true},
		{name: "a directory at any depth", ignore: "**/logs", hasEffect: true},
		{name: "a bare double star", ignore: "**", hasEffect: true},
		// filepathx appends ".log" to every existing path, so this matches only an
		// "x.log" whose sibling "x" also exists. Nothing here, on either side.
		{name: "a double star glued to a suffix", ignore: "**.log"},
		{name: "a nested ignore file under a prefix", ignore: "vendor/**/.kosli_ignore", hasEffect: true},
		{name: "a trailing slash", ignore: "logs/", hasEffect: true},
		{name: "a leading slash", ignore: "/logs", hasEffect: true},
		{name: "a leading dot segment", ignore: "./logs", hasEffect: true},
		{name: "a doubled slash", ignore: "nested-dir//logs", hasEffect: true},
		{name: "a parent segment that leaves the tree", ignore: "../logs"},
		{name: "a rule that matches nothing", ignore: "does-not-exist"},
		{name: "a star alone", ignore: "*", hasEffect: true},
		{name: "a character class", ignore: "app.[jl]?", hasEffect: true},
		{name: "a question mark", ignore: "a/?", hasEffect: true},
		{name: "an escaped star is a literal", ignore: `app\*`},
		{name: "several rules with a comment", ignore: "logs\n# keep vendor\n*/logs\napp.log  # trailing comment\n", hasEffect: true},
		{name: "the ignore file itself", ignore: ".kosli_ignore"},
		{name: "a glob matching the ignore file", ignore: "*ignore*"},
		{name: "a dotted glob matching the ignore file", ignore: ".kosli*"},
		{name: "a double star matching the ignore file", ignore: "**/.kosli_ignore", hasEffect: true},
	} {
		suite.Run(t.name, func() {
			root := suite.T().TempDir()
			files := suite.materialise(root, ignoreTestTree, t.ignore)

			want, err := DirSha256(root, nil, logger.NewStandardLogger())
			require.NoError(suite.T(), err)

			rules, err := ParseIgnoreRules(strings.NewReader(t.ignore))
			require.NoError(suite.T(), err)
			got, err := VirtualDirSha256(files, rules, logger.NewStandardLogger())
			require.NoError(suite.T(), err)
			require.Equal(suite.T(), want, got, "VirtualDirSha256 must equal DirSha256 with rules %q", t.ignore)

			noRules, err := VirtualDirSha256(files, nil, logger.NewStandardLogger())
			require.NoError(suite.T(), err)
			if t.hasEffect {
				require.NotEqual(suite.T(), noRules, got, "rules %q were expected to change the digest", t.ignore)
			} else {
				require.Equal(suite.T(), noRules, got, "rules %q were expected to leave the digest unchanged", t.ignore)
			}
		})
	}
}

// A malformed pattern fails DirSha256, so it must fail the virtual digest too
// rather than silently excluding nothing.
func (suite *VirtualIgnoreTestSuite) TestMalformedRuleIsAnErrorOnBothSides() {
	root := suite.T().TempDir()
	files := suite.materialise(root, ignoreTestTree, "[")

	_, err := DirSha256(root, nil, logger.NewStandardLogger())
	require.Error(suite.T(), err)

	_, err = VirtualDirSha256(files, []string{"["}, logger.NewStandardLogger())
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "[")
}

// Mirrors TestDirSha256IgnoreFileCannotHideItself: an ignore file that lists
// itself cannot hide an added file, because the file's own content stays in the
// digest.
func (suite *VirtualIgnoreTestSuite) TestIgnoreFileCannotHideItself() {
	baseline := map[string]string{
		"app/index.js":    "console.log(1)",
		"app/lib/util.js": "exports.x = 1",
	}
	approvedFiles := virtualFilesFor(baseline, "")
	approved, err := VirtualDirSha256(approvedFiles, nil, logger.NewStandardLogger())
	require.NoError(suite.T(), err)

	for _, ignore := range []string{
		".kosli_ignore\napp/backdoor.js",
		"*ignore*\napp/backdoor.js",
		".kosli*\napp/backdoor.js",
		"**/.kosli_ignore\napp/backdoor.js",
		"**\napp/backdoor.js",
	} {
		suite.Run(ignore, func() {
			deployed := map[string]string{}
			for k, v := range baseline {
				deployed[k] = v
			}
			deployed["app/backdoor.js"] = "BACKDOOR"
			rules, err := ParseIgnoreRules(strings.NewReader(ignore))
			require.NoError(suite.T(), err)

			got, err := VirtualDirSha256(virtualFilesFor(deployed, ignore), rules, logger.NewStandardLogger())
			require.NoError(suite.T(), err)
			require.NotEqual(suite.T(), approved, got, "the added file is invisible to the fingerprint")
		})
	}
}

// Only the root ignore file is protected; a nested one is an ordinary file the
// root list may exclude.
func (suite *VirtualIgnoreTestSuite) TestNestedIgnoreFileIsExcludable() {
	rules := "vendor/**/.kosli_ignore"
	withNested := virtualFilesFor(map[string]string{
		"app.js":                   "app",
		"vendor/lib/v.js":          "v",
		"vendor/lib/.kosli_ignore": "nested-rules",
	}, rules)
	withoutNested := virtualFilesFor(map[string]string{
		"app.js":          "app",
		"vendor/lib/v.js": "v",
	}, rules)

	parsed, err := ParseIgnoreRules(strings.NewReader(rules))
	require.NoError(suite.T(), err)
	a, err := VirtualDirSha256(withNested, parsed, logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	b, err := VirtualDirSha256(withoutNested, parsed, logger.NewStandardLogger())
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), b, a)
}

func (suite *VirtualIgnoreTestSuite) TestParseIgnoreRules() {
	for _, t := range []struct {
		name  string
		input string
		want  []string
	}{
		{name: "empty input", input: "", want: []string{}},
		{name: "one rule", input: "logs", want: []string{"logs"}},
		{name: "blank lines and whitespace are dropped", input: "\n  logs  \n\n\t*.log\t\n", want: []string{"logs", "*.log"}},
		{name: "comment lines are dropped", input: "# all logs\nlogs\n# done", want: []string{"logs"}},
		{name: "a trailing comment is stripped", input: "logs # noisy", want: []string{"logs"}},
		{name: "a comment with no rule before it is dropped", input: "   # only a comment", want: []string{}},
		{name: "windows line endings", input: "logs\r\n*.log\r\n", want: []string{"logs", "*.log"}},
	} {
		suite.Run(t.name, func() {
			got, err := ParseIgnoreRules(strings.NewReader(t.input))
			require.NoError(suite.T(), err)
			require.Equal(suite.T(), t.want, got)
		})
	}
}

// ParseIgnoreRules and excludePathsFromFile must agree, since DirSha256 reads
// the file while the virtual digest is handed the same bytes.
func (suite *VirtualIgnoreTestSuite) TestParseIgnoreRulesAgreesWithTheFileReader() {
	content := "logs\n# comment\n*/logs # trailing\n\n  app.log\n"
	path := filepath.Join(suite.T().TempDir(), ".kosli_ignore")
	require.NoError(suite.T(), utils.CreateFileWithContent(path, content))

	fromFile, err := excludePathsFromFile(path)
	require.NoError(suite.T(), err)
	fromReader, err := ParseIgnoreRules(strings.NewReader(content))
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), fromFile, fromReader)
}

// materialise writes the tree and, when ignore is non-empty, a root
// .kosli_ignore holding it; it returns the matching virtual files.
func (suite *VirtualIgnoreTestSuite) materialise(root string, tree map[string]string, ignore string) []VirtualFile {
	suite.T().Helper()
	for p, content := range tree {
		require.NoError(suite.T(), utils.CreateFileWithContent(filepath.Join(root, filepath.FromSlash(p)), content))
	}
	if ignore != "" {
		require.NoError(suite.T(), utils.CreateFileWithContent(filepath.Join(root, ignoreFileName), ignore))
	}
	return virtualFilesFor(tree, ignore)
}

// virtualFilesFor returns the (path, sha256) pairs for tree plus, when ignore
// is non-empty, a root .kosli_ignore with that content, in a fixed order.
func virtualFilesFor(tree map[string]string, ignore string) []VirtualFile {
	paths := make([]string, 0, len(tree)+1)
	for p := range tree {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	files := make([]VirtualFile, 0, len(paths)+1)
	for _, p := range paths {
		files = append(files, VirtualFile{Path: p, Sha256: sha256OfString(tree[p])})
	}
	if ignore != "" {
		files = append(files, VirtualFile{Path: ignoreFileName, Sha256: sha256OfString(ignore)})
	}
	return files
}

func TestVirtualIgnoreTestSuite(t *testing.T) {
	suite.Run(t, new(VirtualIgnoreTestSuite))
}
