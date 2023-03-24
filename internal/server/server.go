package server

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kosli-dev/cli/internal/digest"
	"github.com/kosli-dev/cli/internal/logger"
)

// ServerEnvRequest represents the PUT request body to be sent to kosli from a server
type ServerEnvRequest struct {
	Artifacts []*ServerData `json:"artifacts"`
	Type      string        `json:"type"`
}

// ServerData represents the harvested server artifacts data
type ServerData struct {
	Digests           map[string]string `json:"digests"`
	CreationTimestamp int64             `json:"creationTimestamp"`
}

// CreateServerArtifactsData creates a list of ServerData for server artifacts at given paths
func CreateServerArtifactsData(paths []string, logger *logger.Logger) ([]*ServerData, error) {
	result := []*ServerData{}
	for _, p := range paths {
		digests := make(map[string]string)

		finfo, err := os.Stat(p)
		if err != nil {
			return []*ServerData{}, fmt.Errorf("failed to open path %s with error: %v", p, err)
		}
		var sha256 string
		if !finfo.IsDir() {
			sha256, err = digest.FileSha256(p)
		} else {
			sha256, err = digest.DirSha256(p, logger)
		}

		if err != nil {
			return []*ServerData{}, fmt.Errorf("failed to get a digest of path %s with error: %v", p, err)
		}
		artifactName, err := filepath.Abs(p)
		if err != nil {
			return []*ServerData{}, fmt.Errorf("failed to get absolute path for %s with error: %v", p, err)
		}
		digests[artifactName] = sha256
		ts, err := getPathLastModifiedTimestamp(p)
		if err != nil {
			return []*ServerData{}, fmt.Errorf("failed to get last modified timestamp of path %s with error: %v", p, err)
		}
		result = append(result, &ServerData{Digests: digests, CreationTimestamp: ts})
	}
	return result, nil
}

// getPathLastModifiedTimestamp returns the last modified timestamp of a directory
func getPathLastModifiedTimestamp(path string) (int64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fileInfo.ModTime().Unix(), nil
}
