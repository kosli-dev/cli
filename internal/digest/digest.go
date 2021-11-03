package digest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/client"
)

var verboseFlag bool

func verboseLogs(format string, a ...interface{}) {
	if verboseFlag {
		fmt.Println(fmt.Sprintf(format, a...))
	}
}

// DirSha256 returns sha256 digest of a directory
func DirSha256(dirPath string, verbose bool) (string, error) {
	verboseFlag = verbose
	verboseLogs("Input path: %v", filepath.Base(dirPath))
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
	err = prepareDirContentSha256(digestsFile, dirPath, tmpDir)
	if err != nil {
		return "", err
	}

	return FileSha256(digestsFile.Name())
}

// prepareDirContentSha256 calculates a sha256 digest for a directory content
func prepareDirContentSha256(digestsFile *os.File, dirPath, tmpDir string) error {

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
			verboseLogs("dirname: %s -- dirname digest: %v", pathed_entry, nameSha256)
			err := prepareDirContentSha256(digestsFile, pathed_entry, tmpDir)
			if err != nil {
				return err
			}
		} else {
			verboseLogs("filename: %s -- filename digest: %s", pathed_entry, nameSha256)
			fileContentSha256, err := FileSha256(pathed_entry)
			if err != nil {
				return err
			}
			verboseLogs("filename: %s -- content digest: %s", pathed_entry, fileContentSha256)
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

// DockerImageSha256 returns a sha256 digest of a docker image. It requires
// the docker deamon to be accessible.
// The docker image must have been pushed into a registry to have a digest.
func DockerImageSha256(imageName string) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}
	imageInspect, _, err := cli.ImageInspectWithRaw(context.Background(), imageName)
	if err != nil {
		return "", err
	}
	repoDigests := imageInspect.RepoDigests
	if len(repoDigests) > 0 {
		fingerprint := strings.Split(repoDigests[0], "@sha256:")[1]
		return fingerprint, nil
	} else {
		return "", fmt.Errorf("failed to get a digest for the image, has it been pushed to a registry?")
	}
}
