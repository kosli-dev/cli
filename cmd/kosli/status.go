package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const statusShortDesc = `Check the status of a Kosli server.  `

const statusLongDesc = statusShortDesc + `
The status is logged and the command always exits with 0 exit code.  
If you like to assert the Kosli server status, you can use the ^--assert^ flag or the "kosli assert status" command.`

type statusOptions struct {
	assert bool
}

func newStatusCmd(out io.Writer) *cobra.Command {
	o := new(statusOptions)
	cmd := &cobra.Command{
		Use:     "status",
		Aliases: []string{"st"},
		Short:   statusShortDesc,
		Long:    statusLongDesc,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	cmd.Flags().BoolVar(&o.assert, "assert", false, assertStatusFlag)

	return cmd
}

func (o *statusOptions) run(out io.Writer) error {
	h := kosliAPI().Query()
	url, err := url.JoinPath(h.URL, "ready")
	if err != nil {
		return err
	}
	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    url,
		Token:  h.Token,
	}

	response, err := h.Do(reqParams)
	if err != nil {
		logger.Debug("failed to check Kosli's readiness: %s", err.Error())
		if o.assert {
			return fmt.Errorf("kosli server %s is unresponsive", h.URL)
		}
		logger.Info("Kosli is Down")
	} else {
		logger.Info("%s", response.Body)
	}
	return nil
}
