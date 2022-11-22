package digest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/docker/docker/client"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/sirupsen/logrus"
)

var (
	// ErrRepoDigestUnavailable returned when repo digest is not available.
	ErrRepoDigestUnavailable = errors.New("repo digest unavailable for the image, " +
		"has it been pushed to or pulled from a registry?")
)

// DirSha256 returns sha256 digest of a directory
func DirSha256(dirPath string, logger *logrus.Logger) (string, error) {
	logger.Debugf("Input path: %v", filepath.Base(dirPath))
	info, err := os.Stat(dirPath)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", dirPath)
	}

	tmpDir, err := ioutil.TempDir("", "*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	digestsFile, err := os.Create(filepath.Join(tmpDir, "digests"))
	if err != nil {
		return "", err
	}
	defer digestsFile.Close()
	err = prepareDirContentSha256(digestsFile, dirPath, tmpDir, logger)
	if err != nil {
		return "", err
	}

	return FileSha256(digestsFile.Name())
}

// prepareDirContentSha256 calculates a sha256 digest for a directory content
func prepareDirContentSha256(digestsFile *os.File, dirPath, tmpDir string, logger *logrus.Logger) error {

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, f := range files {
		nameSha256, err := addNameDigest(tmpDir, f.Name(), digestsFile)
		if err != nil {
			return err
		}

		pathed_entry := filepath.Join(dirPath, f.Name())

		if f.IsDir() {
			logger.Debugf("dirname: %s -- dirname digest: %v", pathed_entry, nameSha256)
			err := prepareDirContentSha256(digestsFile, pathed_entry, tmpDir, logger)
			if err != nil {
				return err
			}
		} else {
			logger.Debugf("filename: %s -- filename digest: %s", pathed_entry, nameSha256)
			fileContentSha256, err := FileSha256(pathed_entry)
			if err != nil {
				return err
			}
			logger.Debugf("filename: %s -- content digest: %s", pathed_entry, fileContentSha256)
			if _, err := digestsFile.Write([]byte(fileContentSha256)); err != nil {
				return err
			}
		}
	}
	return nil
}

// addNameDigest calculates the sha256 digest of the filename and adds it to the digests file
func addNameDigest(tmpDir string, filename string, digestsFile *os.File) (string, error) {
	file, err := os.Create(filepath.Join(tmpDir, "name"))
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := file.Write([]byte(filename)); err != nil {
		return "", err
	}

	nameSha256, err := FileSha256(filepath.Join(tmpDir, "name"))
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
	imageInspect, _, err := cli.ImageInspectWithRaw(context.Background(), imageID)
	if err != nil {
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
func requestManifestFromRegistry(registryEndPoint, imageName, imageTag, registryToken string, dockerHeaders map[string]string) (*requests.HTTPResponse, error) {
	res, err := requests.DoRequestWithToken([]byte{}, registryEndPoint+"/"+imageName+"/"+"manifests/"+imageTag, registryToken, 3, http.MethodGet, dockerHeaders, logrus.New())

	if err != nil {
		return res, fmt.Errorf("failed to get docker digest from registry %v", err)
	}
	return res, nil

}

// RemoteDockerImageSha256 returns a sha256 digest of a docker image by reading it from
// remote docker registry
func RemoteDockerImageSha256(imageName, imageTag, registryEndPoint, registryToken string) (string, error) {
	// Some docker images have Manifest list, aka “fat manifest” which combines
	// image manifests for one or more platforms. Other images don't have such manifest.
	// The response Content-Type header specifies whether an image has it or not.
	// More details here: https://docs.docker.com/registry/spec/manifest-v2-2/
	v2ManifestType := "application/vnd.docker.distribution.manifest.v2+json"
	v2FatManifestType := "application/vnd.docker.distribution.manifest.list.v2+json"

	var res *requests.HTTPResponse
	var err error

	dockerHeaders := map[string]string{"Accept": v2FatManifestType}
	res, err = requestManifestFromRegistry(registryEndPoint, imageName, imageTag, registryToken, dockerHeaders)
	if err != nil {
		return "", err
	}

	responseContentType := res.Resp.Header.Get("Content-Type")
	if responseContentType != v2FatManifestType {
		dockerHeaders = map[string]string{"Accept": v2ManifestType}

		res, err = requestManifestFromRegistry(registryEndPoint, imageName, imageTag, registryToken, dockerHeaders)
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
		return fmt.Errorf("failed to validate the provided SHA256 digest")
	}
	if !r.MatchString(sha256ToCheck) {
		return fmt.Errorf("%s is not a valid SHA256 digest. It should match the pattern %v", sha256ToCheck, validSha256regex)
	}
	return nil
}
