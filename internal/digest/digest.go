package digest

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/types"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/utils"
	"github.com/moby/moby/client"
	godigest "github.com/opencontainers/go-digest"
	"github.com/yargevad/filepathx"
)

var (
	// ErrRepoDigestUnavailable returned when repo digest is not available.
	ErrRepoDigestUnavailable = errors.New("repo digest unavailable for the image, " +
		"has it been pushed to or pulled from a registry?")
)

// ignoreFileName is the exclusion list a directory artifact may carry at its root.
const ignoreFileName = ".kosli_ignore"

// DirSha256 returns sha256 digest of a directory
func DirSha256(dirPath string, excludePaths []string, logger *logger.Logger) (string, error) {
	logger.Debug("calculating fingerprint for path [%s] -- excluding paths: %s", dirPath, excludePaths)
	info, err := os.Stat(dirPath)
	if err != nil {
		if dirPath == " " {
			return "", fmt.Errorf("%s. The directory path is '%s'. https://docs.kosli.com/faq/#pathimage-name-is-a-single-whitespace-character", err, dirPath)
		}
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", dirPath)
	}

	tmpDir, err := os.MkdirTemp("", "*")
	if err != nil {
		return "", err
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			logger.Warn("failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	digestsFile, err := os.Create(filepath.Join(tmpDir, "digests"))
	if err != nil {
		return "", err
	}
	defer func() {
		if err := digestsFile.Close(); err != nil {
			// Log warning for cleanup error
			logger.Warn("failed to close digests file: %v", err)
		}
	}()
	// The ignore file must stay in the fingerprint whatever the tree asks for, or a
	// directory could add files and keep the approved fingerprint by listing them
	// (kosli-dev/server#6785). It is protected at the point the walk decides what to
	// skip, using the path the walk itself emits, so no reasoning about how a
	// pattern happens to be spelled or normalised can get between the two.
	ignoreFileInTree, err := ignoreFilePathInTree(dirPath)
	if err != nil {
		return "", err
	}
	protectedPath := ignoreFileInTree

	pathsToExclude, err := resolveExcludePaths(dirPath, excludePaths)
	if err != nil {
		return "", err
	}

	// An operator flag may still exclude it, which is the migration path off the
	// old behaviour. That keeps the file out of the fingerprint while the entries it
	// carries are still applied, so the tree decides what is measured again.
	flagExcludesIgnoreFile := ignoreFileInTree != "" && utils.Contains(pathsToExclude, ignoreFileInTree)
	if flagExcludesIgnoreFile {
		protectedPath = ""
	}

	// The rules are read from the path that was located, not from one rebuilt out
	// of ignoreFileName. Rebuilding it made "the file protected" and "the file whose
	// rules are applied" two independently resolved names, and a rename between the
	// two reads could make them disagree: rules applied with the wrong file, or no
	// file, protected. Reading the located path makes them the same string, and the
	// walk emits that same string by the same construction.
	//
	// One race is left and is not closable here: the rules are read before the file
	// is hashed, so a tree that rewrites it in between has its old rules applied to
	// newly hashed content. Closing that means hashing the bytes that were read,
	// which is a larger change than this.
	ignoredPaths := []string{}
	if ignoreFileInTree != "" {
		ignoredPaths, err = excludePathsFromFile(ignoreFileInTree)
		if err != nil {
			return "", err
		}
	}
	if len(ignoredPaths) > 0 {
		logger.Debug("  -> ignore file used %s -- excluding paths: %s", ignoreFileInTree, ignoredPaths)
		// Warn only once the file is known to carry rules. An empty or comment-only
		// one that a flag excludes weakens nothing, and the message's subject is the
		// paths it lists.
		if flagExcludesIgnoreFile {
			logger.Warn("%s is excluded by a flag, so the paths it lists are still applied while the file itself is not fingerprinted. "+
				"A file added to the directory and listed in %s stays invisible. "+
				"Move the entries to --exclude and delete the file to get the same fingerprint without that.", ignoreFileInTree, ignoreFileName)
		}
	}
	resolvedIgnoredPaths, err := resolveExcludePaths(dirPath, ignoredPaths)
	if err != nil {
		return "", err
	}
	pathsToExclude = append(pathsToExclude, resolvedIgnoredPaths...)

	err = calculateDirContentSha256(digestsFile, dirPath, tmpDir, pathsToExclude, protectedPath, logger)
	if err != nil {
		return "", err
	}

	return FileSha256(digestsFile.Name(), logger)
}

// credentialSource says which credentials a registry lookup may present.
//
// It is a typed choice rather than a caller-supplied SystemContext so that a
// lookup cannot be handed the wrong credential policy by mistake: containers/image
// falls back to credential discovery whenever DockerAuthConfig is nil, and that
// fallback must not reach a registry named by an untrusted source.
type credentialSource int

const (
	// noCredentials presents nothing. It is the zero value deliberately, so a
	// lookup that forgets to say what it wants does not present host credentials.
	noCredentials credentialSource = iota
	// callerOrHostCredentials presents the caller's credentials when it supplied
	// any, and otherwise lets containers/image discover them from auth files
	// (~/.docker/config.json, ~/.config/containers/auth.json) and credential
	// helpers such as docker-credential-ecr-login, which is needed when Docker is
	// not installed or when using Podman with a private registry like ECR.
	callerOrHostCredentials
)

// credentialContext builds the containers/image context for a credential source.
func credentialContext(source credentialSource, registryUsername, registryPassword string) *types.SystemContext {
	sysCtx := &types.SystemContext{}
	switch source {
	case noCredentials:
		// A non-nil but empty config is what stops the discovery fallback. A nil
		// one silently enables it.
		sysCtx.DockerAuthConfig = &types.DockerAuthConfig{}
	case callerOrHostCredentials:
		if registryUsername != "" || registryPassword != "" {
			sysCtx.DockerAuthConfig = &types.DockerAuthConfig{
				Username: registryUsername,
				Password: registryPassword,
			}
		}
	}
	return sysCtx
}

// OciSha256 gets the digest of a docker/OCI image from its registry, presenting
// the given credentials, or those the host holds when none are given.
func OciSha256(artifactName string, registryUsername string, registryPassword string) (string, error) {
	return ociSha256(artifactName, callerOrHostCredentials, registryUsername, registryPassword)
}

// OciSha256Anonymous gets the digest of a docker/OCI image from its registry
// without presenting any credential, so no credential the host happens to hold
// is offered to a registry the caller does not control.
func OciSha256Anonymous(artifactName string) (string, error) {
	return ociSha256(artifactName, noCredentials, "", "")
}

func ociSha256(artifactName string, source credentialSource, registryUsername, registryPassword string) (string, error) {
	imageName := fmt.Sprintf("//%s", artifactName)
	ctx := context.Background()
	sysCtx := credentialContext(source, registryUsername, registryPassword)

	// Parse image reference
	ref, err := docker.ParseReference(imageName)
	if err != nil {
		if artifactName == " " {
			return "", fmt.Errorf("%w. The artifact name is '%s'. https://docs.kosli.com/faq/#pathimage-name-is-a-single-whitespace-character", err, artifactName)
		}
		return "", fmt.Errorf("failed to parse image reference for %s: %w", imageName, err)
	}

	// Compute digest
	remoteDigest, err := docker.GetDigest(ctx, sysCtx, ref)
	if err != nil {
		return "", fmt.Errorf("failed to get digest for %s: %w", imageName, err)
	}

	fingerprint, err := Sha256Fingerprint(remoteDigest)
	if err != nil {
		return "", fmt.Errorf("registry reported a digest Kosli cannot use for %s: %w", imageName, err)
	}
	return fingerprint, nil
}

// Sha256FingerprintFromDigest turns a registry-supplied digest string into the
// hex fingerprint Kosli uses, rejecting any algorithm other than sha256, because
// a registry chooses the algorithm it answers with.
//
// This is the rule for the OCI and Azure Container Registry lookups. The older
// DockerImageSha256 and RemoteDockerImageSha256 paths still parse digests by
// hand and are not covered by it.
func Sha256FingerprintFromDigest(digestString string) (string, error) {
	parsed, err := godigest.Parse(digestString)
	if err != nil {
		return "", fmt.Errorf("unparseable digest %q: %w", digestString, err)
	}
	return Sha256Fingerprint(parsed)
}

// Sha256Fingerprint is the same rule for a digest that is already parsed, so a
// typed value does not have to be turned back into a string to be checked.
func Sha256Fingerprint(parsed godigest.Digest) (string, error) {
	// godigest.Digest is a string type, so a caller can hand over an unvalidated
	// one and Algorithm()/Encoded() would just split it on the colon.
	if err := parsed.Validate(); err != nil {
		return "", fmt.Errorf("invalid digest %q: %w", parsed.String(), err)
	}
	if parsed.Algorithm() != godigest.SHA256 {
		return "", fmt.Errorf("digest algorithm is %s, but Kosli fingerprints are sha256", parsed.Algorithm())
	}
	// Validate above has checked the charset and length, so the encoded portion
	// is exactly 64 lowercase hex characters here.
	return parsed.Encoded(), nil
}

// ignoreFilePathInTree returns the tree's ignore file as filepath.WalkDir will
// emit it, or "" when the tree has no ignore file.
//
// The name comes from the directory listing rather than from ignoreFileName
// because a case-insensitive filesystem (macOS, Windows) stores whatever name was
// written, a ".KOSLI_IGNORE", while opening it under any case. The walk emits the
// stored name, so this is the exact string the walk will compare, which is what
// makes protecting the file independent of how an exclusion pattern is spelled.
//
// An exact match wins over a folded one to keep the meaning of the ignore file
// stable, not for safety: since the caller reads the rules from whatever this
// returns, either answer would protect the file it read. On a case-sensitive
// filesystem a tree can hold both spellings as two distinct files, and every
// release before this one read the rules from ignoreFileName byte-exact, so
// returning the one that folds first would hand rule authority to a previously
// inert ".KOSLI_IGNORE" and change the fingerprint with it. Other spellings are
// ordinary files there, excludable like any other.
//
// Skipping directories interacts with that: where ignoreFileName is a directory
// and another spelling is a file, the file is returned and its rules apply, where
// before they were not read at all. Contrived, and the alternative is to protect a
// path that can carry no rules.
func ignoreFilePathInTree(dirPath string) (string, error) {
	// Errors are returned rather than swallowed as "no ignore file", because "" is
	// also the answer for a tree that has none, and the caller reads the rules from
	// whatever this returns. A failed read must not be indistinguishable from an
	// absent file: that would fingerprint a tree whose exclusion list was never
	// established.
	if _, err := os.Lstat(filepath.Join(dirPath, ignoreFileName)); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", err
	}
	folded := ""
	for _, entry := range entries {
		// Only a file can carry rules, and protectedPath means "the file whose rules
		// are read". A directory of this name yields no rules, so leaving it
		// unprotected keeps it excludable like any other directory.
		if entry.IsDir() {
			continue
		}
		if entry.Name() == ignoreFileName {
			return filepath.Join(dirPath, entry.Name()), nil
		}
		if folded == "" && strings.EqualFold(entry.Name(), ignoreFileName) {
			folded = filepath.Join(dirPath, entry.Name())
		}
	}
	return folded, nil
}

// resolveExcludePaths expands exclusion patterns, relative to dirPath, into the
// paths they actually match.
func resolveExcludePaths(dirPath string, excludePaths []string) ([]string, error) {
	pathsToExclude := []string{}
	for _, p := range excludePaths {
		found, err := filepathx.Glob(filepath.Join(dirPath, p))
		if err != nil {
			return nil, err
		}
		pathsToExclude = append(pathsToExclude, found...)
	}
	return pathsToExclude, nil
}

// calculateDirContentSha256 calculates a sha256 digest for a directory content
func calculateDirContentSha256(digestsFile *os.File, dirPath, tmpDir string, pathsToExclude []string, protectedPath string, logger *logger.Logger) error {
	return filepath.WalkDir(dirPath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// skip the provided top level dir. Otherwise, the name of that dir is included in
		// the fingerprint calculation (i.e. changing the dir name would change the fingerprint)
		if path == dirPath {
			return nil
		}

		if utils.Contains(pathsToExclude, path) {
			if path == protectedPath {
				logger.Debug("keeping %s although an exclusion matches it: an exclusion list cannot exclude itself. "+
					"Move its entries to --exclude and delete the file to recover the previous fingerprint.", path)
			} else if info.IsDir() {
				logger.Debug("skipping dir %s (and its contents) as it matches excluded paths", path)
				return fs.SkipDir
			} else {
				logger.Debug("skipping %s as it matches excluded paths", path)
				return nil
			}
		}

		// If it's a symlink, resolve the target
		var stat fs.FileInfo
		if info.Type()&os.ModeSymlink != 0 {
			resolved, err := os.Stat(path) // follows the symlink
			if err != nil {
				return err
			}
			stat = resolved
		} else {
			// Convert fs.DirEntry to fs.FileInfo for consistency
			resolved, err := info.Info()
			if err != nil {
				return err
			}
			stat = resolved
		}

		nameSha256, err := addNameDigest(tmpDir, info.Name(), digestsFile, logger)
		if err != nil {
			return err
		}

		if stat.IsDir() {
			if info.Type()&os.ModeSymlink != 0 {
				// This is a symlink to a directory
				logger.Debug("symlink path: %s -- linkname digest: %v", path, nameSha256)

				targetPath, err := os.Readlink(path)
				if err != nil {
					return err
				}

				// Calculate fingerprint of what link points to (for this: a -> c/d calculate the fingerprint of c/d)
				targetSha256, err := addNameDigest(tmpDir, targetPath, digestsFile, logger)
				if err != nil {
					return err
				}
				logger.Debug("symlink: %s (points to %s) -- digest: %v", path, targetPath, targetSha256)
			} else {
				// Normal directory
				logger.Debug("dir path: %s -- dirname digest: %v", path, nameSha256)
			}
		} else {
			// File or symlink -> file
			logger.Debug("file path: %s -- filename digest: %s", path, nameSha256)
			fileContentSha256, err := FileSha256(path, logger)
			if err != nil {
				return err
			}
			logger.Debug("filename: %s -- content digest: %s", path, fileContentSha256)
			if _, err := digestsFile.Write([]byte(fileContentSha256)); err != nil {
				return err
			}
		}

		return nil
	})
}

// addNameDigest calculates the sha256 digest of the filename and adds it to the digests file
func addNameDigest(tmpDir string, filename string, digestsFile *os.File, logger *logger.Logger) (string, error) {
	nameFilePath := filepath.Join(tmpDir, "name")
	err := utils.CreateFileWithContent(nameFilePath, filename)
	if err != nil {
		return "", err
	}

	nameSha256, err := FileSha256(nameFilePath, logger)
	if err != nil {
		return "", err
	}
	if _, err := digestsFile.Write([]byte(nameSha256)); err != nil {
		return "", err
	}
	return nameSha256, nil
}

// FileSha256 returns a sha256 digest of a file.
func FileSha256(filepath string, logger *logger.Logger) (string, error) {
	hasher := sha256.New()
	f, err := os.Open(filepath)
	if err != nil {
		if filepath == " " {
			return "", fmt.Errorf("%s. The filename is '%s'. https://docs.kosli.com/faq/#pathimage-name-is-a-single-whitespace-character", err, filepath)
		}
		return "", err
	}
	defer func() {
		if err := f.Close(); err != nil {
			// Log warning for cleanup error
			logger.Warn("failed to close file %s: %v", filepath, err)
		}
	}()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// DockerImageSha256 returns a sha256 digest of a docker image.
// imageID can be the image name or ID
// It requires the docker daemon to be accessible and the docker image to be locally present.
// The docker image must have been pushed into a registry to have a digest.
func DockerImageSha256(imageID string) (string, error) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return "", err
	}
	imageInspect, err := cli.ImageInspect(context.Background(), imageID)
	if err != nil {
		if imageID == " " {
			return "", fmt.Errorf("%s. The image ID is '%s'. https://docs.kosli.com/faq/#pathimage-name-is-a-single-whitespace-character", err, imageID)
		}
		return "", err
	}
	repoDigests := imageInspect.RepoDigests
	return extractImageDigestFromRepoDigest(imageID, repoDigests)
}

// extractImageDigestFromRepoDigest finds the corresponding digest for an imageName in a list of repoDigests
// imageID can be image name or ID
func extractImageDigestFromRepoDigest(imageID string, repoDigests []string) (string, error) {
	if len(repoDigests) == 0 || imageID == "" {
		return "", ErrRepoDigestUnavailable
	}
	if len(repoDigests) == 1 {
		return strings.Split(repoDigests[0], "@sha256:")[1], nil
	}
	// if the imageID is an ID not a name, there is no way to select
	// a digest from multiple repoDigests entries, so we take the first one
	if err := ValidateDigest(imageID); err == nil && len(repoDigests) > 0 {
		return strings.Split(repoDigests[0], "@sha256:")[1], nil
	}

	// if imageName contains the tag, starts with library or has @sha256, clean it
	imageID = strings.TrimPrefix(imageID, "library/")
	imageID = strings.Split(imageID, ":")[0]
	imageID = strings.TrimSuffix(imageID, "@sha256")

	for _, r := range repoDigests {
		if strings.HasPrefix(r, imageID) {
			return strings.Split(r, "@sha256:")[1], nil
		}
	}
	return "", ErrRepoDigestUnavailable
}

// requestManifestFromRegistry makes an API request to a remote registry to get image manifest
func requestManifestFromRegistry(client *requests.Client, registryEndPoint, imageName, imageTag, registryToken string,
	dockerHeaders map[string]string) (*requests.HTTPResponse, error) {
	url, err := url.JoinPath(registryEndPoint, imageName, "manifests", imageTag)
	if err != nil {
		return nil, fmt.Errorf("failed to build registry URL: %v", err)
	}
	reqParams := &requests.RequestParams{
		Method:            http.MethodGet,
		URL:               url,
		Token:             registryToken,
		AdditionalHeaders: dockerHeaders,
	}
	res, err := client.Do(reqParams)
	if err != nil {
		return res, fmt.Errorf("failed to get docker digest from registry: %v", err)
	}
	return res, nil
}

// RemoteDockerImageSha256 returns a sha256 digest of a docker image by reading
// it from a remote registry. The Accept header advertises every manifest
// media type the OCI Distribution spec defines — Docker single-arch, Docker
// multi-arch ("fat") manifest list, OCI single-arch, and OCI multi-arch
// index — and lets the registry pick the best match in one round trip. The
// canonical sha256 is read from the response's Docker-Content-Digest header,
// which reflects whatever manifest was actually returned.
//
// Callers pass a reusable *requests.Client so a single connection pool can
// be reused across many manifest lookups (e.g. when resolving digests for
// every artifact in a Cloud Run snapshot).
func RemoteDockerImageSha256(client *requests.Client, imageName, imageTag, registryEndPoint, registryToken string) (string, error) {
	acceptHeader := strings.Join([]string{
		"application/vnd.docker.distribution.manifest.list.v2+json",
		"application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.oci.image.index.v1+json",
		"application/vnd.oci.image.manifest.v1+json",
	}, ", ")

	res, err := requestManifestFromRegistry(client, registryEndPoint, imageName, imageTag, registryToken,
		map[string]string{"Accept": acceptHeader})
	if err != nil {
		return "", err
	}

	digestHeader := res.Resp.Header.Get("docker-content-digest")
	return strings.TrimPrefix(digestHeader, "sha256:"), nil
}

const validSha256Pattern = "^([a-f0-9]{64})$"

// validSha256Regexp is compiled once, so ValidateDigest does not re-compile a
// constant pattern on every call.
var validSha256Regexp = regexp.MustCompile(validSha256Pattern)

// ValidateDigest checks if a digest matches the sha256 regex
func ValidateDigest(sha256ToCheck string) error {
	if !validSha256Regexp.MatchString(sha256ToCheck) {
		return fmt.Errorf("%s is not a valid SHA256 fingerprint. It should match the pattern %v", sha256ToCheck, validSha256Pattern)
	}
	return nil
}

func excludePathsFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err == nil {
		defer func() {
			if err := file.Close(); err != nil {
				// Log warning for cleanup error
				fmt.Printf("warning: failed to close file %s: %v\n", path, err)
			}
		}()
		var excludes = []string{}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			line = removeComments(line)
			line = strings.TrimSpace(line)
			if len(line) > 0 {
				excludes = append(excludes, line)
			}
		}
		return excludes, nil
	} else if errors.Is(err, fs.ErrNotExist) {
		return []string{}, nil
	}
	return nil, err
}

func removeComments(line string) string {
	parts := strings.SplitN(line, "#", 2)
	return strings.TrimRight(parts[0], " ")
}
