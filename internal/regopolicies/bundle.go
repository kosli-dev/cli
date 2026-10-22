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
	"path/filepath"
)

type bundleFile struct {
	path    string
	content []byte
}

// ExtractBundle writes the files of a version's bundle into dir and returns their
// paths in archive order.
func ExtractBundle(archive []byte, dir string) ([]string, error) {
	files, err := readBundle(archive)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		target := filepath.Join(dir, filepath.FromSlash(f.path))
		if _, err := os.Lstat(target); err == nil {
			return nil, fmt.Errorf("%s already exists, so the bundle was not written", target)
		} else if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	written := make([]string, 0, len(files))
	for _, f := range files {
		target := filepath.Join(dir, filepath.FromSlash(f.path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return written, err
		}
		if err := os.WriteFile(target, f.content, 0o644); err != nil {
			return written, err
		}
		written = append(written, f.path)
	}
	return written, nil
}

func readBundle(archive []byte) ([]bundleFile, error) {
	zipped, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("the rego policy bundle is not a gzipped tar: %w", err)
	}
	tr := tar.NewReader(zipped)
	var files []bundleFile
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return files, nil
		}
		if err != nil {
			return nil, fmt.Errorf("the rego policy bundle is not a gzipped tar: %w", err)
		}
		if header.Typeflag != tar.TypeReg {
			return nil, fmt.Errorf("the rego policy bundle holds %q, which is not a regular file", header.Name)
		}
		if !filepath.IsLocal(filepath.FromSlash(header.Name)) {
			return nil, fmt.Errorf("the rego policy bundle holds %q, a path outside the directory", header.Name)
		}
		content, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		files = append(files, bundleFile{path: header.Name, content: content})
	}
}
