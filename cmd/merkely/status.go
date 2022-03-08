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
Check the status of Merkely server.
`

func newStatusCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: statusDesc,
		Long:  statusDesc,
		Args:  NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(out)
		},
	}
	return cmd
}

func run(out io.Writer) error {
	url := fmt.Sprintf("%s/ready", global.Host)
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", "", global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())
	if err != nil {
		log.Info("Down")
	} else {
		log.Info(response.Body)
	}
	return nil
}
