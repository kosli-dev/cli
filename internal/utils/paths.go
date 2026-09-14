package utils

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// The rejections LocalRelativePath returns. Each message is a sentence fragment
// meant to be prefixed by the caller with the offending name.
var (
	ErrPathTraversal = errors.New(`contains a segment that resolves to ".."`)
	ErrNamesNoFile   = errors.New("names no file")
	ErrNotLocalPath  = errors.New("is not a local path")
)

// LocalRelativePath turns a name supplied by an external system, such as an S3
// object key or a zip entry, into a relative path that stays under whichever
// directory it is joined onto, or rejects it with one of the Err* sentinels.
// Only a leading "/" is trimmed. A name holding a ".." segment resolves onto a
// path it does not name, taking another name's place or leaving the directory.
func LocalRelativePath(name string) (string, error) {
	// Windows separates on '\\' and drops trailing dots and spaces from a
	// name, so ".. " and "..." resolve as ".." there.
	segments := strings.FieldsFunc(name, func(r rune) bool { return r == '/' || r == '\\' })
	for _, segment := range segments {
		if strings.HasPrefix(segment, "..") && strings.TrimRight(segment, ". ") == "" {
			return "", ErrPathTraversal
		}
	}

	// A leading '\\' is left for filepath.IsLocal: rooted on Windows, an
	// ordinary filename elsewhere.
	rel := strings.TrimLeft(name, "/")
	if filepath.Clean(rel) == "." {
		return "", ErrNamesNoFile
	}
	if !filepath.IsLocal(rel) {
		return "", ErrNotLocalPath
	}

	return rel, nil
}

// ContainedPath joins name onto dir after checking with LocalRelativePath
// that the result cannot leave dir. The error names the entry and wraps the
// sentinel, so errors.Is still tells the rejections apart.
func ContainedPath(dir, name string) (string, error) {
	rel, err := LocalRelativePath(name)
	if err != nil {
		return "", fmt.Errorf("[%s] %w", name, err)
	}
	return filepath.Join(dir, rel), nil
}
