package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/merkely-development/reporter/internal/digest"
	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

type evidenceOptions struct {
	artifactType string
	inputSha256  string
	sha256       string // This is calculated or provided by the user
	pipelineName string
	description  string
	isCompliant  bool
	buildUrl     string
	userDataFile string
	payload      EvidencePayload
}

type EvidencePayload struct {
	EvidenceType string                 `json:"evidence_type"`
	Contents     map[string]interface{} `json:"contents"`
}

func newEvidenceCmd(out io.Writer) *cobra.Command {
	o := new(evidenceOptions)
	cmd := &cobra.Command{
		Use:   "evidence ARTIFACT-NAME-OR-PATH",
		Short: "Report/Log an evidence to an artifact in Merkely. ",
		Long:  evidenceDesc(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("only one argument (docker image name or file/dir path) is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return fmt.Errorf("docker image name or file/dir path is required")
			}

			if o.artifactType == "" && o.inputSha256 == "" {
				return fmt.Errorf("either --type or --sha256 must be specified")
			}

			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return err
			}

			if o.inputSha256 != "" {
				if err := digest.ValidateDigest(o.inputSha256); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if o.inputSha256 != "" {
				o.sha256 = o.inputSha256
			} else {
				o.sha256, err = GetSha256Digest(o.artifactType, args[0])
				if err != nil {
					return err
				}
			}

			url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s", global.Host, global.Owner, o.pipelineName, o.sha256)
			o.payload.Contents = map[string]interface{}{}
			o.payload.Contents["is_compliant"] = o.isCompliant
			o.payload.Contents["url"] = o.buildUrl
			o.payload.Contents["description"] = o.description
			o.payload.Contents["user_data"], err = LoadUserData(o.userDataFile)
			if err != nil {
				return err
			}
			js, _ := json.MarshalIndent(o.payload, "", "    ")

			return requests.SendPayload(js, url, global.ApiToken,
				global.MaxAPIRetries, global.DryRun, "PUT", log)
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVarP(&o.artifactType, "artifact-type", "t", "", "The type of the artifact related to the evidence. Options are [dir, file, docker].")
	cmd.Flags().StringVarP(&o.inputSha256, "sha256", "s", "", "The SHA256 fingerprint for the artifact. Only required if you don't specify --type.")
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", "The Merkely pipeline name.")
	cmd.Flags().StringVarP(&o.description, "description", "d", "", "[optional] The evidence description.")
	cmd.Flags().StringVarP(&o.buildUrl, "build-url", "b", DefaultValue(ci, "build-url"), "The url of CI pipeline that generated the evidence.")
	cmd.Flags().BoolVarP(&o.isCompliant, "compliant", "C", true, "Whether the evidence is compliant or not.")
	cmd.Flags().StringVarP(&o.payload.EvidenceType, "evidence-type", "e", "", "The type of evidence being reported.")
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", "The path to a JSON file containing additional data you would like to attach to this evidence.")

	err := RequireFlags(cmd, []string{"pipeline", "build-url", "evidence-type"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func evidenceDesc() string {
	return `
   Report an evidence to an artifact in Merkely. 
   The artifact SHA256 fingerprint is calculated or alternatively it can be provided directly. 
   ` + GetCIDefaultsTemplates(supportedCIs, []string{"build-url"})
}
