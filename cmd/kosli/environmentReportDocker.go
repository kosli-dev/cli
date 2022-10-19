package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/server"
	"github.com/spf13/cobra"
)

const environmentReportDockerDesc = `
List the artifacts running as containers and their digests 
and report them to Kosli. 
`

const environmentReportDockerExample = `
# report what is running in a docker host:
kosli environment report docker yourEnvironmentName \
	--api-token yourAPIToken \
	--owner yourOrgName
`

type environmentReportDockerOptions struct {
}

func newEnvironmentReportDockerCmd(out io.Writer) *cobra.Command {
	o := new(environmentReportDockerOptions)
	cmd := &cobra.Command{
		Use:     "docker ENVIRONMENT-NAME",
		Short:   "Report running containers data from docker host to Kosli.",
		Long:    environmentReportDockerDesc,
		Example: environmentReportDockerExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return ErrorBeforePrintingUsage(cmd, "only env-name argument is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return ErrorBeforePrintingUsage(cmd, "env-name argument is required")
			}

			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}
	return cmd
}

func (o *environmentReportDockerOptions) run(args []string) error {
	envName := args[0]

	url := fmt.Sprintf("%s/api/v1/environments/%s/%s/data", global.Host, global.Owner, envName)

	artifacts, err := CreateDockerArtifactsData()
	if err != nil {
		return err
	}

	requestBody := &server.ServerEnvRequest{
		Artifacts: artifacts,
		Type:      "docker",
		Id:        envName,
	}

	_, err = requests.SendPayload(requestBody, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
	return err
}

func CreateDockerArtifactsData() ([]*server.ServerData, error) {
	result := []*server.ServerData{}
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return result, err
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return result, err
	}

	for _, c := range containers {
		digests := make(map[string]string)
		digests[c.Image] = strings.TrimPrefix(c.ImageID, "sha256:")
		result = append(result, &server.ServerData{Digests: digests, CreationTimestamp: c.Created})
	}
	return result, nil
}
