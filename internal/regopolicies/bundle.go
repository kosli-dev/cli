package regopolicies

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

// The server caps a bundle at 1 MiB; the margin is for tar headers and a future raise.
const maxBundleBytes = 8 << 20

type bundleFile struct {
	path    string
	content []byte
}

// ExtractBundle writes the files of a version's bundle into dir and returns their
// paths in archive order. It writes nothing unless every file can be written.
func ExtractBundle(archive []byte, dir string) ([]string, error) {
	files, err := readBundle(archive)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	// Writing through a Root keeps a symlink already inside dir from leading a write out of it.
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()

	for _, f := range files {
		if _, err := root.Lstat(f.path); err == nil {
			return nil, fmt.Errorf("%s already exists, so the bundle was not written", filepath.Join(dir, f.path))
		} else if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("cannot write %q into %s: %w", f.path, dir, err)
		}
	}
	written := make([]string, 0, len(files))
	for _, f := range files {
		if err := writeFile(root, f); err != nil {
			return written, fmt.Errorf("cannot write %q into %s: %w", f.path, dir, err)
		}
		written = append(written, f.path)
	}
	return written, nil
}

func writeFile(root *os.Root, f bundleFile) error {
	if err := root.MkdirAll(path.Dir(f.path), 0o755); err != nil {
		return err
	}
	out, err := root.OpenFile(f.path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := out.Write(f.content); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func readBundle(archive []byte) ([]bundleFile, error) {
	zipped, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("the rego policy bundle is not a gzipped tar: %w", err)
	}
	limited := &io.LimitedReader{R: zipped, N: maxBundleBytes + 1}
	tr := tar.NewReader(limited)
	var files []bundleFile
	isFile := map[string]bool{}
	isDir := map[string]bool{}
	for {
		header, err := tr.Next()
		if limited.N <= 0 {
			return nil, errBundleTooLarge
		}
		if errors.Is(err, io.EOF) {
			return files, nil
		}
		if err != nil {
			return nil, fmt.Errorf("the rego policy bundle is not a gzipped tar: %w", err)
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeDir {
			return nil, fmt.Errorf("the rego policy bundle holds %q, which is not a regular file", header.Name)
		}
		if !filepath.IsLocal(filepath.FromSlash(header.Name)) {
			return nil, fmt.Errorf("the rego policy bundle holds %q, a path outside the directory", header.Name)
		}
		name := path.Clean(header.Name)
		if err := markParents(name, isFile, isDir); err != nil {
			return nil, err
		}
		if header.Typeflag == tar.TypeDir {
			if isFile[name] {
				return nil, bothFileAndDirectory(name)
			}
			isDir[name] = true
			continue
		}
		if isFile[name] {
			return nil, fmt.Errorf("the rego policy bundle holds %q twice", name)
		}
		if isDir[name] {
			return nil, bothFileAndDirectory(name)
		}
		isFile[name] = true
		content, err := io.ReadAll(tr)
		if limited.N <= 0 {
			return nil, errBundleTooLarge
		}
		if err != nil {
			return nil, err
		}
		files = append(files, bundleFile{path: name, content: content})
	}
}

var errBundleTooLarge = fmt.Errorf("the rego policy bundle is over the %d MiB limit", maxBundleBytes>>20)

func markParents(name string, isFile, isDir map[string]bool) error {
	for parent := path.Dir(name); parent != "."; parent = path.Dir(parent) {
		if isFile[parent] {
			return bothFileAndDirectory(parent)
		}
		isDir[parent] = true
	}
	return nil
}

func bothFileAndDirectory(name string) error {
	return fmt.Errorf("the rego policy bundle holds %q both as a file and as a directory", name)
}
