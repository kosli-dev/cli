package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/registry"
	"github.com/moby/moby/client"
)

func newDockerClient() (*client.Client, error) {
	return client.New(client.FromEnv)
}

// PullDockerImage pulls a docker image or returns an error
func PullDockerImage(imageName string) error {
	cli, err := newDockerClient()
	if err != nil {
		return err
	}

	rc, err := cli.ImagePull(context.Background(), imageName, client.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err := rc.Close(); err != nil {
			fmt.Printf("warning: failed to close image pull reader: %v\n", err)
		}
	}()
	_, err = io.Copy(os.Stdout, rc)
	if err != nil {
		return err
	}

	return nil
}

// PushDockerImage pushes a docker image to the local registry or returns an error
func PushDockerImage(imageName string) error {
	cli, err := newDockerClient()
	if err != nil {
		return err
	}

	authConfig := registry.AuthConfig{
		ServerAddress: "http://localhost:5001/",
	}
	authConfigBytes, _ := json.Marshal(authConfig)
	authConfigEncoded := base64.URLEncoding.EncodeToString(authConfigBytes)
	opts := client.ImagePushOptions{RegistryAuth: authConfigEncoded}

	rc, err := cli.ImagePush(context.Background(), imageName, opts)
	if err != nil {
		return err
	}
	defer func() {
		if err := rc.Close(); err != nil {
			fmt.Printf("warning: failed to close image push reader: %v\n", err)
		}
	}()
	_, err = io.Copy(os.Stdout, rc)
	if err != nil {
		return err
	}

	return nil
}

// TagDockerImage tags a docker image or returns an error
func TagDockerImage(sourceName, targetName string) error {
	cli, err := newDockerClient()
	if err != nil {
		return err
	}

	_, err = cli.ImageTag(context.Background(), client.ImageTagOptions{Source: sourceName, Target: targetName})
	return err
}

// RemoveDockerImage deletes a docker image or return an error
func RemoveDockerImage(imageName string) error {
	cli, err := newDockerClient()
	if err != nil {
		return err
	}

	_, err = cli.ImageRemove(context.Background(), imageName, client.ImageRemoveOptions{Force: true})
	if err != nil {
		return err
	}

	return nil
}

// RunDockerContainer runs a docker container that sleeps for 6 minutes and returns its ID or returns an error
func RunDockerContainer(imageName string) (string, error) {
	cli, err := newDockerClient()
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	resp, err := cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Image: imageName,
			Cmd:   []string{"sleep", "360"},
		},
	})

	if err != nil {
		return "", err
	}
	containerID := resp.ID
	_, err = cli.ContainerStart(ctx, containerID, client.ContainerStartOptions{})
	return containerID, err
}

// RemoveDockerContainer remove a docker container or returns an error
func RemoveDockerContainer(containerID string) error {
	cli, err := newDockerClient()
	if err != nil {
		return err
	}

	_, err = cli.ContainerRemove(context.Background(), containerID, client.ContainerRemoveOptions{Force: true})
	return err
}
