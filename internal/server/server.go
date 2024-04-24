package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kosli-dev/cli/internal/digest"
	"github.com/kosli-dev/cli/internal/logger"
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

// CreateServerArtifactsData creates a list of ServerData for server artifacts at given paths
// paths and excludePaths can contain Glob patterns
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
		digests := make(map[string]string)

		finfo, err := os.Stat(p)
		if err != nil {
			return []*ServerData{}, fmt.Errorf("failed to open path %s with error: %v", p, err)
		}
		var fingerprint string
		if !finfo.IsDir() {
			fingerprint, err = digest.FileSha256(p)
		} else {
			fingerprint, err = digest.DirSha256(p, excludePaths, logger)
		}

		if err != nil {
			return []*ServerData{}, fmt.Errorf("failed to get a digest of path %s with error: %v", p, err)
		}
		artifactName, err := filepath.Abs(p)
		if err != nil {
			return []*ServerData{}, fmt.Errorf("failed to get absolute path for %s with error: %v", p, err)
		}
		digests[artifactName] = fingerprint
		ts, err := getPathLastModifiedTimestamp(p)
		if err != nil {
			return []*ServerData{}, fmt.Errorf("failed to get last modified timestamp of path %s with error: %v", p, err)
		}
		result = append(result, &ServerData{Digests: digests, CreationTimestamp: ts})
	}
	return result, nil
}

// CreateServerArtifactsData creates a list of ServerData for server artifacts at given paths
// paths and excludePaths can contain Glob patterns
// if paths have Glob patterns, each path matching the pattern will be treated as an artifact
func FingerprintPaths(artifacName string, paths, excludePaths []string, logger *logger.Logger) (*ServerData, error) {
	result := &ServerData{}

	pathsToInclude := []string{}
	for _, p := range paths {
		found, err := filepathx.Glob(p)
		if err != nil {
			return result, err
		}
		pathsToInclude = append(pathsToInclude, found...)
	}

	if len(pathsToInclude) == 0 {
		return result, fmt.Errorf("no matches found for %v", paths)
	}

	fingerprintAfterEachOther := ""
	for _, p := range pathsToInclude {
		// digests := make(map[string]string)

		finfo, err := os.Stat(p)
		if err != nil {
			return result, fmt.Errorf("failed to open path %s with error: %v", p, err)
		}
		var fingerprint string
		if !finfo.IsDir() {
			fingerprint, err = digest.FileSha256(p)
		} else {
			fingerprint, err = digest.DirSha256(p, excludePaths, logger)
		}
		if err != nil {
			return result, fmt.Errorf("failed to get a digest of path %s with error: %v", p, err)
		}
		fingerprintAfterEachOther += fingerprint
		fmt.Printf("%s\n", fingerprintAfterEachOther)

		// artifactName, err := filepath.Abs(p)
		// if err != nil {
		// 	return result, fmt.Errorf("failed to get absolute path for %s with error: %v", p, err)
		// }
		// digests[artifactName] = fingerprint
		// ts, err := getPathLastModifiedTimestamp(p)
		// if err != nil {
		// 	return result, fmt.Errorf("failed to get last modified timestamp of path %s with error: %v", p, err)
		// }
		// result = append(result, &ServerData{Digests: digests, CreationTimestamp: ts})
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

// # koslipaths.yaml
// artifacts:
//   differ:
//     include-paths: [differ/bin, differ/lib, koslipaths.yaml]
//     exclude-paths: [differ/bin/ls]
//   runner:
//     include-paths: [runner/bin]

type ArtifactSpec struct {
	Include []string `mapstructure:"include-paths"`
	Exclude []string `mapstructure:"exclude-paths"`
}

type PathSpec struct {
	Artifacts map[string]ArtifactSpec `mapstructure:"artifacts"`
}

// CreatePathsArtifactsData
func CreatePathsArtifactsData(pathSpecFile string, logger *logger.Logger) ([]*ServerData, error) {
	result := []*ServerData{}

	v := viper.New()
	dir, file := filepath.Split(pathSpecFile)
	file = strings.TrimSuffix(file, filepath.Ext(file))

	// Set the base name of the pathspec file, without the file extension.
	v.SetConfigName(file)

	// Set as many paths as you like where viper should look for the
	// pathspec file. By default, we are looking in the current working directory.
	if dir == "" {
		dir = "."
	}
	v.AddConfigPath(dir)

	if err := v.ReadInConfig(); err != nil {
		return result, fmt.Errorf("failed to parse path spec file [%s] : %v", pathSpecFile, err)
	}

	var ps PathSpec
	if err := v.UnmarshalExact(&ps); err != nil {
		return result, fmt.Errorf("failed to unmarshal path spec file [%s] : %v", pathSpecFile, err)
	}

	for artifactName, paths := range ps.Artifacts {
		fmt.Printf("%s with %v\n", artifactName, paths)
		res, err := FingerprintPaths(artifactName, paths.Include, paths.Exclude, logger)
		if err != nil {
			return result, err
		}
		result = append(result, res)
	}

	fmt.Println(result[0])

	// pathsToInclude := []string{}
	// for _, p := range paths {
	// 	found, err := filepathx.Glob(p)
	// 	if err != nil {
	// 		return []*ServerData{}, err
	// 	}
	// 	pathsToInclude = append(pathsToInclude, found...)
	// }

	// if len(pathsToInclude) == 0 {
	// 	return []*ServerData{}, fmt.Errorf("no matches found for %v", paths)
	// }

	// for _, p := range pathsToInclude {
	// 	digests := make(map[string]string)

	// 	finfo, err := os.Stat(p)
	// 	if err != nil {
	// 		return []*ServerData{}, fmt.Errorf("failed to open path %s with error: %v", p, err)
	// 	}
	// 	var fingerprint string
	// 	if !finfo.IsDir() {
	// 		fingerprint, err = digest.FileSha256(p)
	// 	} else {
	// 		fingerprint, err = digest.DirSha256(p, excludePaths, logger)
	// 	}

	// 	if err != nil {
	// 		return []*ServerData{}, fmt.Errorf("failed to get a digest of path %s with error: %v", p, err)
	// 	}
	// 	artifactName, err := filepath.Abs(p)
	// 	if err != nil {
	// 		return []*ServerData{}, fmt.Errorf("failed to get absolute path for %s with error: %v", p, err)
	// 	}
	// 	digests[artifactName] = fingerprint
	// 	ts, err := getPathLastModifiedTimestamp(p)
	// 	if err != nil {
	// 		return []*ServerData{}, fmt.Errorf("failed to get last modified timestamp of path %s with error: %v", p, err)
	// 	}
	// 	result = append(result, &ServerData{Digests: digests, CreationTimestamp: ts})
	// }
	return result, nil
}
