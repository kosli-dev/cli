package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const environmentLsDesc = `List environments.`

type environmentLsOptions struct {
	long bool
	json bool
}

func newEnvironmentLsCmd(out io.Writer) *cobra.Command {
	o := new(environmentLsOptions)
	cmd := &cobra.Command{
		Use:     "ls [ENVIRONMENT-NAME]",
		Aliases: []string{"list"},
		Short:   environmentLsDesc,
		Long:    environmentLsDesc,
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

	cmd.Flags().BoolVarP(&o.long, "long", "l", false, environmentLongFlag)
	cmd.Flags().BoolVarP(&o.json, "json", "j", false, environmentJsonFlag)

	return cmd
}

func (o *environmentLsOptions) run(out io.Writer, args []string) error {
	if len(args) > 0 {
		return snapshotLs(out, o, args)
	}

	url := fmt.Sprintf("%s/api/v1/environments/%s/", global.Host, global.Owner)
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

	var envs []map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &envs)
	if err != nil {
		return err
	}

	if len(envs) == 0 {
		_, err := out.Write([]byte("No environments found\n"))
		if err != nil {
			return err
		}
		return nil
	}

	if o.long {
		header := []string{"NAME", "TYPE", "LAST REPORT", "LAST MODIFIED"}
		rows := []string{}
		for _, env := range envs {
			last_reported_str := ""
			last_reported_at := env["last_reported_at"]
			if last_reported_at != nil {
				last_reported_str = time.Unix(int64(last_reported_at.(float64)), 0).Format(time.RFC3339)
			}
			last_modified_str := ""
			last_modified_at := env["last_modified_at"]
			if last_modified_at != nil {
				last_modified_str = time.Unix(int64(last_modified_at.(float64)), 0).Format(time.RFC3339)
			}
			row := fmt.Sprintf("%s\t%s\t%s\t%s", env["name"], env["type"], last_reported_str, last_modified_str)
			rows = append(rows, row)
		}
		printTable(out, header, rows)
	} else {
		for _, env := range envs {
			_, err := out.Write([]byte(env["name"].(string) + "\n"))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
