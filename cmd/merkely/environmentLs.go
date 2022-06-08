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

const environmentLsDesc = `
List environments.
`

type environmentLsOptions struct {
	long bool
}

func newEnvironmentLsCmd(out io.Writer) *cobra.Command {
	o := new(environmentLsOptions)
	cmd := &cobra.Command{
		Use:   "ls",
		Short: environmentLsDesc,
		Long:  environmentLsDesc,
		// Args:  NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	cmd.Flags().BoolVarP(&o.long, "long", "l", false, environmentLongFlag)

	// Add subcommands
	cmd.AddCommand(
		newSnapshotLsCmd(out),
	)

	return cmd
}

func (o *environmentLsOptions) run(out io.Writer) error {
	url := fmt.Sprintf("%s/api/v1/environments/%s/", global.Host, global.Owner)
	var outErr error
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())

	if err != nil {
		// if o.assert {
		// 	return fmt.Errorf("merkely server %s is unresponsive", global.Host)
		// }
		_, outErr = out.Write([]byte(err.Error()))
	} else {
		var envs []map[string]interface{}
		err = json.Unmarshal([]byte(response.Body), &envs)
		if err != nil {
			return err
		}

		if o.long {
			fmt.Printf("%-15s %-10s %-27s %s\n", "NAME", "TYPE", "LAST REPORT", "LAST MODIFIED")
		}
		for _, env := range envs {
			if o.long {
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
				fmt.Printf("%-15s %-10s %-27s %s\n", env["name"], env["type"], last_reported_str, last_modified_str)
			} else {
				fmt.Println(env["name"])
			}
		}
	}
	if outErr != nil {
		return outErr
	}
	return nil
}
