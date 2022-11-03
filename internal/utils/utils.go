package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// PullDockerImage pulls a docker image or returns an error
func PullDockerImage(imageName string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	rc, err := cli.ImagePull(context.Background(), imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer rc.Close()
	_, err = io.Copy(os.Stdout, rc)
	if err != nil {
		return err
	}

	return nil
}

// PushDockerImage pushes a docker image or returns an error
func PushDockerImage(imageName string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	authConfig := types.AuthConfig{
		ServerAddress: "http://localhost:5001/",
	}
	authConfigBytes, _ := json.Marshal(authConfig)
	authConfigEncoded := base64.URLEncoding.EncodeToString(authConfigBytes)
	opts := types.ImagePushOptions{RegistryAuth: authConfigEncoded}

	rc, err := cli.ImagePush(context.Background(), imageName, opts)
	if err != nil {
		return err
	}
	defer rc.Close()
	_, err = io.Copy(os.Stdout, rc)
	if err != nil {
		return err
	}

	return nil
}

// TagDockerImage tags a docker image or returns an error
func TagDockerImage(sourceName, targetName string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	return cli.ImageTag(context.Background(), sourceName, targetName)
}

// RemoveDockerImage deletes a docker image or return an error
func RemoveDockerImage(imageName string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	_, err = cli.ImageRemove(context.Background(), imageName, types.ImageRemoveOptions{Force: true})
	if err != nil {
		return err
	}

	return nil
}

// RunDockerContainer runs a docker container and returns its ID or returns an error
func RunDockerContainer(imageName string) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Cmd:   []string{"sleep", "360"},
	}, nil, nil, nil, "")

	if err != nil {
		return "", err
	}
	containerID := resp.ID
	return containerID, cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
}

// RemoveDockerContainer remove a docker container or returns an error
func RemoveDockerContainer(containerID string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	return cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{Force: true})
}

// Contains checks if a string is contained in a string slice
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// LoadFileContent loads file content
func LoadFileContent(filepath string) (string, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(data), err
}

// IsJSON checks if a string is in JSON format
func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// Creates a file under specified path
func CreateFile(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return nil, err
	}
	return os.Create(p)
}
