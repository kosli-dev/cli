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
	// assert bool
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

	// cmd.Flags().BoolVar(&o.assert, "assert", false, assertStatusFlag)

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

		for _, env := range envs {
			fmt.Println(env["name"])
		}
	}
	if outErr != nil {
		return outErr
	}
	return nil
}
