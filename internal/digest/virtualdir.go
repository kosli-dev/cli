package digest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
// ignoreRules are the entries of the tree's root .kosli_ignore, as
// ParseIgnoreRules returns them. They are resolved against the tree exactly as
// DirSha256 resolves them against a directory (see virtualFS), and the root
// .kosli_ignore itself is never excluded by them, so a tree cannot change its
// exclusion list without changing its fingerprint. Callers that hold the file's
// content pass its rules here; the file is an ordinary entry of files.
func VirtualDirSha256(files []VirtualFile, ignoreRules []string, logger *logger.Logger) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("cannot calculate a fingerprint: no files were provided")
	}

	root, err := buildVirtualTree(files)
	if err != nil {
		return "", err
	}

	excluded, err := virtualFS{root: root}.excludedPaths(ignoreRules)
	if err != nil {
		return "", err
	}

	logger.Debug("calculating fingerprint for a virtual tree of %d files -- excluding %d paths", len(files), len(excluded))
	hasher := sha256.New()
	err = root.walkIncluded(virtualRoot, excluded, protectedVirtualPath(), logger, func(childPath string, child *virtualNode) error {
		nameSha256 := sha256OfString(child.name)
		hasher.Write([]byte(nameSha256)) //nolint:errcheck // hash.Hash never returns an error
		if child.isDir {
			logger.Debug("dir: %s -- dirname digest: %s", child.name, nameSha256)
			return nil
		}
		if child.sha256 == "" {
			return fmt.Errorf("no content digest for %q, whose content the fingerprint needs", relativeVirtualPath(childPath))
		}
		logger.Debug("file: %s -- filename digest: %s -- content digest: %s", child.name, nameSha256, child.sha256)
		hasher.Write([]byte(child.sha256)) //nolint:errcheck // hash.Hash never returns an error
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// FilesNeedingContent reports which of paths VirtualDirSha256 reads the content
// digest of under these ignore rules. A file it leaves out is skipped by the
// rules, so its content need not be fetched and it may be passed with an empty
// Sha256 without changing the fingerprint. The two share one walk, so they
// cannot disagree.
func FilesNeedingContent(paths []string, ignoreRules []string) (map[string]bool, error) {
	files := make([]VirtualFile, len(paths))
	for i, p := range paths {
		files[i] = VirtualFile{Path: p}
	}
	root, err := buildVirtualTree(files)
	if err != nil {
		return nil, err
	}
	excluded, err := virtualFS{root: root}.excludedPaths(ignoreRules)
	if err != nil {
		return nil, err
	}
	needed := map[string]bool{}
	err = root.walkIncluded(virtualRoot, excluded, protectedVirtualPath(), nil, func(childPath string, child *virtualNode) error {
		if !child.isDir {
			needed[relativeVirtualPath(childPath)] = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return needed, nil
}

// protectedVirtualPath is the root ignore file, which its own rules never exclude.
func protectedVirtualPath() string {
	return path.Join(virtualRoot, ignoreFileName)
}

// relativeVirtualPath strips the synthetic root from a tree path.
func relativeVirtualPath(p string) string {
	return strings.TrimPrefix(p, virtualRoot+"/")
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
		// An empty digest means the content was not read. That is only acceptable
		// for a file the rules exclude, which writeDigests enforces when it gets there.
		if file.Sha256 != "" {
			if err := ValidateDigest(file.Sha256); err != nil {
				return nil, fmt.Errorf("invalid fingerprint for %q: %w", file.Path, err)
			}
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

// walkIncluded visits this node's children in WalkDir order, skipping excluded
// entries as calculateDirContentSha256 does: an excluded directory takes its
// subtree with it, and the protected path is kept whatever the rules say.
// A nil logger is allowed for callers that only want the visits.
func (n *virtualNode) walkIncluded(dir string, excluded map[string]bool, protected string, logger *logger.Logger,
	visit func(childPath string, child *virtualNode) error) error {
	for _, name := range n.sortedChildNames() {
		child := n.children[name]
		childPath := path.Join(dir, name)
		if excluded[childPath] {
			if childPath != protected {
				if logger != nil {
					logger.Debug("skipping %s as it matches excluded paths", childPath)
				}
				continue
			}
			if logger != nil {
				logger.Debug("keeping %s although an exclusion matches it: an exclusion list cannot exclude itself", childPath)
			}
		}
		if err := visit(childPath, child); err != nil {
			return err
		}
		if child.isDir {
			if err := child.walkIncluded(childPath, excluded, protected, logger, visit); err != nil {
				return err
			}
		}
	}
	return nil
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
