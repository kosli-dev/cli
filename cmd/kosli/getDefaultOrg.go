package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const getDefaultOrgShortDesc = `Get the default organization for the current user.`

const getDefaultOrgLongDesc = getDefaultOrgShortDesc + `
The default organization is the one selected by default in the Kosli Web UI when you log in.`

const getDefaultOrgExample = `
# get the default organization for the current user:
kosli get default-org \
	--api-token yourAPIToken
`

type getDefaultOrgOptions struct {
	output string
}

// defaultOrg models the response of GET api/v2/user/default-org.
type defaultOrg struct {
	DefaultOrgName string `json:"default_org_name"`
}

func newGetDefaultOrgCmd(out io.Writer) *cobra.Command {
	o := new(getDefaultOrgOptions)
	cmd := &cobra.Command{
		Use:     "default-org",
		Short:   getDefaultOrgShortDesc,
		Long:    getDefaultOrgLongDesc,
		Example: getDefaultOrgExample,
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := RequireGlobalFlags(global, []string{"ApiToken"}); err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)

	return cmd
}

func (o *getDefaultOrgOptions) run(out io.Writer) error {
	url, err := url.JoinPath(global.Host, "api/v2/user/default-org")
	if err != nil {
		return err
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    url,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printDefaultOrgAsTable,
			"json":  output.PrintJson,
		})
}

// printDefaultOrgAsTable renders the default organization name.
func printDefaultOrgAsTable(raw string, out io.Writer, page int) error {
	var org defaultOrg
	if err := json.Unmarshal([]byte(raw), &org); err != nil {
		return err
	}

	name := org.DefaultOrgName
	if name == "" {
		name = "(none set)"
	}
	rows := []string{"Default user organization: " + name}
	tabFormattedPrint(out, []string{}, rows)

	return nil
}
