package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const createPolicyShortDesc = `Create or update a Kosli policy.`

const createPolicyLongDesc = `Updating policy content creates a new version of the policy.`

const createPolicyExample = `
# create a Kosli policy:
kosli create policy yourPolicyName \
	--description yourPolicyDescription \
	--type environment \
	--api-token yourAPIToken \
	--org yourOrgName

# update a Kosli policy:
kosli create policy yourFlowName \
	--description yourPolicyDescription \
	--type environment \
	--comment yourChangeComment \
	--api-token yourAPIToken \
	--org yourOrgName
`

type createPolicyOptions struct {
	payload PolicyPayload
}

type PolicyPayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Comment     string `json:"comment"`
	Type        string `json:"type"`
}

func newCreatePolicyCmd(out io.Writer) *cobra.Command {
	o := new(createPolicyOptions)
	cmd := &cobra.Command{
		Use:     "policy POLICY-NAME POLICY-FILE-PATH",
		Short:   createPolicyShortDesc,
		Long:    createPolicyLongDesc,
		Example: createPolicyExample,
		Args:    cobra.ExactArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVar(&o.payload.Description, "description", "", policyDescriptionFlag)
	cmd.Flags().StringVar(&o.payload.Comment, "comment", "", policyCommentFlag)
	cmd.Flags().StringVar(&o.payload.Type, "type", "env", policyTypeFlag)

	addDryRunFlag(cmd)

	return cmd
}

func (o *createPolicyOptions) run(args []string) error {
	var reqParams *requests.RequestParams
	var url string
	o.payload.Name = args[0]
	policyFile := args[1]

	url = fmt.Sprintf("%s/api/v2/policies/%s", global.Host, global.Org)

	form, err := newPolicyForm(o.payload, policyFile)
	if err != nil {
		return err
	}

	reqParams = &requests.RequestParams{
		Method: http.MethodPut,
		URL:    url,
		Form:   form,
		DryRun: global.DryRun,
		Token:  global.ApiToken,
	}

	res, err := kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		verb := "created"
		if res.Resp.StatusCode == 200 {
			verb = "updated"
		}
		logger.Info("policy '%s' was %s", o.payload.Name, verb)
	}
	return err
}

// newPolicyForm constructs a list of FormItems for a policy with a policy file
// form submission.
func newPolicyForm(payload interface{}, policyFile string) ([]requests.FormItem, error) {
	if policyFile == "" {
		return []requests.FormItem{}, fmt.Errorf("cannot create a policy form without a policy file")
	}
	form := []requests.FormItem{
		{Type: "field", FieldName: "payload", Content: payload},
		{Type: "file", FieldName: "policy_file", Content: policyFile},
	}

	return form, nil
}
