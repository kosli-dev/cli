package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const statusShortDesc = `Check the status of a Kosli server.  `

const statusLongDesc = statusShortDesc + `
The status is logged and the command always exits with 0 exit code.  
If you like to assert the Kosli server status, you can use the --assert flag or the "kosli assert status" command.`

type statusOptions struct {
	assert bool
}

func newStatusCmd(out io.Writer) *cobra.Command {
	o := new(statusOptions)
	cmd := &cobra.Command{
		Use:   "status",
		Short: statusShortDesc,
		Long:  statusLongDesc,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	cmd.Flags().BoolVar(&o.assert, "assert", false, assertStatusFlag)

	return cmd
}

func (o *statusOptions) run(out io.Writer) error {
	url := fmt.Sprintf("%s/ready", global.Host)
	reqParams := &requests.RequestParams{
		Method:   http.MethodGet,
		URL:      url,
		Password: global.ApiToken,
	}

	response, err := kosliClient.Do(reqParams)
	if err != nil {
		if o.assert {
			return fmt.Errorf("kosli server %s is unresponsive", global.Host)
		}
		logger.Info("Kosli is Down")
	} else {
		logger.Info(response.Body)
	}
	return nil
}
