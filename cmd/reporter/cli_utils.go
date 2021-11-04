package main

import (
	"fmt"
	"os"

	"github.com/merkely-development/reporter/internal/digest"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const bitbucket = "Bitbucket"
const github = "Github"
const unknown = "Unknown"

// githubDefaults a map of merkely flags and corresponding default values in Github actions
var githubDefaults = map[string]string{
	"git-commit": os.Getenv("GITHUB_SHA"),
	"commit-url": fmt.Sprintf("%s/%s/commit/%s",
		os.Getenv("GITHUB_SERVER_URL"),
		os.Getenv("GITHUB_REPOSITORY"),
		os.Getenv("GITHUB_SHA")),
	"build-url": fmt.Sprintf("%s/%s/actions/runs/%s",
		os.Getenv("GITHUB_SERVER_URL"),
		os.Getenv("GITHUB_REPOSITORY"),
		os.Getenv("GITHUB_RUN_ID")),
}

// bitbucketDefaults a map of merkely flags and corresponding default values in Bitbucket pipelines
var bitbucketDefaults = map[string]string{
	"git-commit": os.Getenv("BITBUCKET_COMMIT"),
	"commit-url": fmt.Sprintf("https://bitbucket.org/%s/%s/commits/%s",
		os.Getenv("BITBUCKET_WORKSPACE"),
		os.Getenv("BITBUCKET_REPO_SLUG"),
		os.Getenv("BITBUCKET_COMMIT")),
	"build-url": fmt.Sprintf("https://bitbucket.org/%s/%s/addon/pipelines/home#!/results/%s",
		os.Getenv("BITBUCKET_WORKSPACE"),
		os.Getenv("BITBUCKET_REPO_SLUG"),
		os.Getenv("BITBUCKET_BUILD_NUMBER")),
}

// WhichCI detects which CI tool we are in based on env variables
func WhichCI() string {
	if _, ok := os.LookupEnv("BITBUCKET_BUILD_NUMBER"); ok {
		return bitbucket
	} else if _, ok := os.LookupEnv("GITHUB_RUN_NUMBER"); ok {
		return github
	} else {
		return unknown
	}
}

// DefaultValue looks up the default value of a given flag in a given CI tool
func DefaultValue(ci, flag string) string {
	switch ci {
	case github:
		if v, ok := githubDefaults[flag]; ok {
			return v
		}
	case bitbucket:
		if v, ok := bitbucketDefaults[flag]; ok {
			return v
		}
	}
	return ""
}

// RequireFlags decalres a list of flags as required for a given command
func RequireFlags(cmd *cobra.Command, flagNames []string) error {
	for _, name := range flagNames {
		if cmd.Flags().Lookup(name).DefValue == "" {
			err := cobra.MarkFlagRequired(cmd.Flags(), name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// NoArgs returns an error if any args are included.
func NoArgs(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return errors.Errorf(
			"%q accepts no arguments\n\nUsage:  %s",
			cmd.CommandPath(),
			cmd.UseLine(),
		)
	}
	return nil
}

// GetSha256Digest calculates the sha256 digest of an artifact.
// Supported artifact types are: dir, file, docker
func GetSha256Digest(artifactType, name string) (string, error) {
	var err error
	var fingerprint string
	switch artifactType {
	case "file":
		fingerprint, err = digest.FileSha256(name)
	case "dir":
		fingerprint, err = digest.DirSha256(name, false)
	case "docker":
		fingerprint, err = digest.DockerImageSha256(name)
	default:
		return "", fmt.Errorf("%s is not a supported artifact type", artifactType)
	}

	return fingerprint, err
}

func handleError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v", err)
		os.Exit(1)
	}
}
