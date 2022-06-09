package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/merkely-development/reporter/internal/requests"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const statusDesc = `
Check the status of Kosli server.
`

type statusOptions struct {
	assert bool
}

func newStatusCmd(out io.Writer) *cobra.Command {
	o := new(statusOptions)
	cmd := &cobra.Command{
		Use:   "status",
		Short: statusDesc,
		Long:  statusDesc,
		Args:  NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	cmd.Flags().BoolVar(&o.assert, "assert", false, assertStatusFlag)

	return cmd
}

func (o *statusOptions) run(out io.Writer) error {
	url := fmt.Sprintf("%s/ready", global.Host)
	var outErr error
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", "", global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())
	if err != nil {
		if o.assert {
			return fmt.Errorf("merkely server %s is unresponsive", global.Host)
		}
		_, outErr = out.Write([]byte("Down\n"))
	} else {
		_, outErr = out.Write([]byte(response.Body + "\n"))
	}
	if outErr != nil {
		return outErr
	}
	return nil
}
