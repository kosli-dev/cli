package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/utils"
	"github.com/spf13/cobra"
)

const createFlowShortDesc = `Create or update a Kosli flow.`

const createFlowLongDesc = createFlowShortDesc + `
You can specify flow parameters in flags.`

const createFlowExample = `
# create/update a Kosli flow (with empty template):
kosli create flow yourFlowName \
	--description yourFlowDescription \
	--visibility private OR public \
	--use-empty-template \
	--api-token yourAPIToken \
	--org yourOrgName

# create/update a Kosli flow (with template file):
kosli create flow yourFlowName \
	--description yourFlowDescription \
	--visibility private OR public \
	--template-file /path/to/your/template/file.yml \
	--api-token yourAPIToken \
	--org yourOrgName
`

type createFlowOptions struct {
	payload          FlowPayload
	TemplateFile     string
	UseEmptyTemplate bool
}

type FlowPayload struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Visibility  string   `json:"visibility"`
	Template    []string `json:"template,omitempty"`
}

func newCreateFlowCmd(out io.Writer) *cobra.Command {
	o := new(createFlowOptions)
	cmd := &cobra.Command{
		Use:     "flow FLOW-NAME",
		Short:   createFlowShortDesc,
		Long:    createFlowLongDesc,
		Example: createFlowExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = MuXRequiredFlags(cmd, []string{"template", "template-file"}, false)
			if err != nil {
				return err
			}

			err = MuXRequiredFlags(cmd, []string{"template-file", "use-empty-template"}, false)
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
	cmd.Flags().StringSliceVarP(&o.payload.Template, "template", "t", []string{}, templateFlag)
	cmd.Flags().MarkDeprecated("template", "use --template-file instead")
	cmd.Flags().StringVarP(&o.TemplateFile, "template-file", "f", "", templateFileFlag)
	cmd.Flags().BoolVar(&o.UseEmptyTemplate, "use-empty-template", false, useEmptyTemplateFlag)
	addDryRunFlag(cmd)

	return cmd
}

func (o *createFlowOptions) run(args []string) error {
	var reqParams *requests.RequestParams
	var url string
	o.payload.Name = args[0]

	if o.TemplateFile != "" || o.UseEmptyTemplate {
		url = fmt.Sprintf("%s/api/v2/flows/%s/template_file", global.Host, global.Org)
		if o.TemplateFile == "" {
			tmpDir, err := os.MkdirTemp("", "default-template")
			if err != nil {
				return fmt.Errorf("failed to create tmp directory for default template: %v", err)
			}
			defaultTemplatePath := filepath.Join(tmpDir, "template.yml")
			err = utils.CreateFileWithContent(defaultTemplatePath, "version: 1")
			if err != nil {
				return fmt.Errorf("failed to create default template: %v", err)
			}
			o.TemplateFile = defaultTemplatePath
			defer os.RemoveAll(tmpDir)
		}

		form, err := newFlowForm(o.payload, o.TemplateFile, false)
		if err != nil {
			return err
		}

		reqParams = &requests.RequestParams{
			Method:   http.MethodPut,
			URL:      url,
			Form:     form,
			DryRun:   global.DryRun,
			Password: global.ApiToken,
		}
	} else {
		// legacy flows
		url = fmt.Sprintf("%s/api/v2/flows/%s", global.Host, global.Org)
		o.payload.Template = injectArtifactIntoTemplateIfNotExisting(o.payload.Template)

		reqParams = &requests.RequestParams{
			Method:   http.MethodPut,
			URL:      url,
			Payload:  o.payload,
			DryRun:   global.DryRun,
			Password: global.ApiToken,
		}
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
