package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const createFlowShortDesc = `Create or update a Kosli flow.`

const createFlowLongDesc = createFlowShortDesc + `
You can provide a JSON pipefile or specify flow parameters in flags. 
The pipefile contains the flow metadata and compliance policy (template).`

const createFlowExample = `
# create/update a Kosli flow without a pipefile:
kosli create flow \
	--flow yourFlowName \
	--description yourFlowDescription \
    --visibility private OR public \
	--template artifact,evidence-type1,evidence-type2 \
	--api-token yourAPIToken \
	--owner yourOrgName

# create/update a Kosli flow with a pipefile (this is a legacy way which will be removed in the future):
kosli create flow \
	--pipefile /path/to/pipefile.json \
	--api-token yourAPIToken \
	--owner yourOrgName

The pipefile format is:
{
    "name": "yourFlowName",
    "description": "yourFlowDescription",
    "visibility": "public or private",
    "template": [
        "artifact",
        "evidence-type1",
        "evidence-type2"
    ]
}
`

type createFlowOptions struct {
	pipefile string
	payload  FlowPayload
}

type FlowPayload struct {
	Owner       string   `json:"owner"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Visibility  string   `json:"visibility"`
	Template    []string `json:"template"`
}

func newCreateFlowCmd(out io.Writer) *cobra.Command {
	o := new(createFlowOptions)
	cmd := &cobra.Command{
		Use:     "flow",
		Short:   createFlowShortDesc,
		Long:    createFlowLongDesc,
		Example: createFlowExample,
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = MuXRequiredFlags(cmd, []string{"flow", "pipefile"}, true)
			if err != nil {
				return err
			}

			err = MuXRequiredFlags(cmd, []string{"description", "pipefile"}, false)
			if err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run()
		},
	}

	cmd.Flags().StringVar(&o.payload.Name, "flow", "", newFlowFlag)
	cmd.Flags().StringVar(&o.pipefile, "pipefile", "", pipefileFlag)
	cmd.Flags().StringVar(&o.payload.Description, "description", "", pipelineDescriptionFlag)
	cmd.Flags().StringVar(&o.payload.Visibility, "visibility", "private", visibilityFlag)
	cmd.Flags().StringSliceVarP(&o.payload.Template, "template", "t", []string{"artifact"}, templateFlag)
	addDryRunFlag(cmd)

	return cmd
}

func (o *createFlowOptions) run() error {
	var err error
	url := fmt.Sprintf("%s/api/v1/projects/%s/", global.Host, global.Owner)
	if o.pipefile != "" {
		o.payload, err = loadPipefile(o.pipefile)
		if err != nil {
			return err
		}
	}
	o.payload.Owner = global.Owner
	o.payload.Template = injectArtifactIntoTemplateIfNotExisting(o.payload.Template)

	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Payload:  o.payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("flow '%s' was created", o.payload.Name)
	}
	return err
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
func loadPipefile(pipefilePath string) (FlowPayload, error) {
	var pipe FlowPayload
	jsonFile, err := os.Open(pipefilePath)
	if err != nil {
		return pipe, fmt.Errorf("failed to open pipefile: %v", err)
	}
	byteValue, _ := io.ReadAll(jsonFile)

	err = json.Unmarshal(byteValue, &pipe)
	if err != nil {
		return pipe, fmt.Errorf("failed to unmarshal json: %v", err)
	}
	return pipe, nil
}
