package digest

import (
	"fmt"
	"path"
	"strings"
)

// virtualRoot stands in for the directory path DirSha256 is given. Every path in
// the virtual tree is spelled "tree/<relative path>", so a rule joined onto the
// root goes through exactly the string handling filepath.Join, filepath.Glob and
// filepath.Walk apply on disk.
const virtualRoot = "tree"

// virtualFS answers the questions filepath.Glob and filepath.Walk ask of a
// filesystem, over a virtual tree instead.
//
// DirSha256 resolves .kosli_ignore rules with filepathx.Glob, which has its own
// reading of "**" and inherits filepath.Glob's: a pattern without wildcards is
// returned as written, uncleaned, while one with wildcards is rebuilt from the
// directory listing, cleaned. So "**/x" never matches x at the root (the pieces
// concatenate to "tree//x") but "**/*.log" does. Reproducing the algorithm step
// for step, rather than its apparent meaning, is what keeps the virtual digest
// equal to the on-disk one for every rule anyone has already written.
type virtualFS struct {
	root *virtualNode
}

// excludedPaths resolves ignore rules to the set of tree paths DirSha256 would
// skip, spelled exactly as the resolving glob returned them.
func (fs virtualFS) excludedPaths(rules []string) (map[string]bool, error) {
	excluded := map[string]bool{}
	for _, rule := range rules {
		pattern := path.Join(virtualRoot, rule)
		// On disk the root is a temp directory with an unguessable name, so a rule
		// that resolves to the root or outside it cannot match anything there.
		if pattern == virtualRoot || !strings.HasPrefix(pattern, virtualRoot+"/") {
			continue
		}
		matches, err := fs.globDoubleStar(pattern)
		if err != nil {
			return nil, fmt.Errorf("ignore rule %q: %w", rule, err)
		}
		for _, match := range matches {
			excluded[match] = true
		}
	}
	return excluded, nil
}

// globDoubleStar mirrors filepathx.Glob: split the pattern on "**", glob each
// piece appended to every match so far, and walk every hit so the next piece is
// tried under all of its descendants.
func (fs virtualFS) globDoubleStar(pattern string) ([]string, error) {
	if !strings.Contains(pattern, "**") {
		return fs.glob(pattern)
	}
	matches := []string{""}
	for _, piece := range strings.Split(pattern, "**") {
		var hits []string
		seen := map[string]bool{}
		for _, match := range matches {
			paths, err := fs.glob(match + piece)
			if err != nil {
				return nil, err
			}
			for _, p := range paths {
				err := fs.walk(p, func(visited string) {
					if !seen[visited] {
						hits = append(hits, visited)
						seen[visited] = true
					}
				})
				if err != nil {
					return nil, err
				}
			}
		}
		matches = hits
	}
	return matches, nil
}

// glob mirrors filepath.Glob on a Unix filesystem.
func (fs virtualFS) glob(pattern string) ([]string, error) {
	if !hasGlobMeta(pattern) {
		// A literal pattern is returned as written, not cleaned.
		if _, ok := fs.lookup(pattern); !ok {
			return nil, nil
		}
		return []string{pattern}, nil
	}

	dir, file := path.Split(pattern)
	dir = cleanGlobPath(dir)
	if !hasGlobMeta(dir) {
		return fs.globDir(dir, file, nil)
	}
	if dir == pattern {
		return nil, path.ErrBadPattern
	}
	dirs, err := fs.glob(dir)
	if err != nil {
		return nil, err
	}
	var matches []string
	for _, d := range dirs {
		matches, err = fs.globDir(d, file, matches)
		if err != nil {
			return nil, err
		}
	}
	return matches, nil
}

// globDir mirrors filepath.glob: match the pattern against each name in one
// directory and return the joined, cleaned paths.
func (fs virtualFS) globDir(dir, pattern string, matches []string) ([]string, error) {
	node, ok := fs.lookup(dir)
	if !ok || !node.isDir {
		return matches, nil
	}
	for _, name := range node.sortedChildNames() {
		matched, err := path.Match(pattern, name)
		if err != nil {
			return matches, err
		}
		if matched {
			matches = append(matches, path.Join(dir, name))
		}
	}
	return matches, nil
}

// walk mirrors filepath.Walk as filepathx drives it: the root is reported as
// given, descendants as joined, cleaned paths, in lexical order.
func (fs virtualFS) walk(root string, visit func(string)) error {
	node, ok := fs.lookup(root)
	if !ok {
		return fmt.Errorf("lstat %s: no such file or directory", root)
	}
	fs.walkNode(root, node, visit)
	return nil
}

func (fs virtualFS) walkNode(p string, node *virtualNode, visit func(string)) {
	visit(p)
	if !node.isDir {
		return
	}
	for _, name := range node.sortedChildNames() {
		fs.walkNode(path.Join(p, name), node.children[name], visit)
	}
}

// lookup finds the node a possibly uncleaned path spells, or reports that no
// such entry exists.
func (fs virtualFS) lookup(p string) (*virtualNode, bool) {
	cleaned := path.Clean(p)
	if cleaned == virtualRoot {
		return fs.root, true
	}
	if !strings.HasPrefix(cleaned, virtualRoot+"/") {
		return nil, false
	}
	node := fs.root
	for _, segment := range strings.Split(cleaned[len(virtualRoot)+1:], "/") {
		if !node.isDir {
			return nil, false
		}
		child, ok := node.children[segment]
		if !ok {
			return nil, false
		}
		node = child
	}
	return node, true
}

// hasGlobMeta reports whether filepath.Glob would treat the pattern as a glob
// on Unix, where a backslash is an escape character.
func hasGlobMeta(pattern string) bool {
	return strings.ContainsAny(pattern, `*?[\`)
}

// cleanGlobPath is filepath's helper of the same name: drop the trailing
// separator path.Split leaves on a directory.
func cleanGlobPath(dir string) string {
	switch dir {
	case "":
		return "."
	case "/":
		return dir
	default:
		return dir[:len(dir)-1]
	}
}
