package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const createFlowShortDesc = `Create or update a Kosli flow.`

const createFlowLongDesc = createFlowShortDesc + `
You can specify flow parameters in flags.`

const createFlowExample = `
# create/update a Kosli flow:
kosli create flow yourFlowName \
	--description yourFlowDescription \
    --visibility private OR public \
	--template artifact,evidence-type1,evidence-type2 \
	--api-token yourAPIToken \
	--org yourOrgName
`

type createFlowOptions struct {
	payload FlowPayload
}

type FlowPayload struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Visibility  string   `json:"visibility"`
	Template    []string `json:"template"`
}

func newCreateFlowCmd(out io.Writer) *cobra.Command {
	o := new(createFlowOptions)
	cmd := &cobra.Command{
		Use:     "flow FLOW-NAME",
		Short:   createFlowShortDesc,
		Long:    createFlowLongDesc,
		Example: createFlowExample,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = MuXRequiredFlags(cmd, []string{"description"}, false)
			if err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVar(&o.payload.Description, "description", "", flowDescriptionFlag)
	cmd.Flags().StringVar(&o.payload.Visibility, "visibility", "private", visibilityFlag)
	cmd.Flags().StringSliceVarP(&o.payload.Template, "template", "t", []string{"artifact"}, templateFlag)
	addDryRunFlag(cmd)

	return cmd
}

func (o *createFlowOptions) run(args []string) error {
	var err error
	url := fmt.Sprintf("%s/api/v2/flows/%s", global.Host, global.Org)

	if o.payload.Name == "" {
		if len(args) == 0 {
			return fmt.Errorf("flow name must be provided as an argument")
		}
		o.payload.Name = args[0]
	}
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
