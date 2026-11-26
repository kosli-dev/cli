package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const createControlShortDesc = `Create a Kosli control.`

const createControlLongDesc = createControlShortDesc + `

^CONTROL-IDENTIFIER^ must start with a letter or number, and only contain letters, numbers, ^.^, ^-^, ^_^, and ^~^.
`

const createControlExample = `
# create a Kosli control:
kosli create control yourControlIdentifier \
	--name "Your control name" \
	--description "what this control checks" \
	--api-token yourAPIToken \
	--org yourOrgName
`

type createControlOptions struct {
	payload ControlPayload
}

type ControlPayload struct {
	Identifier  string `json:"identifier"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func newCreateControlCmd(out io.Writer) *cobra.Command {
	o := new(createControlOptions)
	cmd := &cobra.Command{
		Use:     "control CONTROL-IDENTIFIER",
		Short:   createControlShortDesc,
		Long:    createControlLongDesc,
		Example: createControlExample,
		Args:    cobra.ExactArgs(1),
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

	cmd.Flags().StringVarP(&o.payload.Name, "name", "n", "", controlNameFlag)
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", controlDescriptionFlag)

	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"name"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *createControlOptions) run(args []string) error {
	o.payload.Identifier = args[0]
	url, err := url.JoinPath(global.Host, "api/v2/controls", global.Org)
	if err != nil {
		return err
	}

	reqParams := &requests.RequestParams{
		Method:  http.MethodPost,
		URL:     url,
		Payload: o.payload,
		DryRun:  global.DryRun,
		Token:   global.ApiToken,
		// A 409 here means the identifier already exists — a permanent error,
		// so surface it immediately rather than retrying.
		DisableConflictRetry: true,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("control %s was created", o.payload.Identifier)
	}
	return err
}
