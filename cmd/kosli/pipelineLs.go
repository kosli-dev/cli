package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const pipelineLsDesc = `List pipelines for an org.`

type pipelineLsOptions struct {
	json bool
}

func newPipelineLsCmd(out io.Writer) *cobra.Command {
	o := new(pipelineLsOptions)
	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   pipelineLsDesc,
		Long:    pipelineLsDesc,
		Args:    NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorAfterPrintingHelp(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	cmd.Flags().BoolVarP(&o.json, "json", "j", false, pipelineJsonFlag)

	return cmd
}

func (o *pipelineLsOptions) run(out io.Writer) error {
	url := fmt.Sprintf("%s/api/v1/projects/%s/", global.Host, global.Owner)
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())

	if err != nil {
		return err
	}

	if o.json {
		pj, err := prettyJson(response.Body)
		if err != nil {
			return err
		}
		fmt.Println(pj)
		return nil
	}

	var pipelines []map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &pipelines)
	if err != nil {
		return err
	}

	header := []string{"NAME", "DESCRIPTION", "VISIBILITY"}
	rows := []string{}
	for _, pipeline := range pipelines {
		row := fmt.Sprintf("%s\t%s\t%s", pipeline["name"], pipeline["description"], pipeline["visibility"])
		rows = append(rows, row)
	}
	printTable(out, header, rows)

	return nil
}
