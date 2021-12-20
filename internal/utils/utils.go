package utils

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"

	"github.com/docker/docker/api/types"
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
