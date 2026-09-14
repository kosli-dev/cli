package utils

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// Accept rows compare joined paths because the helper returns the name
// uncleaned; filepath.Join is what collapses "a//b" and "./a.txt".
func TestLocalRelativePath(t *testing.T) {
	for _, tc := range []struct {
		name     string
		input    string
		wantPath string
		wantErr  error
	}{
		{name: "an ordinary nested name", input: "static/app.js", wantPath: "static/app.js"},
		{name: "a plain filename", input: "a.txt", wantPath: "a.txt"},
		{name: "a dotfile", input: ".kosli_ignore", wantPath: ".kosli_ignore"},
		{name: "a name that merely starts with two dots", input: "..hidden", wantPath: "..hidden"},
		{name: "a directory name with a trailing slash", input: "static/", wantPath: "static"},
		{name: "a leading slash is trimmed", input: "/etc/passwd", wantPath: "etc/passwd"},
		{name: "a leading dot segment is dropped by Join", input: "./a.txt", wantPath: "a.txt"},
		{name: "a dot segment is dropped by Join", input: "a/./b", wantPath: "a/b"},
		// filepath.IsLocal rejects reserved device names and colons on Windows only.
		{name: "a reserved Windows name", input: "CON", wantPath: "CON", wantErr: onWindows(ErrNotLocalPath)},
		{name: "a drive-looking segment", input: "C:evil", wantPath: "C:evil", wantErr: onWindows(ErrNotLocalPath)},
		{name: "a traversing name is rejected", input: "a/../../b", wantErr: ErrPathTraversal},
		{name: "a backslash-separated traversal is rejected", input: `a\..\..\b`, wantErr: ErrPathTraversal},
		{name: "a bare \"..\" is rejected", input: "..", wantErr: ErrPathTraversal},
		{name: "a trailing \"..\" segment is rejected", input: "a/..", wantErr: ErrPathTraversal},
		// Windows drops trailing spaces and dots from a name, so these resolve as "..".
		{name: "a \"..\" with a trailing space is rejected", input: "a/.. /x", wantErr: ErrPathTraversal},
		{name: "three dots are rejected", input: "a/.../b", wantErr: ErrPathTraversal},
		{name: "an empty name is rejected", input: "", wantErr: ErrNamesNoFile},
		{name: "a bare dot is rejected", input: ".", wantErr: ErrNamesNoFile},
		{name: "a dot directory is rejected", input: "./", wantErr: ErrNamesNoFile},
		{name: "a bare slash is rejected", input: "/", wantErr: ErrNamesNoFile},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := LocalRelativePath(tc.input)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, filepath.Join("base", tc.wantPath), filepath.Join("base", got))
		})
	}
}

func TestContainedPath(t *testing.T) {
	got, err := ContainedPath("base", "static/app.js")
	require.NoError(t, err)
	require.Equal(t, filepath.Join("base", "static", "app.js"), got)

	_, err = ContainedPath("base", "../escape")
	require.ErrorIs(t, err, ErrPathTraversal)
	require.Equal(t, `[../escape] contains a segment that resolves to ".."`, err.Error(),
		"the error must name the offending entry and read as a sentence once prefixed")
}

func onWindows(err error) error {
	if runtime.GOOS == "windows" {
		return err
	}
	return nil
}
