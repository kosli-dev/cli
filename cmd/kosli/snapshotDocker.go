package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/kosli-dev/cli/internal/digest"
	log "github.com/kosli-dev/cli/internal/logger"
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

	url, err := url.JoinPath(global.Host, "api/v2/environments", global.Org, envName, "report/docker")
	if err != nil {
		return err
	}

	artifacts, err := CreateDockerArtifactsData()
	if err != nil {
		return err
	}

	payload := &server.ServerEnvRequest{
		Artifacts: artifacts,
	}

	reqParams := &requests.RequestParams{
		Method:  http.MethodPut,
		URL:     url,
		Payload: payload,
		DryRun:  global.DryRun,
		Token:   global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("[%d] containers were reported to environment %s", len(payload.Artifacts), envName)
	}
	return err
}

func CreateDockerArtifactsData() ([]*server.ServerData, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return []*server.ServerData{}, err
	}

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return []*server.ServerData{}, err
	}

	return dockerArtifactsFromContainers(containers, digest.DockerImageSha256, logger)
}

func dockerArtifactsFromContainers(
	containers []container.Summary,
	getDigest func(imageID string) (string, error),
	log *log.Logger,
) ([]*server.ServerData, error) {
	result := []*server.ServerData{}
	for _, c := range containers {
		d, err := getDigest(c.Image)
		if err != nil {
			containerName := containerName(c)
			switch {
			case errors.Is(err, digest.ErrRepoDigestUnavailable):
				log.Info("ignoring container '%s' as it uses an image with no repo digest", containerName)
				continue
			case cerrdefs.IsNotFound(err):
				log.Warn("ignoring container '%s' as its image is no longer present locally: %v", containerName, err)
				continue
			default:
				return []*server.ServerData{}, err
			}
		}
		result = append(result, &server.ServerData{
			Digests:           map[string]string{c.Image: d},
			CreationTimestamp: c.Created,
		})
	}
	return result, nil
}

func containerName(c container.Summary) string {
	if len(c.Names) == 0 {
		return c.Image
	}
	return strings.TrimPrefix(c.Names[0], "/")
}
