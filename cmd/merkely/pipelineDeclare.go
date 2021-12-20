package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

const pipelineDeclareDesc = `
Declare or update a Merkely pipeline by providing a JSON pipefile or by providing pipeline parameters in flags. 
The pipefile contains the pipeline metadata and compliance policy.
`

const pipelineDeclareExample = `
* create a Merkely pipeline with a pipefile:
merkely pipeline declare myPipe --owner owner-name --api-token topSecret --pipefile /path/to/pipefile.json

* The pipefile format is:
{
    "description": "pipeline short description",
    "visibility": "public or private",
    "template": [
        "artifact",
        "evidence-type1",
		"evidence-type2"
    ]
}

* create a Merkely pipeline without a pipefile:
merkely pipeline declare myPipe --description desc \
   --visibility private --template artifact,evidence-type1,evidence-type2 \
   --owner owner-name --api-token topSecret
`

type pipelineDeclareOptions struct {
	pipefile string
	payload  PipelinePayload
}

type PipelinePayload struct {
	Owner       string   `json:"owner"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Visibility  string   `json:"visibility"`
	Template    []string `json:"template"`
}

func newPipelineDeclareCmd(out io.Writer) *cobra.Command {
	o := new(pipelineDeclareOptions)
	cmd := &cobra.Command{
		Use:     "declare PIPELINE-NAME",
		Short:   "Declare a Merkely pipeline",
		Long:    pipelineDeclareDesc,
		Example: pipelineDeclareExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("only pipeline name argument is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return fmt.Errorf("pipeline name argument is required")
			}

			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVar(&o.pipefile, "pipefile", "", "The path to the JSON pipefile.")
	cmd.Flags().StringVar(&o.payload.Description, "description", "", "[optional] The Merkely pipeline description.")
	cmd.Flags().StringVar(&o.payload.Visibility, "visibility", "private", "The visibility of the Merkely pipeline. Options are [public, private].")
	cmd.Flags().StringSliceVarP(&o.payload.Template, "template", "t", []string{"artifact"}, "The comma-separated list of required compliance controls names.")

	return cmd
}

func (o *pipelineDeclareOptions) run(args []string) error {
	pipelineName := args[0]
	url := fmt.Sprintf("%s/api/v1/projects/%s/", global.Host, global.Owner)
	if o.pipefile != "" {
		pipePayload, err := loadPipefile(o.pipefile)
		if err != nil {
			return err
		}
		pipePayload.Name = pipelineName
		pipePayload.Owner = global.Owner
		o.payload.Template = injectArtifactIntoTemplateIfNotExisting(pipePayload.Template)
		_, err = requests.SendPayload(pipePayload, url, "", global.ApiToken,
			global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
		return err
	} else {
		o.payload.Name = pipelineName
		o.payload.Owner = global.Owner
		o.payload.Template = injectArtifactIntoTemplateIfNotExisting(o.payload.Template)
		_, err := requests.SendPayload(o.payload, url, "", global.ApiToken,
			global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
		return err
	}
}

// injectArtifactIntoTemplateIfNotExisting injects 'artifact' into the template if it is not there.
// and cleans any spaces around control names in the template
func injectArtifactIntoTemplateIfNotExisting(template []string) []string {
	found := false
	result := []string{}
	for _, s := range template {
		result = append(result, strings.TrimSpace(s))
		if strings.TrimSpace(s) == "artifact" {
			found = true
		}
	}
	if !found {
		result = append(result, "artifact")
	}
	return result
}

// loadPipefile deserializes a JSON file into a PipelinePayload struct
func loadPipefile(pipefilePath string) (PipelinePayload, error) {
	var pipe PipelinePayload
	jsonFile, err := os.Open(pipefilePath)
	if err != nil {
		return pipe, fmt.Errorf("failed to open pipefile: %v", err)
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)

	err = json.Unmarshal(byteValue, &pipe)
	if err != nil {
		return pipe, fmt.Errorf("failed to unmarshal json: %v", err)
	}
	return pipe, nil
}
