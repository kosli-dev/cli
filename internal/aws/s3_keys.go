package aws

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// maxReportedS3KeyProblems caps how many keys one error lists before
// summarising the rest, so a bucket-wide problem stays readable.
const maxReportedS3KeyProblems = 10

// virtualPathForS3Key returns the path an object key occupies in the virtual
// tree that is fingerprinted, or rejects a key no directory tree can hold.
//
// Nothing is ever created under the returned path, so the operating system's
// naming rules do not apply: reserved names, colons, backslashes and overlong
// components are all ordinary names here. The fold of "." segments, doubled
// slashes and a leading slash is what filepath.Join did when objects were
// written under their keys, and keeps existing fingerprints unchanged.
//
// A ".." segment is checked on the raw key rather than left to path.Clean,
// which would fold "a/../b" onto "b" silently. DirSha256 walks real
// directories, which never hold an entry named ".", ".." or "", so rejecting
// these keeps every virtual fingerprint inside the space an attested directory
// can match.
func virtualPathForS3Key(key string) (string, error) {
	trimmed := strings.TrimLeft(key, "/")
	if trimmed == "" {
		return "", s3KeyProblem(key, "names no file")
	}
	if strings.HasSuffix(trimmed, "/") {
		return "", s3KeyProblem(key, "is a folder marker, not an object")
	}
	for _, segment := range strings.Split(trimmed, "/") {
		if segment == ".." {
			return "", s3KeyProblem(key, `contains a ".." segment`)
		}
	}
	cleaned := path.Clean(trimmed)
	if cleaned == "." {
		return "", s3KeyProblem(key, "names no file")
	}
	return cleaned, nil
}

// virtualPathsForS3Keys maps every object key to its virtual path, or reports
// every key that cannot take part in one directory tree: keys the rule above
// rejects, keys that fold onto the same path, and an object whose path is also
// a directory holding other objects. All problems are reported together so one
// run tells the operator about every key they need to act on.
//
// Folder markers (keys ending in "/") are the caller's to filter out first.
func virtualPathsForS3Keys(keys []string) (map[string]string, error) {
	paths := make(map[string]string, len(keys))
	keysByPath := map[string][]string{}
	problems := []string{}

	for _, key := range keys {
		virtualPath, err := virtualPathForS3Key(key)
		if err != nil {
			problems = append(problems, err.Error())
			continue
		}
		paths[key] = virtualPath
		keysByPath[virtualPath] = append(keysByPath[virtualPath], key)
	}

	for virtualPath, colliding := range keysByPath {
		if len(colliding) > 1 {
			sort.Strings(colliding)
			problems = append(problems, fmt.Sprintf("object keys %s fingerprint as the same path [%s]",
				bracketed(colliding), virtualPath))
		}
	}

	// The lexically smallest key under each directory stands as the example in
	// the message, so the report does not depend on listing order.
	exampleObjectUnder := map[string]string{}
	for virtualPath, keys := range keysByPath {
		sort.Strings(keys)
		for dir := path.Dir(virtualPath); dir != "."; dir = path.Dir(dir) {
			if existing, ok := exampleObjectUnder[dir]; !ok || keys[0] < existing {
				exampleObjectUnder[dir] = keys[0]
			}
		}
	}
	for virtualPath, keys := range keysByPath {
		child, isAlsoDir := exampleObjectUnder[virtualPath]
		if !isAlsoDir {
			continue
		}
		for _, key := range keys {
			problems = append(problems, fmt.Sprintf("object key [%s] fingerprints as [%s], which is also a directory holding object key [%s]",
				key, virtualPath, child))
		}
	}

	if len(problems) > 0 {
		return nil, s3KeyProblemsError(problems)
	}
	return paths, nil
}

// s3KeyProblem describes a failure the key itself causes. The advice belongs
// only on such failures: suggesting exclusion for a machine fault would drop a
// legitimate object from the snapshot.
func s3KeyProblem(key, reason string) error {
	return fmt.Errorf("object key [%s] cannot be fingerprinted: %s", key, reason)
}

// s3KeyProblemsError joins every key problem into one error with the remedy.
func s3KeyProblemsError(problems []string) error {
	// Map iteration supplied these in any order; sorting keeps the message stable.
	sort.Strings(problems)
	if len(problems) == 1 {
		return fmt.Errorf("%s; exclude it with --exclude-regex, or narrow the include filter if one is set", problems[0])
	}

	shown := problems
	suffix := ""
	if len(shown) > maxReportedS3KeyProblems {
		shown = shown[:maxReportedS3KeyProblems]
		suffix = fmt.Sprintf("\n(and %d more)", len(problems)-maxReportedS3KeyProblems)
	}
	return fmt.Errorf("%d object keys cannot be fingerprinted:\n%s%s\nexclude them with --exclude-regex, or narrow the include filter if one is set",
		len(problems), strings.Join(shown, "\n"), suffix)
}

// bracketed formats keys as "[a], [b]" for error messages.
func bracketed(keys []string) string {
	parts := make([]string, len(keys))
	for i, key := range keys {
		parts[i] = "[" + key + "]"
	}
	return strings.Join(parts, ", ")
}
