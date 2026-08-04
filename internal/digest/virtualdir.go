package digest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"path"
	"sort"
	"strings"

	"github.com/kosli-dev/cli/internal/logger"
)

// VirtualFile is one file in a virtual directory tree: a slash-separated path
// relative to the tree root, plus the hex sha256 of the file's content.
type VirtualFile struct {
	// Path is relative to the tree root, slash-separated, with no leading or
	// trailing slash and no "." or ".." segments (e.g. "dummy/template.yml").
	Path string
	// Sha256 is the hex-encoded sha256 of the file content.
	Sha256 string
}

// Name returns the last segment of the file's path.
func (f VirtualFile) Name() string {
	return path.Base(f.Path)
}

// SingleVirtualFile reports whether the tree holds exactly one file, and
// returns it.
//
// This mirrors what containsSingleFile decides for a tree on disk: a tree built
// only from file paths has no empty directories, so a single leaf means every
// level has exactly one child, and two distinct leaves must diverge at some
// node and give it two children. Counting the files is therefore equivalent to
// walking the tree, and callers can pick the FileSha256 branch on len == 1.
func SingleVirtualFile(files []VirtualFile) (VirtualFile, bool) {
	if len(files) != 1 {
		return VirtualFile{}, false
	}
	return files[0], true
}

// VirtualDirSha256 returns the fingerprint DirSha256 would return for a
// directory containing exactly these files, without touching the disk.
//
// It reproduces calculateDirContentSha256 exactly: walk the tree in
// filepath.WalkDir order -- which is lexical by name within each directory,
// depth-first, with directories and files interleaved -- and append, for every
// entry, the hex sha256 of its base name, plus for files the hex sha256 of
// their content. The fingerprint is the sha256 of that concatenation.
//
// Note that the tree has to be built before sorting: object stores list keys in
// byte order of the whole key, and '.' (0x2E) sorts before '/' (0x2F), so keys
// "a.txt" and "a/z" list as [a.txt, a/z] while WalkDir yields [a, a/z, a.txt].
// Sorting the flat path list instead of the tree produces a different digest
// whenever a directory shares a name prefix with a sibling file.
//
// .kosli_ignore is deliberately not handled here: reading it needs the file's
// content, which the caller may not have, so exclusions stay the caller's
// concern.
func VirtualDirSha256(files []VirtualFile, logger *logger.Logger) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("cannot calculate a fingerprint: no files were provided")
	}

	root, err := buildVirtualTree(files)
	if err != nil {
		return "", err
	}

	logger.Debug("calculating fingerprint for a virtual tree of %d files", len(files))
	hasher := sha256.New()
	root.writeDigests(hasher, logger)
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// virtualNode is a directory or a file in the virtual tree. Files are leaves
// and carry a content digest; directories carry children keyed by base name.
type virtualNode struct {
	name     string
	sha256   string
	isDir    bool
	children map[string]*virtualNode
}

// buildVirtualTree turns a flat list of files into a tree, rejecting anything
// that cannot be represented as one: unclean paths, duplicates, and names used
// as both a file and a directory.
func buildVirtualTree(files []VirtualFile) (*virtualNode, error) {
	root := &virtualNode{isDir: true, children: map[string]*virtualNode{}}

	for _, file := range files {
		if err := validateVirtualPath(file.Path); err != nil {
			return nil, err
		}
		if err := ValidateDigest(file.Sha256); err != nil {
			return nil, fmt.Errorf("invalid fingerprint for %q: %w", file.Path, err)
		}

		segments := strings.Split(file.Path, "/")
		parent := root
		for i, segment := range segments[:len(segments)-1] {
			child, ok := parent.children[segment]
			if !ok {
				child = &virtualNode{name: segment, isDir: true, children: map[string]*virtualNode{}}
				parent.children[segment] = child
			}
			if !child.isDir {
				return nil, fmt.Errorf("path %q is both a file and a directory",
					strings.Join(segments[:i+1], "/"))
			}
			parent = child
		}

		name := segments[len(segments)-1]
		if existing, ok := parent.children[name]; ok {
			if existing.isDir {
				return nil, fmt.Errorf("path %q is both a file and a directory", file.Path)
			}
			return nil, fmt.Errorf("duplicate path %q", file.Path)
		}
		parent.children[name] = &virtualNode{name: name, sha256: file.Sha256}
	}

	return root, nil
}

// writeDigests appends this node's children to the hash in WalkDir order.
func (n *virtualNode) writeDigests(hasher hash.Hash, logger *logger.Logger) {
	for _, name := range n.sortedChildNames() {
		child := n.children[name]
		nameSha256 := sha256OfString(child.name)
		hasher.Write([]byte(nameSha256)) //nolint:errcheck // hash.Hash never returns an error

		if child.isDir {
			logger.Debug("dir: %s -- dirname digest: %s", child.name, nameSha256)
			child.writeDigests(hasher, logger)
			continue
		}
		logger.Debug("file: %s -- filename digest: %s -- content digest: %s",
			child.name, nameSha256, child.sha256)
		hasher.Write([]byte(child.sha256)) //nolint:errcheck // hash.Hash never returns an error
	}
}

// sortedChildNames returns child names in the byte order os.ReadDir uses, so
// directories and files interleave exactly as filepath.WalkDir visits them.
func (n *virtualNode) sortedChildNames() []string {
	names := make([]string, 0, len(n.children))
	for name := range n.children {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// validateVirtualPath rejects paths that cannot be mapped onto a directory tree
// unambiguously. path.Clean collapses "a//b" to "a/b" and resolves "." and
// "..", so a path that differs from its cleaned form would silently collide
// with, or escape, another entry.
func validateVirtualPath(p string) error {
	if p == "" || p != path.Clean(p) || path.IsAbs(p) || strings.HasPrefix(p, "../") || p == ".." {
		return fmt.Errorf("path %q is not a clean relative path: it must not be empty, absolute, "+
			"or contain empty, \".\" or \"..\" segments", p)
	}
	return nil
}

// sha256OfString returns the hex sha256 of s. DirSha256 hashes an entry's name
// by writing it to a file and hashing that file, which is the same bytes.
func sha256OfString(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
