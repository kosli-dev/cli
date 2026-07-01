package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const getControlShortDesc = `Get a Kosli control.`

const getControlExample = `
# get a control:
kosli get control yourControlIdentifier \
	--api-token yourAPIToken \
	--org yourOrgName
`

type getControlOptions struct {
	output string
}

func newGetControlCmd(out io.Writer) *cobra.Command {
	o := new(getControlOptions)
	cmd := &cobra.Command{
		Use:     "control CONTROL-IDENTIFIER",
		Short:   getControlShortDesc,
		Long:    getControlShortDesc,
		Example: getControlExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)

	return cmd
}

func (o *getControlOptions) run(out io.Writer, args []string) error {
	url, err := url.JoinPath(global.Host, "api/v2/controls", global.Org, args[0])
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
			"table": printControlAsTable,
			"json":  output.PrintJson,
		})
}

func printControlAsTable(raw string, out io.Writer, page int) error {
	var control map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &control); err != nil {
		return err
	}

	rows := []string{}
	rows = append(rows, fmt.Sprintf("Identifier:\t%s", control["identifier"]))
	rows = append(rows, fmt.Sprintf("Name:\t%s", control["name"]))
	if description, ok := control["description"]; ok && description != nil {
		rows = append(rows, fmt.Sprintf("Description:\t%s", description))
	}
	if version, ok := control["version"]; ok && version != nil {
		rows = append(rows, fmt.Sprintf("Version:\t%.0f", version))
	}
	if archived, ok := control["archived"]; ok && archived != nil {
		rows = append(rows, fmt.Sprintf("Archived:\t%t", archived))
	}
	if createdBy, ok := control["created_by"]; ok && createdBy != nil {
		rows = append(rows, fmt.Sprintf("Created by:\t%s", createdBy))
	}
	if createdAt, ok := control["created_at"]; ok && createdAt != nil {
		createdAtFormatted, err := formattedTimestamp(createdAt, false)
		if err != nil {
			return err
		}
		rows = append(rows, fmt.Sprintf("Created at:\t%s", createdAtFormatted))
	}

	if tags, ok := control["tags"].(map[string]interface{}); ok && len(tags) > 0 {
		tagKeys := make([]string, 0, len(tags))
		for key := range tags {
			tagKeys = append(tagKeys, key)
		}
		sort.Strings(tagKeys)
		tagPairs := make([]string, 0, len(tags))
		for _, key := range tagKeys {
			tagPairs = append(tagPairs, fmt.Sprintf("%s=%s", key, tags[key]))
		}
		rows = append(rows, fmt.Sprintf("Tags:\t%s", strings.Join(tagPairs, ", ")))
	}

	if links, ok := control["links"].(map[string]interface{}); ok && len(links) > 0 {
		rows = append(rows, "Links:\t")
		linkNames := make([]string, 0, len(links))
		for name := range links {
			linkNames = append(linkNames, name)
		}
		sort.Strings(linkNames)
		for _, name := range linkNames {
			rows = append(rows, fmt.Sprintf("\t%s:\t%s", name, links[name]))
		}
	}

	if policies, ok := control["policies_referencing"].([]interface{}); ok && len(policies) > 0 {
		policyNames := make([]string, 0, len(policies))
		for _, p := range policies {
			policyNames = append(policyNames, fmt.Sprintf("%s", p))
		}
		rows = append(rows, fmt.Sprintf("Policies referencing:\t%s", strings.Join(policyNames, ", ")))
	}

	tabFormattedPrint(out, []string{}, rows)
	return nil
}
