package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const beginTrailShortDesc = `Begin or update a Kosli flow trail.`

const beginTrailExample = `
# begin/update a Kosli flow trail:
kosli begin trail yourTrailName \
	--description yourTrailDescription \
	--template-file /path/to/your/template/file.yml \
	--user-data /path/to/your/user-data/file.json \
	--api-token yourAPIToken \
	--org yourOrgName
`

type beginTrailOptions struct {
	payload      TrailPayload
	templateFile string
	userDataFile string
	flow         string
}

type TrailPayload struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	UserData    interface{} `json:"user_data"`
}

func newBeginTrailCmd(out io.Writer) *cobra.Command {
	o := new(beginTrailOptions)
	cmd := &cobra.Command{
		Use:     "trail TRAIL-NAME",
		Hidden:  true,
		Short:   beginTrailShortDesc,
		Long:    beginTrailShortDesc,
		Example: beginTrailExample,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) == 0 {
				return fmt.Errorf("trail name must be provided as an argument")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVar(&o.flow, "flow", "", flowNameFlag)
	cmd.Flags().StringVar(&o.payload.Description, "description", "", trailDescriptionFlag)
	cmd.Flags().StringVarP(&o.templateFile, "template-file", "f", "", templateFileFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", trailUserDataFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"flow"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *beginTrailOptions) run(args []string) error {
	url := fmt.Sprintf("%s/api/v2/trails/%s/%s", global.Host, global.Org, o.flow)

	o.payload.Name = args[0]

	var err error
	o.payload.UserData, err = LoadJsonData(o.userDataFile)
	if err != nil {
		return err
	}

	form, err := newFlowForm(o.payload, o.templateFile, false)
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
		verb := "begun"
		if res.Resp.StatusCode == 200 {
			verb = "updated"
		}
		logger.Info("trail '%s' was %s", o.payload.Name, verb)
	}
	return err
}
