package main

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

type artifactCreationOptions struct {
	artifactType string
	pipelineName string
	payload      ArtifactPayload
}

type ArtifactPayload struct {
	Sha256      string `json:"sha256"`
	Filename    string `json:"filename"`
	Description string `json:"description"`
	GitCommit   string `json:"git_commit"`
	IsCompliant bool   `json:"is_compliant"`
	BuildUrl    string `json:"build_url"`
	CommitUrl   string `json:"commit_url"`
}

func newArtifactCreationCmd(out io.Writer) *cobra.Command {
	o := new(artifactCreationOptions)
	cmd := &cobra.Command{
		Use:   "creation ARTIFACT-NAME-OR-PATH",
		Short: "Report an artifact creation to a Merkely pipeline. ",
		Long:  artifactCreationDesc(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return err
			}

			return ValidateArtifactArg(args, o.artifactType, o.payload.Sha256)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVarP(&o.artifactType, "artifact-type", "t", "", "The type of the artifact. Options are [dir, file, docker].")
	cmd.Flags().StringVarP(&o.payload.Sha256, "sha256", "s", "", "The SHA256 fingerprint for the artifact. Only required if you don't specify --artifact-type.")
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", "The Merkely pipeline name.")
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", "[optional] The artifact description.")
	cmd.Flags().StringVarP(&o.payload.GitCommit, "git-commit", "g", DefaultValue(ci, "git-commit"), "The git commit from which the artifact was created.")
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), "The url of CI pipeline that built the artifact.")
	cmd.Flags().StringVarP(&o.payload.CommitUrl, "commit-url", "u", DefaultValue(ci, "commit-url"), "The url for the git commit that created the artifact.")
	cmd.Flags().BoolVarP(&o.payload.IsCompliant, "compliant", "C", true, "Whether the artifact is compliant or not.")

	err := RequireFlags(cmd, []string{"pipeline", "git-commit", "build-url", "commit-url"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *artifactCreationOptions) run(args []string) error {
	if o.payload.Sha256 != "" {
		o.payload.Filename = args[0]
	} else {
		var err error
		o.payload.Sha256, err = GetSha256Digest(o.artifactType, args[0])
		if err != nil {
			return err
		}
		if o.artifactType == "dir" || o.artifactType == "file" {
			o.payload.Filename = filepath.Base(args[0])
		} else {
			o.payload.Filename = args[0]
		}
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/", global.Host, global.Owner, o.pipelineName)

	_, err := requests.SendPayload(o.payload, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
	return err
}

func artifactCreationDesc() string {
	return `
   Report an artifact creation to a pipeline in Merkely. 
   The artifact SHA256 fingerprint is calculated and reported 
   or,alternatively, can be provided directly. 
   ` + GetCIDefaultsTemplates(supportedCIs, []string{"git-commit", "build-url", "commit-url"})
}
