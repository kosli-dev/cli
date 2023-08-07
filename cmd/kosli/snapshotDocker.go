package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/kosli-dev/cli/internal/digest"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/server"
	"github.com/spf13/cobra"
)

const snapshotDockerShortDesc = `Report a snapshot of running containers from docker host to Kosli.  `

const snapshotDockerLongDesc = snapshotDockerShortDesc + `
The reported data includes container image digests 
and creation timestamps. Containers running images which have not
been pushed to or pulled from a registry will be ignored.`

const snapshotDockerExample = `
# report what is running in a docker host:
kosli snapshot docker yourEnvironmentName \
	--api-token yourAPIToken \
	--org yourOrgName`

type snapshotDockerOptions struct{}

func newSnapshotDockerCmd(out io.Writer) *cobra.Command {
	o := new(snapshotDockerOptions)
	cmd := &cobra.Command{
		Use:     "docker ENVIRONMENT-NAME",
		Short:   snapshotDockerShortDesc,
		Long:    snapshotDockerLongDesc,
		Example: snapshotDockerExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}
	addDryRunFlag(cmd)
	return cmd
}

func (o *snapshotDockerOptions) run(args []string) error {
	envName := args[0]

	url := fmt.Sprintf("%s/api/v2/environments/%s/%s/report/docker", global.Host, global.Org, envName)

	artifacts, err := CreateDockerArtifactsData()
	if err != nil {
		return err
	}

	payload := &server.ServerEnvRequest{
		Artifacts: artifacts,
	}

	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Payload:  payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("[%d] containers were reported to environment %s", len(payload.Artifacts), envName)
	}
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
		digests[c.Image], err = digest.DockerImageSha256(c.Image)
		if err != nil {
			if errors.Is(err, digest.ErrRepoDigestUnavailable) {
				containerName := strings.TrimPrefix(c.Names[0], "/")
				logger.Info("ignoring container '%s' as it uses an image with no repo digest", containerName)
				continue
			}
			return result, err
		}
		result = append(result, &server.ServerData{Digests: digests, CreationTimestamp: c.Created})
	}
	return result, nil
}
