package utils

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Contains checks if a string is contained in a string slice
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// LoadFileContent loads file content
func LoadFileContent(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(data), err
}

// IsJSON checks if a string is in JSON format
func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// Creates a file under specified path
func CreateFile(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return nil, err
	}
	return os.Create(p)
}

// IsFile checks if a path is a file
func IsFile(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.Mode().IsRegular(), err
}

// IsDir checks if a path is a directory
func IsDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.Mode().IsDir(), err
}

// Tar creates a tar file from src in a temp directory with the name
// provided in tarFileName. It returns the path of the generated tar file.
func Tar(src, tarFileName string) (string, error) {
	// ensure the src actually exists before trying to tar it
	if _, err := os.Stat(src); err != nil {
		return "", fmt.Errorf("unable to tar file - %v", err.Error())
	}

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}

	tarFilePath := filepath.Join(tmpDir, tarFileName)
	f, err := os.OpenFile(tarFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}

	gzw := gzip.NewWriter(f)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// walk path
	return tarFilePath, filepath.WalkDir(src, func(path string, di fs.DirEntry, err error) error {

		// return on any error
		if err != nil {
			return err
		}

		// return on non-regular files
		if !di.Type().IsRegular() {
			return nil
		}

		fi, err := di.Info()
		if err != nil {
			return err
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// update the name to correctly reflect the desired destination when untaring
		header.Name = strings.TrimPrefix(strings.Replace(path, src, "", -1), string(filepath.Separator))

		// write the header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// open files for taring
		f, err := os.Open(path)
		if err != nil {
			return err
		}

		// copy file data into tar writer
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		f.Close()

		return nil
	})
}

func CreateFileWithContent(path, content string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write([]byte(content))
	return err
}
