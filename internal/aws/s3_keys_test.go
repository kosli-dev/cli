package aws

import (
	"fmt"
	"math/rand"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type S3KeysTestSuite struct {
	suite.Suite
}

// Accepted keys land on exactly the path filepath.Join(dir, key) produced on
// Linux before objects stopped being written under their own names, so every
// bucket that fingerprints today keeps its fingerprint.
func (suite *S3KeysTestSuite) TestVirtualPathForS3Key() {
	longComponent := strings.Repeat("n", 300)
	for _, t := range []struct {
		name       string
		key        string
		wantPath   string
		wantErrMsg string
	}{
		{name: "an ordinary nested key", key: "protected/release.bin", wantPath: "protected/release.bin"},
		{name: "a plain filename", key: "a.txt", wantPath: "a.txt"},
		{name: "a short nested key", key: "a/z", wantPath: "a/z"},
		{name: "a dotfile", key: ".kosli_ignore", wantPath: ".kosli_ignore"},
		{name: "a key with spaces", key: "file with spaces.txt", wantPath: "file with spaces.txt"},
		{name: "a key with punctuation", key: "weird!*'().txt", wantPath: "weird!*'().txt"},
		{name: "a unicode key", key: "ünïcödé/файл.txt", wantPath: "ünïcödé/файл.txt"},
		{name: "a dot followed by a space is a literal name", key: ". ", wantPath: ". "},
		{name: "a name that merely starts with two dots", key: "..hidden", wantPath: "..hidden"},
		{name: "a leading slash is trimmed", key: "/etc/passwd", wantPath: "etc/passwd"},
		{name: "doubled leading slashes are trimmed", key: "//x", wantPath: "x"},
		{name: "a leading dot segment is folded", key: "./a.txt", wantPath: "a.txt"},
		{name: "a doubled interior slash is folded", key: "a//b", wantPath: "a/b"},
		{name: "a dot segment is folded", key: "a/./b", wantPath: "a/b"},
		// Nothing is created under these names, so the operating system's rules
		// about them no longer apply.
		{name: "a backslash is a literal character", key: `dir\file.txt`, wantPath: `dir\file.txt`},
		{name: "a leading backslash is a literal character", key: `\evil.txt`, wantPath: `\evil.txt`},
		{name: "a backslash-separated \"..\" is a literal name", key: `uploads/user-a/..\..\x`, wantPath: `uploads/user-a/..\..\x`},
		{name: "a reserved Windows name", key: "CON", wantPath: "CON"},
		{name: "a drive-looking segment", key: "C:evil", wantPath: "C:evil"},
		{name: "a colon segment", key: "a:b", wantPath: "a:b"},
		{name: "three dots", key: "a/.../b", wantPath: "a/.../b"},
		{name: "two dots and a space", key: "a/.. /x", wantPath: "a/.. /x"},
		{name: "a component longer than any filesystem allows", key: "dir/" + longComponent, wantPath: "dir/" + longComponent},
		{name: "a key of the maximum S3 length", key: strings.Repeat("s/", 511) + "ab", wantPath: strings.Repeat("s/", 511) + "ab"},
		// Rejected: these cannot be entries of any directory tree.
		{name: "a traversing key", key: "uploads/user-a/../../protected/release.bin", wantErrMsg: `contains a ".." segment`},
		{name: "a bare \"..\"", key: "..", wantErrMsg: `contains a ".." segment`},
		{name: "a leading \"..\"", key: "../x", wantErrMsg: `contains a ".." segment`},
		{name: "a trailing \"..\"", key: "a/..", wantErrMsg: `contains a ".." segment`},
		{name: "a \"..\" that would fold onto a sibling", key: "a/../b", wantErrMsg: `contains a ".." segment`},
		{name: "an empty key", key: "", wantErrMsg: "names no file"},
		{name: "a bare slash", key: "/", wantErrMsg: "names no file"},
		{name: "doubled slashes with nothing else", key: "//", wantErrMsg: "names no file"},
		{name: "a bare dot", key: ".", wantErrMsg: "names no file"},
		{name: "a dot slash dot", key: "./.", wantErrMsg: "names no file"},
		{name: "a folder marker", key: "dir/", wantErrMsg: "folder marker"},
		{name: "a folder marker with a dot segment", key: "a/./", wantErrMsg: "folder marker"},
	} {
		suite.Run(t.name, func() {
			got, err := virtualPathForS3Key(t.key)
			if t.wantErrMsg != "" {
				require.Error(suite.T(), err)
				require.Contains(suite.T(), err.Error(), fmt.Sprintf("object key [%s]", t.key))
				require.Contains(suite.T(), err.Error(), t.wantErrMsg)
				return
			}
			require.NoError(suite.T(), err)
			require.Equal(suite.T(), t.wantPath, got)
		})
	}
}

// Any key S3 accepts is representable unless it holds a ".." segment, whatever
// operating system runs the snapshot. Components are drawn from an alphabet with
// no '/' and are never exactly "..", so every generated key must be accepted.
func (suite *S3KeysTestSuite) TestEveryS3KeyIsRepresentable() {
	alphabet := []rune("abcXYZ019 !-_.*'()&$@=;:+,?\\\t\x01ünï文файл")
	random := rand.New(rand.NewSource(20260911))
	component := func() string {
		for {
			length := 1 + random.Intn(40)
			runes := make([]rune, length)
			for i := range runes {
				runes[i] = alphabet[random.Intn(len(alphabet))]
			}
			if s := string(runes); s != ".." {
				return s
			}
		}
	}

	for i := 0; i < 2000; i++ {
		depth := 1 + random.Intn(6)
		components := make([]string, depth)
		for j := range components {
			components[j] = component()
		}
		if components[0] == "." {
			components[0] = "x"
		}
		key := strings.Join(components, "/")
		if len(key) > 1024 {
			continue
		}

		got, err := virtualPathForS3Key(key)
		require.NoError(suite.T(), err, "key %q must be representable", key)
		require.Equal(suite.T(), path.Clean(key), got)
		require.NotContains(suite.T(), strings.Split(got, "/"), "..")
	}
}

func (suite *S3KeysTestSuite) TestVirtualPathsForS3KeysMapsEveryKey() {
	got, err := virtualPathsForS3Keys([]string{"README.md", "/lead.txt", "a//b", "./c.txt", `d\e.txt`})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), map[string]string{
		"README.md": "README.md",
		"/lead.txt": "lead.txt",
		"a//b":      "a/b",
		"./c.txt":   "c.txt",
		`d\e.txt`:   `d\e.txt`,
	}, got)
}

func (suite *S3KeysTestSuite) TestVirtualPathsForS3KeysNamesEveryCollidingKey() {
	_, err := virtualPathsForS3Keys([]string{"x", "a/b", "a//b", "./a/b"})
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "[./a/b]")
	require.Contains(suite.T(), err.Error(), "[a//b]")
	require.Contains(suite.T(), err.Error(), "[a/b]")
	require.Contains(suite.T(), err.Error(), "--exclude-regex")
	require.NotContains(suite.T(), err.Error(), "[x]", "an unaffected key must not be named")
}

// An object "a" beside objects under "a/" is legal in S3 and impossible in a
// directory tree, whichever order S3 lists them in.
func (suite *S3KeysTestSuite) TestVirtualPathsForS3KeysObjectAndPrefix() {
	for _, keys := range [][]string{
		{"a", "a/b"},
		{"a/b", "a"},
		{"a/b/c", "a/b", "z"},
		{"lib", "lib/x", "lib/y"},
	} {
		suite.Run(strings.Join(keys, ","), func() {
			_, err := virtualPathsForS3Keys(keys)
			require.Error(suite.T(), err)
			require.Contains(suite.T(), err.Error(), "also a directory")
			require.Contains(suite.T(), err.Error(), "--exclude-regex")
		})
	}

	_, err := virtualPathsForS3Keys([]string{"a", "a/b"})
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "object key [a] fingerprints as [a], which is also a directory holding object key [a/b]")
}

func (suite *S3KeysTestSuite) TestVirtualPathsForS3KeysReportsEveryBadKey() {
	_, err := virtualPathsForS3Keys([]string{"ok.txt", "../one", "two/..", "", "dup", "./dup"})
	require.Error(suite.T(), err)
	msg := err.Error()
	require.Contains(suite.T(), msg, "4 object keys cannot be fingerprinted")
	require.Contains(suite.T(), msg, "[../one]")
	require.Contains(suite.T(), msg, "[two/..]")
	require.Contains(suite.T(), msg, "object key []")
	require.Contains(suite.T(), msg, "[./dup]")
	require.Contains(suite.T(), msg, "--exclude-regex")
	require.NotContains(suite.T(), msg, "[ok.txt]")
}

func (suite *S3KeysTestSuite) TestVirtualPathsForS3KeysCapsTheReport() {
	keys := make([]string, 0, 13)
	for i := 0; i < 13; i++ {
		keys = append(keys, fmt.Sprintf("%02d/../x", i))
	}
	_, err := virtualPathsForS3Keys(keys)
	require.Error(suite.T(), err)
	msg := err.Error()
	require.Contains(suite.T(), msg, "13 object keys cannot be fingerprinted")
	require.Contains(suite.T(), msg, "(and 3 more)")
	require.Equal(suite.T(), maxReportedS3KeyProblems, strings.Count(msg, "object key ["))
}

// The reported attack shape from #1155: the traversing key is rejected for its
// ".." segment, and is never allowed to fold onto the object it names.
func (suite *S3KeysTestSuite) TestVirtualPathsForS3KeysTraversalKeyIsRejected() {
	_, err := virtualPathsForS3Keys([]string{
		"protected/release.bin",
		"uploads/user-a/../../protected/release.bin",
	})
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "object key [uploads/user-a/../../protected/release.bin]")
	require.Contains(suite.T(), err.Error(), `contains a ".." segment`)
	require.NotContains(suite.T(), err.Error(), "fingerprint as the same path",
		"the traversing key must be rejected outright, not reported as a collision")
}

func (suite *S3KeysTestSuite) TestVirtualPathsForS3KeysSingleProblemReadsAsOneLine() {
	_, err := virtualPathsForS3Keys([]string{"good", "bad/.."})
	require.Error(suite.T(), err)
	require.Equal(suite.T(),
		`object key [bad/..] cannot be fingerprinted: contains a ".." segment; exclude it with --exclude-regex, or narrow the include filter if one is set`,
		err.Error())
}

func (suite *S3KeysTestSuite) TestVirtualPathsForS3KeysIsDeterministic() {
	keys := []string{"c/..", "b/..", "a/..", "d", "d/e"}
	first, err := virtualPathsForS3Keys(keys)
	require.Error(suite.T(), err)
	require.Nil(suite.T(), first)
	for i := 0; i < 20; i++ {
		_, again := virtualPathsForS3Keys(keys)
		require.Equal(suite.T(), err.Error(), again.Error())
	}
}

func TestS3KeysTestSuite(t *testing.T) {
	suite.Run(t, new(S3KeysTestSuite))
}
