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
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/types"
	"github.com/docker/docker/client"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/utils"
	"github.com/yargevad/filepathx"
)

var (
	// ErrRepoDigestUnavailable returned when repo digest is not available.
	ErrRepoDigestUnavailable = errors.New("repo digest unavailable for the image, " +
		"has it been pushed to or pulled from a registry?")
)

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
	defer os.RemoveAll(tmpDir)

	digestsFile, err := os.Create(filepath.Join(tmpDir, "digests"))
	if err != nil {
		return "", err
	}
	defer digestsFile.Close()
	ignoreFilePath := filepath.Join(dirPath, ".kosli_ignore")
	ignoredPaths, err := excludePathsFromFile(ignoreFilePath)
	if err != nil {
		return "", err
	}
	if len(ignoredPaths) > 0 {
		logger.Debug("  -> ignore file used %s -- excluding paths: %s", ignoreFilePath, ignoredPaths)
	}
	excludePaths = append(excludePaths, ignoredPaths...)
	err = calculateDirContentSha256(digestsFile, dirPath, tmpDir, excludePaths, logger)
	if err != nil {
		return "", err
	}

	return FileSha256(digestsFile.Name())
}

// OciSha256 gets the digest of a docker/OCI image from its registry
func OciSha256(artifactName string, registryUsername string, registryPassword string) (string, error) {
	imageName := fmt.Sprintf("//%s", artifactName)
	ctx := context.Background()
	sysCtx := &types.SystemContext{
		DockerAuthConfig: &types.DockerAuthConfig{
			Username: registryUsername,
			Password: registryPassword,
		},
	}

	// Parse image reference
	ref, err := docker.ParseReference(imageName)
	if err != nil {
		if artifactName == " " {
			return "", fmt.Errorf("%w. The artifact name is '%s'. https://docs.kosli.com/faq/#pathimage-name-is-a-single-whitespace-character", err, artifactName)
		}
		return "", fmt.Errorf("failed to parse image reference for %s: %w", imageName, err)
	}

	// Compute digest
	digest, err := docker.GetDigest(ctx, sysCtx, ref)
	if err != nil {
		return "", fmt.Errorf("failed to get digest for %s: %w", imageName, err)
	}
	return strings.Split(digest.String(), "sha256:")[1], nil
}

// calculateDirContentSha256 calculates a sha256 digest for a directory content
func calculateDirContentSha256(digestsFile *os.File, dirPath, tmpDir string, excludePaths []string, logger *logger.Logger) error {
	pathsToExclude := []string{}
	for _, p := range excludePaths {
		found, err := filepathx.Glob(filepath.Join(dirPath, p))
		if err != nil {
			return err
		}
		pathsToExclude = append(pathsToExclude, found...)
	}

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
			if info.IsDir() {
				logger.Debug("skipping dir %s (and its contents) as it matches excluded paths", path)
				return fs.SkipDir
			}
			logger.Debug("skipping %s as it matches excluded paths", path)
			return nil
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

		nameSha256, err := addNameDigest(tmpDir, info.Name(), digestsFile)
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
				targetSha256, err := addNameDigest(tmpDir, targetPath, digestsFile)
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
			fileContentSha256, err := FileSha256(path)
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
func addNameDigest(tmpDir string, filename string, digestsFile *os.File) (string, error) {
	nameFilePath := filepath.Join(tmpDir, "name")
	err := utils.CreateFileWithContent(nameFilePath, filename)
	if err != nil {
		return "", err
	}

	nameSha256, err := FileSha256(nameFilePath)
	if err != nil {
		return "", err
	}
	if _, err := digestsFile.Write([]byte(nameSha256)); err != nil {
		return "", err
	}
	return nameSha256, nil
}

// FileSha256 returns a sha256 digest of a file.
func FileSha256(filepath string) (string, error) {
	hasher := sha256.New()
	f, err := os.Open(filepath)
	if err != nil {
		if filepath == " " {
			return "", fmt.Errorf("%s. The filename is '%s'. https://docs.kosli.com/faq/#pathimage-name-is-a-single-whitespace-character", err, filepath)
		}
		return "", err
	}
	defer f.Close()
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
	cli, err := client.NewClientWithOpts(client.FromEnv)
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
func requestManifestFromRegistry(registryEndPoint, imageName, imageTag, registryToken string,
	dockerHeaders map[string]string, logger *logger.Logger) (*requests.HTTPResponse, error) {
	// res, err := requests.DoRequestWithToken([]byte{}, registryEndPoint+"/"+imageName+"/"+"manifests/"+imageTag, registryToken, 3, http.MethodGet, dockerHeaders)
	url := registryEndPoint + "/" + imageName + "/" + "manifests/" + imageTag
	reqParams := &requests.RequestParams{
		Method:            http.MethodGet,
		URL:               url,
		Token:             registryToken,
		AdditionalHeaders: dockerHeaders,
	}
	kosliClient, err := requests.NewKosliClient("", 1, logger.DebugEnabled, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get docker digest from registry: %v", err)
	}
	res, err := kosliClient.Do(reqParams)
	if err != nil {
		return res, fmt.Errorf("failed to get docker digest from registry: %v", err)
	}
	return res, nil

}

// RemoteDockerImageSha256 returns a sha256 digest of a docker image by reading it from
// remote docker registry
func RemoteDockerImageSha256(imageName, imageTag, registryEndPoint, registryToken string, logger *logger.Logger) (string, error) {
	// Some docker images have Manifest list, aka “fat manifest” which combines
	// image manifests for one or more platforms. Other images don't have such manifest.
	// The response Content-Type header specifies whether an image has it or not.
	// More details here: https://docs.docker.com/registry/spec/manifest-v2-2/
	v2ManifestType := "application/vnd.docker.distribution.manifest.v2+json"
	v2FatManifestType := "application/vnd.docker.distribution.manifest.list.v2+json"

	var res *requests.HTTPResponse
	var err error

	dockerHeaders := map[string]string{"Accept": v2FatManifestType}
	res, err = requestManifestFromRegistry(registryEndPoint, imageName, imageTag, registryToken, dockerHeaders, logger)
	if err != nil {
		return "", err
	}

	responseContentType := res.Resp.Header.Get("Content-Type")
	if responseContentType != v2FatManifestType {
		dockerHeaders = map[string]string{"Accept": v2ManifestType}

		res, err = requestManifestFromRegistry(registryEndPoint, imageName, imageTag, registryToken, dockerHeaders, logger)
		if err != nil {
			return "", err
		}
	}

	digestHeader := res.Resp.Header.Get("docker-content-digest")

	fingerprint := strings.TrimPrefix(digestHeader, "sha256:")
	return fingerprint, nil
}

// ValidateDigest checks if a digest matches the sha256 regex
func ValidateDigest(sha256ToCheck string) error {
	validSha256regex := "^([a-f0-9]{64})$"
	r, err := regexp.Compile(validSha256regex)
	if err != nil {
		return fmt.Errorf("failed to validate the provided SHA256 fingerprint")
	}
	if !r.MatchString(sha256ToCheck) {
		return fmt.Errorf("%s is not a valid SHA256 fingerprint. It should match the pattern %v", sha256ToCheck, validSha256regex)
	}
	return nil
}

func excludePathsFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err == nil {
		defer file.Close()
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
