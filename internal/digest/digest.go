package digest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// DirSha256 returns sha256 digest of a directory
func DirSha256(dirPath string) (string, error) {
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
		pathed_entry := filepath.Join(dirPath, f.Name())

		file, err := os.Create(filepath.Join(tmpDir, "name"))
		if err != nil {
			return err
		}
		if _, err := file.Write([]byte(f.Name())); err != nil {
			return err
		}

		nameSha256, err := FileSha256(filepath.Join(tmpDir, "name"))
		if err != nil {
			return err
		}
		if _, err := digestsFile.Write([]byte(nameSha256)); err != nil {
			return err
		}

		if f.IsDir() {
			err := prepareDirContentSha256(digestsFile, pathed_entry, tmpDir)
			if err != nil {
				return err
			}
		} else {
			fileContentSha256, err := FileSha256(pathed_entry)
			if err != nil {
				return err
			}
			if _, err := digestsFile.Write([]byte(fileContentSha256)); err != nil {
				return err
			}
		}
	}
	return nil
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
