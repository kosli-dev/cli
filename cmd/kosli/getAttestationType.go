package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const getAttestationTypeDesc = `Get a Kosli attestation type.`

type getAttestationTypeOptions struct {
	output string
}

func newGetAttestationTypeCmd(out io.Writer) *cobra.Command {
	o := new(getAttestationTypeOptions)
	cmd := &cobra.Command{
		Use:   "attestation-type TYPE-NAME",
		Short: getAttestationTypeDesc,
		Long:  getAttestationTypeDesc,
		Args:  cobra.ExactArgs(1),
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

func (o *getAttestationTypeOptions) run(out io.Writer, args []string) error {
	attestationType := args[0]
	var version string
	var er error
	if strings.Contains(attestationType, "@") {
		attestationType, version, er = handleCustomAttestationTypeExpression(attestationType)
		if er != nil {
			return er
		}
	}

	url := fmt.Sprintf("%s/api/v2/custom-attestation-types/%s/%s", global.Host, global.Org, attestationType)
	if version != "" {
		url = fmt.Sprintf("%s?version=%s", url, version)
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
			"table": printAttestationTypeAsTable,
			"json":  output.PrintJson,
		})
}

func printAttestationTypeAsTable(raw string, out io.Writer, page int) error {
	var attestationType map[string]interface{}
	err := json.Unmarshal([]byte(raw), &attestationType)
	if err != nil {
		return err
	}

	header := []string{}
	rows := []string{}

	rows = append(rows, fmt.Sprintf("Name:\t%s", attestationType["name"]))
	if description, ok := attestationType["description"]; ok {
		rows = append(rows, fmt.Sprintf("Description:\t%s", description))
	}
	rows = append(rows, fmt.Sprintf("Organization:\t%s", attestationType["org"]))
	rows = append(rows, fmt.Sprintf("Created By:\t%s", attestationType["created_by"]))

	if createdAt, ok := attestationType["created_at"]; ok {
		createdAtFormatted, err := formattedTimestamp(createdAt, false)
		if err != nil {
			return err
		}
		rows = append(rows, fmt.Sprintf("Created at:\t%s", createdAtFormatted))
	}

	if lastModifiedAt, ok := attestationType["last_modified_at"]; ok {
		lastModifiedAtFormatted, err := formattedTimestamp(lastModifiedAt, false)
		if err != nil {
			return err
		}
		rows = append(rows, fmt.Sprintf("Last modified at:\t%s", lastModifiedAtFormatted))
	}

	if versions, ok := attestationType["versions"]; ok {
		rows = append(rows, "Versions:\t")
		for _, version := range versions.([]interface{}) {
			versionMap := version.(map[string]interface{})
			rows, err = printVersionedAttestationTypeAsTable(versionMap, rows)
			if err != nil {
				return err
			}
			rows = append(rows, "	")
		}
	}

	tabFormattedPrint(out, header, rows)
	return nil
}

func printVersionedAttestationTypeAsTable(raw map[string]interface{}, rows []string) ([]string, error) {
	attestationType := raw

	timestamp, err := formattedTimestamp(attestationType["timestamp"], false)
	if err != nil {
		return []string{}, err
	}

	version := int(attestationType["version"].(float64))

	rows = append(rows, fmt.Sprintf("	Version:\t%d", version))
	if comment, ok := attestationType["comment"]; ok {
		rows = append(rows, fmt.Sprintf("	Comment:\t%s", comment))
	}
	rows = append(rows, fmt.Sprintf("	Timestamp:\t%s", timestamp))
	rows = append(rows, fmt.Sprintf("	Created By:\t%s", attestationType["created_by"]))

	if typeSchema, ok := attestationType["type_schema"]; ok {
		rows = append(rows, fmt.Sprintf("	Type schema:\t%s", typeSchema))
	}

	if evaluator, ok := attestationType["evaluator"].(map[string]interface{}); ok {
		rows = append(rows, "	Evaluator:\t")
		rows = append(rows, fmt.Sprintf("		Content Type:\t%s", evaluator["content_type"]))
		if rules, ok := evaluator["rules"].([]interface{}); ok {
			rows = append(rows, "		Rules:")
			for _, rule := range rules {
				rows = append(rows, fmt.Sprintf("	\t\t%s", rule))
			}
		}
	}

	return rows, nil
}
