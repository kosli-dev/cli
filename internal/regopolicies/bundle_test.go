package regopolicies

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type member struct {
	name     string
	body     string
	typeflag byte
}

func tarball(t *testing.T, members ...member) []byte {
	t.Helper()
	var archive bytes.Buffer
	zipped := gzip.NewWriter(&archive)
	tw := tar.NewWriter(zipped)
	for _, m := range members {
		typeflag := m.typeflag
		if typeflag == 0 {
			typeflag = tar.TypeReg
		}
		require.NoError(t, tw.WriteHeader(&tar.Header{
			Name: m.name, Mode: 0o644, Size: int64(len(m.body)), Typeflag: typeflag,
		}))
		_, err := tw.Write([]byte(m.body))
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	require.NoError(t, zipped.Close())
	return archive.Bytes()
}

func filesIn(t *testing.T, dir string) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		files[filepath.ToSlash(rel)] = string(content)
		return err
	})
	require.NoError(t, err)
	return files
}

func TestExtractBundleWritesEachFileAtItsPath(t *testing.T) {
	dir := t.TempDir()
	archive := tarball(t,
		member{name: "lib/my café.rego", body: "package lib\n"},
		member{name: "policy.rego", body: "package policy\n"},
	)

	written, err := ExtractBundle(archive, dir)

	require.NoError(t, err)
	require.Equal(t, []string{"lib/my café.rego", "policy.rego"}, written)
	require.Equal(t, map[string]string{
		"lib/my café.rego": "package lib\n",
		"policy.rego":      "package policy\n",
	}, filesIn(t, dir))
}

func TestExtractBundleRefusesAPathOutsideTheDirectoryAndWritesNothing(t *testing.T) {
	for _, name := range []string{"../escape.rego", "lib/../../escape.rego", "/etc/escape.rego"} {
		parent := t.TempDir()
		dir := filepath.Join(parent, "out")
		archive := tarball(t,
			member{name: "policy.rego", body: "package policy\n"},
			member{name: name, body: "package escape\n"},
		)

		_, err := ExtractBundle(archive, dir)

		require.ErrorContains(t, err, name, name)
		require.Empty(t, filesIn(t, parent), name)
	}
}

func TestExtractBundleRefusesAMemberThatIsNotARegularFileAndWritesNothing(t *testing.T) {
	dir := t.TempDir()
	archive := tarball(t,
		member{name: "policy.rego", body: "package policy\n"},
		member{name: "link.rego", typeflag: tar.TypeSymlink},
	)

	_, err := ExtractBundle(archive, dir)

	require.ErrorContains(t, err, "link.rego")
	require.Empty(t, filesIn(t, dir))
}

func TestExtractBundleRefusesToOverwriteAFileAndWritesNothing(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "policy.rego"), []byte("mine\n"), 0o644))
	archive := tarball(t,
		member{name: "a.rego", body: "package a\n"},
		member{name: "policy.rego", body: "package policy\n"},
	)

	_, err := ExtractBundle(archive, dir)

	require.ErrorContains(t, err, "policy.rego")
	require.Equal(t, map[string]string{"policy.rego": "mine\n"}, filesIn(t, dir))
}

func TestExtractBundleCreatesTheDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "pr-approved-v3")
	archive := tarball(t, member{name: "policy.rego", body: "package policy\n"})

	_, err := ExtractBundle(archive, dir)

	require.NoError(t, err)
	require.Equal(t, map[string]string{"policy.rego": "package policy\n"}, filesIn(t, dir))
}
