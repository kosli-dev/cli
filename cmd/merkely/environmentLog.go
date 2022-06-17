package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/merkely-development/reporter/internal/requests"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const environmentLogDesc = `Show log of snapshots.`

type environmentLogOptions struct {
	// long bool
	json bool
}

// type EnvironmentDiffPayload struct {
// 	Snappish1 string `json:"snappish1"`
// 	Snappish2 string `json:"snappish2"`
// }

func newEnvironmentLogCmd(out io.Writer) *cobra.Command {
	o := new(environmentLogOptions)
	cmd := &cobra.Command{
		Use:   "log [ENVIRONMENT-NAME]",
		Short: environmentLogDesc,
		Long:  environmentLogDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorAfterPrintingHelp(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	// cmd.Flags().BoolVarP(&o.long, "long", "l", false, environmentLongFlag)
	cmd.Flags().BoolVarP(&o.json, "json", "j", false, environmentJsonFlag)

	return cmd
}

func (o *environmentLogOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v1/environments/%s/%s/log/0/5", global.Host, global.Owner, args[0])
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())

	if err != nil {
		return fmt.Errorf("kosli server %s is unresponsive %s", global.Host, err)
	}

	if o.json {
		pj, err := prettyJson(response.Body)
		if err != nil {
			return err
		}
		fmt.Println(pj)
	} else {
		var snapshots []map[string]interface{}
		err = json.Unmarshal([]byte(response.Body), &snapshots)
		if err != nil {
			return err
		}

		fmt.Printf("SNAPSHOT  FROM\n")
		for _, snapshot := range snapshots {
			tsFrom := time.Unix(int64(snapshot["from"].(float64)), 0).Format(time.RFC3339)
			index := int64(snapshot["index"].(float64))
			fmt.Printf("%-8d  %s\n", index, tsFrom)
		}

	}

	return nil
}
