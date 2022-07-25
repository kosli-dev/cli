package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const pipelineDeclareDesc = `
Declare or update a Kosli pipeline by providing a JSON pipefile or by providing pipeline parameters in flags. 
The pipefile contains the pipeline metadata and compliance policy.
`

const pipelineDeclareExample = `
# create/update a Kosli pipeline without a pipefile:
kosli pipeline declare \
	--pipeline yourPipelineName \
	--description yourPipelineDescription \
    --visibility private OR public \
	--template artifact,evidence-type1,evidence-type2 \
	--api-token yourAPIToken \
	--owner yourOrgName

# create/update a Kosli pipeline with a pipefile (this is a legacy way which will be removed in the future):
kosli pipeline declare \
	--pipefile /path/to/pipefile.json \
	--api-token yourAPIToken \
	--owner yourOrgName

The pipefile format is:
{
    "name": "yourPipelineName",
    "description": "yourPipelinedescription",
    "visibility": "public or private",
    "template": [
        "artifact",
        "evidence-type1",
        "evidence-type2"
    ]
}
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
		Use:     "declare",
		Short:   "Declare or update a Kosli pipeline",
		Long:    pipelineDeclareDesc,
		Example: pipelineDeclareExample,
		Args:    NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorAfterPrintingHelp(cmd, err.Error())
			}

			if o.pipefile != "" {
				// This check does not catch if --template or --visibility is provided by the user
				// as they both have defaults.
				// When a pipefile is provided, the flags are ignored anyway
				if o.payload.Description != "" || o.payload.Name != "" {
					return ErrorAfterPrintingHelp(cmd, "--pipefile cannot be used together with any of"+
						" --description, --pipeline flags")
				}
			} else {
				if o.payload.Name == "" {
					return ErrorAfterPrintingHelp(cmd, "--pipeline is required when you are not using --pipefile")
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVar(&o.payload.Name, "pipeline", "", newPipelineFlag)
	cmd.Flags().StringVar(&o.pipefile, "pipefile", "", pipefileFlag)
	cmd.Flags().StringVar(&o.payload.Description, "description", "", pipelineDescriptionFlag)
	cmd.Flags().StringVar(&o.payload.Visibility, "visibility", "private", visibilityFlag)
	cmd.Flags().StringSliceVarP(&o.payload.Template, "template", "t", []string{"artifact"}, templateFlag)

	return cmd
}

func (o *pipelineDeclareOptions) run(args []string) error {
	url := fmt.Sprintf("%s/api/v1/projects/%s/", global.Host, global.Owner)
	if o.pipefile != "" {
		pipePayload, err := loadPipefile(o.pipefile)
		if err != nil {
			return err
		}
		pipePayload.Owner = global.Owner
		o.payload.Template = injectArtifactIntoTemplateIfNotExisting(pipePayload.Template)
		_, err = requests.SendPayload(pipePayload, url, "", global.ApiToken,
			global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
		return err
	} else {
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
