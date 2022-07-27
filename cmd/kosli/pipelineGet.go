package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/xeonx/timeago"
)

const pipelineGetDesc = `Get the metadata of a single pipeline`

type pipelineGetOptions struct {
	json bool
}

func newPipelineGetCmd(out io.Writer) *cobra.Command {
	o := new(pipelineGetOptions)
	cmd := &cobra.Command{
		Use:   "get PIPELINE-NAME",
		Short: pipelineGetDesc,
		Long:  pipelineGetDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorAfterPrintingHelp(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorAfterPrintingHelp(cmd, "pipeline name argument is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().BoolVarP(&o.json, "json", "j", false, jsonOutputFlag)

	return cmd
}

func (o *pipelineGetOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v1/projects/%s/%s", global.Host, global.Owner, args[0])
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

	var pipeline map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &pipeline)
	if err != nil {
		return err
	}

	fmt.Printf("Name: %s\n", pipeline["name"])
	fmt.Printf("Description: %s\n", pipeline["description"])
	fmt.Printf("Visibility: %s\n", pipeline["visibility"])
	template := fmt.Sprintf("%s", pipeline["template"])
	template = strings.Replace(template, " ", ", ", -1)
	fmt.Printf("Template: %s\n", template)
	timeago.English.Max = 36 * timeago.Month
	timestampFloat, err := strconv.ParseFloat(pipeline["last_deployment_at"].(string), 64)
	if err != nil {
		return err
	}
	timestamp := time.Unix(int64(timestampFloat), 0)
	last_deployment_at := timeago.English.Format(timestamp)
	fmt.Printf("Last deployment at: %s \u2022 %s\n", timestamp.Format(time.RFC822), last_deployment_at)

	return nil
}
