package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
		Args:  NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	cmd.Flags().BoolVarP(&o.long, "long", "l", false, environmentLongFlag)

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
			fmt.Printf("%-15s %s\n", "NAME", "TYPE")
		}
		for _, env := range envs {
			if o.long {
				fmt.Printf("%-15s %s\n", env["name"], env["type"])
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
