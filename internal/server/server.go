package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kosli-dev/cli/internal/digest"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/utils"
	"github.com/spf13/viper"
	"github.com/yargevad/filepathx"
)

// ServerEnvRequest represents the PUT request body to be sent to kosli from a server
type ServerEnvRequest struct {
	Artifacts []*ServerData `json:"artifacts"`
}

// ServerData represents the harvested server artifacts data
type ServerData struct {
	Digests           map[string]string `json:"digests"`
	CreationTimestamp int64             `json:"creationTimestamp"`
}

// ArtifactPathSpec represents specification for how to fingerprint an artifact
type ArtifactPathSpec struct {
	Path   string   `mapstructure:"path"`
	Ignore []string `mapstructure:"ignore"`
}

// PathsSpec represents specification for how to fingerprint a list of artifacts
type PathsSpec struct {
	Version   int                         `mapstructure:"version"`
	Artifacts map[string]ArtifactPathSpec `mapstructure:"artifacts"`
}

// CreateServerArtifactsData creates a list of ServerData for server artifacts at given paths
// and excludePaths can contain Glob patterns
// if paths have Glob patterns, each path matching the pattern will be treated as an artifact
func CreateServerArtifactsData(paths, excludePaths []string, logger *logger.Logger) ([]*ServerData, error) {
	result := []*ServerData{}

	pathsToInclude := []string{}
	for _, p := range paths {
		found, err := filepathx.Glob(p)
		if err != nil {
			return []*ServerData{}, err
		}
		pathsToInclude = append(pathsToInclude, found...)
	}

	if len(pathsToInclude) == 0 {
		return []*ServerData{}, fmt.Errorf("no matches found for %v", paths)
	}

	for _, p := range pathsToInclude {
		data, err := getArtifactDataForPath(p, "", excludePaths, logger)
		if err != nil {
			return result, err
		}
		result = append(result, data)
	}
	return result, nil
}

// getArtifactDataForPath calculates the artifact fingerprint for path (while excluding excludePaths)
// and returns a ServerData object
func getArtifactDataForPath(path, artifactName string, excludePaths []string, logger *logger.Logger) (*ServerData, error) {
	data := &ServerData{}
	digests := make(map[string]string)

	finfo, err := os.Stat(path)
	if err != nil {
		return data, fmt.Errorf("failed to open path %s with error: %v", path, err)
	}

	if artifactName == "" {
		artifactName, err = filepath.Abs(path)
		if err != nil {
			return data, fmt.Errorf("failed to get absolute path for %s with error: %v", path, err)
		}
	}

	var fingerprint string
	if !finfo.IsDir() {
		if utils.Contains(excludePaths, path) {
			return data, fmt.Errorf("path [%s] is both required and ignored", path)
		}
		fingerprint, err = digest.FileSha256(path)
	} else {
		fingerprint, err = digest.DirSha256(path, excludePaths, logger)
	}

	if err != nil {
		return data, fmt.Errorf("failed to get a digest of path %s with error: %v", path, err)
	}

	digests[artifactName] = fingerprint
	ts, err := getPathLastModifiedTimestamp(path)
	if err != nil {
		return data, fmt.Errorf("failed to get last modified timestamp of path %s with error: %v", path, err)
	}

	data.Digests = digests
	data.CreationTimestamp = ts
	return data, nil
}

// getPathLastModifiedTimestamp returns the last modified timestamp of a directory
func getPathLastModifiedTimestamp(path string) (int64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fileInfo.ModTime().Unix(), nil
}

// CreatePathsArtifactsData creates a list of ServerData for artifacts as defined in a pathSpecFile
func CreatePathsArtifactsData(pathsSpecFile string, logger *logger.Logger) ([]*ServerData, error) {
	result := []*ServerData{}

	v := viper.New()
	dir, file := filepath.Split(pathsSpecFile)
	file = strings.TrimSuffix(file, filepath.Ext(file))

	// Set the base name of the pathspec file, without the file extension.
	v.SetConfigName(file)

	// Set the dir path where viper should look for the
	// pathspec file. By default, we are looking in the current working directory.
	if dir == "" {
		dir = "."
	}
	v.AddConfigPath(dir)

	if err := v.ReadInConfig(); err != nil {
		return result, fmt.Errorf("failed to parse path spec file [%s] : %v", pathsSpecFile, err)
	}

	var ps PathsSpec
	if err := v.UnmarshalExact(&ps); err != nil {
		return result, fmt.Errorf("failed to unmarshal path spec file [%s] : %v", pathsSpecFile, err)
	}

	for artifactName, pathSpec := range ps.Artifacts {
		logger.Debug("fingerprinting artifact [%s] with spec [ Include: %s, Ignore: %s]", artifactName, pathSpec.Path, pathSpec.Ignore)
		data, err := getArtifactDataForPath(pathSpec.Path, artifactName, pathSpec.Ignore, logger)
		if err != nil {
			return result, fmt.Errorf("failed to calculate fingerprint for artifact [%s]: %v", artifactName, err)
		}

		logger.Debug("fingerprint for artifact [%s]: %s", artifactName, data.Digests[artifactName])
		result = append(result, data)
	}

	return result, nil
}
