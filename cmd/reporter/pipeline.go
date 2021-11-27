package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

const pipelineDesc = `
Create a Merkely pipeline by providing a JSON pipefile.
The pipefile contains the pipeline metadata and compliance template.
`

const pipelineExample = `
* create a Merkely pipeline with a pipefile:
merkely create pipeline --api-token 1234 /path/to/pipefile.json

* The pipefile format is:
{
    "owner": "organization-name",
    "name": "pipeline-name",
    "description": "pipeline short description",
    "visibility": "public or private",
    "template": [
        "artifact",
        "evidence-type1",
		"evidence-type2"
    ]
}
`

type Pipefile struct {
	Owner       string   `json:"owner"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Visibility  string   `json:"visibility"`
	Template    []string `json:"template"`
}

func newPipelineCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "pipeline",
		Short:             "Create a Merkely pipeline",
		Long:              pipelineDesc,
		Example:           pipelineExample,
		DisableAutoGenTag: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("only pipefile path argument is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return fmt.Errorf("pipefile path is required")
			}

			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			pipefilePath := args[0]
			pipe, err := getPipe(pipefilePath)
			if err != nil {
				return err
			}
			owner := pipe.Owner
			url := fmt.Sprintf("%s/api/v1/projects/%s/", global.Host, owner)

			_, err = requests.SendPayload(pipe, url, "", global.ApiToken,
				global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
			return err
		},
	}

	return cmd
}

// getPipe deserializes a JSON file into a Pipefile struct
func getPipe(pipefilePath string) (Pipefile, error) {
	var pipe Pipefile
	jsonFile, err := os.Open(pipefilePath)
	if err != nil {
		return pipe, fmt.Errorf("failed to open pipefile: %v", err)
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)

	err = json.Unmarshal(byteValue, &pipe)
	if err != nil {
		return pipe, fmt.Errorf("failed to unmarshal json: %v", err)
	}
	return pipe, nil
}
