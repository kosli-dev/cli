package azure

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/stretchr/testify/require"
)

type zipEntry struct {
	name    string
	content string
	isDir   bool
}

func writeZip(t *testing.T, path string, entries []zipEntry) {
	t.Helper()
	f, err := os.Create(path)
	require.NoError(t, err)
	w := zip.NewWriter(f)
	for _, e := range entries {
		header := &zip.FileHeader{Name: e.name, Method: zip.Deflate}
		if e.isDir {
			header.SetMode(os.ModeDir | 0o755)
		} else {
			header.SetMode(0o644)
		}
		entry, err := w.CreateHeader(header)
		require.NoError(t, err)
		if !e.isDir {
			_, err = entry.Write([]byte(e.content))
			require.NoError(t, err)
		}
	}
	require.NoError(t, w.Close())
	require.NoError(t, f.Close())
}

func TestUnzipExtractsAnOrdinaryPackage(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "package.zip")
	writeZip(t, zipPath, []zipEntry{
		{name: "index.html", content: "<html/>"},
		{name: "static/", isDir: true},
		{name: "static/app.js", content: "console.log(1)"},
		{name: "deep/er/file.txt", content: "no directory entry for the parents"},
	})

	destDir := filepath.Join(tmpDir, "extracted")
	require.NoError(t, unzip(zipPath, destDir, logger.NewStandardLogger()))

	for path, want := range map[string]string{
		"index.html":       "<html/>",
		"static/app.js":    "console.log(1)",
		"deep/er/file.txt": "no directory entry for the parents",
	} {
		got, err := os.ReadFile(filepath.Join(destDir, path))
		require.NoError(t, err, path)
		require.Equal(t, want, string(got), path)
	}
}

func TestUnzipRejectsEntriesThatEscapeTheDestination(t *testing.T) {
	for _, tc := range []struct {
		name       string
		entry      string
		wantErrMsg string
	}{
		{name: "a traversing file", entry: "../escape.txt", wantErrMsg: `resolves to ".."`},
		{name: "a deeply traversing file", entry: "static/../../../escape.txt", wantErrMsg: `resolves to ".."`},
		{name: "a traversing directory", entry: "../escaped-dir/", wantErrMsg: `resolves to ".."`},
		// Windows separates on '\' and drops trailing dots and spaces from a name.
		{name: "a backslash traversal", entry: `..\escape.txt`, wantErrMsg: `resolves to ".."`},
		{name: "a traversal with a trailing space", entry: ".. /escape.txt", wantErrMsg: `resolves to ".."`},
		{name: "a file entry naming no file", entry: ".", wantErrMsg: "names no file"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			zipPath := filepath.Join(tmpDir, "package.zip")
			writeZip(t, zipPath, []zipEntry{
				{name: "index.html", content: "<html/>"},
				{name: tc.entry, content: "attacker controlled", isDir: tc.entry[len(tc.entry)-1] == '/'},
			})

			// destDir is two levels below tmpDir so "../.." lands inside tmpDir,
			// where the test can see it, rather than in the system temp dir.
			destDir := filepath.Join(tmpDir, "work", "extracted")
			err := unzip(zipPath, destDir, logger.NewStandardLogger())
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.entry, "the error must name the offending entry")
			require.Contains(t, err.Error(), tc.wantErrMsg)

			// Nothing may have been written outside destDir.
			entries, readErr := os.ReadDir(tmpDir)
			require.NoError(t, readErr)
			for _, e := range entries {
				require.Contains(t, []string{"package.zip", "work"}, e.Name(), "an entry escaped into %s", tmpDir)
			}
			workEntries, readErr := os.ReadDir(filepath.Join(tmpDir, "work"))
			require.NoError(t, readErr)
			for _, e := range workEntries {
				require.Equal(t, "extracted", e.Name(), "an entry escaped into %s", filepath.Join(tmpDir, "work"))
			}
		})
	}
}

// A "./" directory entry names the destination itself, which filepath.Join
// turned into a harmless MkdirAll before the containment rule.
func TestUnzipToleratesADirectoryEntryNamingTheDestination(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "package.zip")
	writeZip(t, zipPath, []zipEntry{
		{name: "./", isDir: true},
		{name: "index.html", content: "<html/>"},
	})

	destDir := filepath.Join(tmpDir, "extracted")
	require.NoError(t, unzip(zipPath, destDir, logger.NewStandardLogger()))

	got, err := os.ReadFile(filepath.Join(destDir, "index.html"))
	require.NoError(t, err)
	require.Equal(t, "<html/>", string(got))
}

// Name containment only holds while nothing in destDir is a real symlink, so
// a symlink entry must land as a regular file holding the target text.
func TestUnzipWritesASymlinkEntryAsARegularFile(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "package.zip")
	f, err := os.Create(zipPath)
	require.NoError(t, err)
	w := zip.NewWriter(f)
	header := &zip.FileHeader{Name: "link", Method: zip.Deflate}
	header.SetMode(os.ModeSymlink | 0o777)
	entry, err := w.CreateHeader(header)
	require.NoError(t, err)
	_, err = entry.Write([]byte("/etc/passwd"))
	require.NoError(t, err)
	require.NoError(t, w.Close())
	require.NoError(t, f.Close())

	destDir := filepath.Join(tmpDir, "extracted")
	require.NoError(t, unzip(zipPath, destDir, logger.NewStandardLogger()))

	info, err := os.Lstat(filepath.Join(destDir, "link"))
	require.NoError(t, err)
	require.True(t, info.Mode().IsRegular(), "a symlink entry must not become a symlink, got mode %v", info.Mode())
	got, err := os.ReadFile(filepath.Join(destDir, "link"))
	require.NoError(t, err)
	require.Equal(t, "/etc/passwd", string(got))
}

// Legal in a zip, impossible on disk: one name used as both a file and a
// directory, or two entries landing on one path. The error must name the
// entry that collided, not a temp path.
func TestUnzipNamesTheEntryThatCollides(t *testing.T) {
	for _, tc := range []struct {
		name       string
		entries    []zipEntry
		wantEntry  string
		wantErrMsg string
	}{
		{
			name:       "a file entry under an earlier file",
			entries:    []zipEntry{{name: "a", content: "a file"}, {name: "a/b", content: "under the file"}},
			wantEntry:  "a/b",
			wantErrMsg: "one of its parent directories was already extracted as a file",
		},
		{
			name:       "a directory entry over an earlier file",
			entries:    []zipEntry{{name: "a", content: "a file"}, {name: "a/", isDir: true}},
			wantEntry:  "a/",
			wantErrMsg: "was already extracted as a file",
		},
		{
			name:       "a file entry over an earlier directory",
			entries:    []zipEntry{{name: "a/", isDir: true}, {name: "a", content: "a file"}},
			wantEntry:  "a",
			wantErrMsg: "was already extracted as a directory",
		},
		{
			name:       "a file entry over an implied directory",
			entries:    []zipEntry{{name: "a/b", content: "under a"}, {name: "a", content: "a file"}},
			wantEntry:  "a",
			wantErrMsg: "was already extracted as a directory",
		},
		{
			name:       "two file entries with the same name",
			entries:    []zipEntry{{name: "x", content: "first"}, {name: "x", content: "second"}},
			wantEntry:  "x",
			wantErrMsg: "another entry has already been extracted to the same local path",
		},
		{
			name:       "two file entries that containment resolves to one path",
			entries:    []zipEntry{{name: "x", content: "first"}, {name: "/x", content: "second"}},
			wantEntry:  "/x",
			wantErrMsg: "another entry has already been extracted to the same local path",
		},
		{
			name:       "a dot-prefixed duplicate of an earlier file",
			entries:    []zipEntry{{name: "x", content: "first"}, {name: "./x", content: "second"}},
			wantEntry:  "./x",
			wantErrMsg: "another entry has already been extracted to the same local path",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			zipPath := filepath.Join(tmpDir, "package.zip")
			writeZip(t, zipPath, tc.entries)

			destDir := filepath.Join(tmpDir, "extracted")
			err := unzip(zipPath, destDir, logger.NewStandardLogger())
			require.Error(t, err)
			require.Contains(t, err.Error(), "zip entry ["+tc.wantEntry+"]")
			require.Contains(t, err.Error(), tc.wantErrMsg)
			require.NotContains(t, err.Error(), destDir, "the error must not leak the temp path")
		})
	}
}

// A zero-mode entry (unknown creator, or Unix external attributes left at 0)
// used to be written unreadable, failing later inside the fingerprinter.
func TestUnzipWritesEveryEntryReadable(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "package.zip")
	f, err := os.Create(zipPath)
	require.NoError(t, err)
	w := zip.NewWriter(f)
	header := &zip.FileHeader{Name: "zero-mode.txt", Method: zip.Deflate}
	header.SetMode(0)
	entry, err := w.CreateHeader(header)
	require.NoError(t, err)
	_, err = entry.Write([]byte("content"))
	require.NoError(t, err)
	require.NoError(t, w.Close())
	require.NoError(t, f.Close())

	destDir := filepath.Join(tmpDir, "extracted")
	require.NoError(t, unzip(zipPath, destDir, logger.NewStandardLogger()))

	info, err := os.Stat(filepath.Join(destDir, "zero-mode.txt"))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

// A rooted name is contained by dropping the root, which is what filepath.Join
// did before the containment rule; the extracted layout is unchanged.
func TestUnzipContainsARootedEntry(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "package.zip")
	writeZip(t, zipPath, []zipEntry{{name: "/etc/passwd", content: "not really"}})

	destDir := filepath.Join(tmpDir, "extracted")
	require.NoError(t, unzip(zipPath, destDir, logger.NewStandardLogger()))

	got, err := os.ReadFile(filepath.Join(destDir, "etc", "passwd"))
	require.NoError(t, err)
	require.Equal(t, "not really", string(got))
}
