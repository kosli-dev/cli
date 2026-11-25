package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/kosli"
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
	// Every host is checked, so --assert fails when any one is down.
	return kosliAPI().Send(func(h kosli.Host) error {
		url, err := url.JoinPath(h.URL, "ready")
		if err != nil {
			return err
		}
		response, err := h.Do(&requests.RequestParams{
			Method: http.MethodGet,
			URL:    url,
			Token:  h.Token,
		})
		if err != nil {
			h.Debug("failed to check Kosli's readiness: %s", err.Error())
			if o.assert {
				return fmt.Errorf("kosli server %s is unresponsive", h.URL)
			}
			h.Info("Kosli is Down")
			return nil
		}
		h.Info("%s", response.Body)
		return nil
	})
}
