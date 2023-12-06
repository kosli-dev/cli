package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const createFlowWithTemplateShortDesc = `Create or update a Kosli flow.`

const createFlowWithTemplateLongDesc = createFlowShortDesc + `
You can specify flow parameters in flags.`

const createFlowWithTemplateExample = `
# create/update a Kosli flow:
kosli create flow yourFlowName \
	--description yourFlowDescription \
    --visibility private OR public \
	--template-file /path/to/your/template/file.yml \
	--api-token yourAPIToken \
	--org yourOrgName
`

type createFlowWithTemplateOptions struct {
	payload      FlowWithTemplatePayload
	TemplateFile string
}

type FlowWithTemplatePayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
}

func newCreateFlowWithTemplateCmd(out io.Writer) *cobra.Command {
	o := new(createFlowWithTemplateOptions)
	cmd := &cobra.Command{
		Use:     "flow2 FLOW-NAME",
		Hidden:  true,
		Short:   createFlowWithTemplateShortDesc,
		Long:    createFlowWithTemplateLongDesc,
		Example: createFlowWithTemplateExample,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) == 0 {
				return fmt.Errorf("flow name must be provided as an argument")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVar(&o.payload.Description, "description", "", flowDescriptionFlag)
	cmd.Flags().StringVar(&o.payload.Visibility, "visibility", "private", visibilityFlag)
	cmd.Flags().StringVarP(&o.TemplateFile, "template-file", "f", "", templateFileFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"template-file"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *createFlowWithTemplateOptions) run(args []string) error {
	url := fmt.Sprintf("%s/api/v2/flows/%s/template_file", global.Host, global.Org)

	o.payload.Name = args[0]
	form, err := newFlowForm(o.payload, o.TemplateFile, false)
	if err != nil {
		return err
	}

	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Form:     form,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	res, err := kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		verb := "created"
		if res.Resp.StatusCode == 200 {
			verb = "updated"
		}
		logger.Info("flow '%s' was %s", o.payload.Name, verb)
	}
	return err
}

// newFlowForm constructs a list of FormItems for a flow with a template file
// form submission.
func newFlowForm(payload interface{}, templateFile string, templateRequired bool) ([]requests.FormItem, error) {
	if templateFile == "" && templateRequired {
		return []requests.FormItem{}, fmt.Errorf("cannot create a flow form without a template file")
	}
	form := []requests.FormItem{
		{Type: "field", FieldName: "data_json", Content: payload},
	}

	if templateFile != "" {
		form = append(form, requests.FormItem{Type: "file", FieldName: "template_file", Content: templateFile})
		logger.Debug("template file %s will be uploaded", templateFile)
	}

	return form, nil
}
