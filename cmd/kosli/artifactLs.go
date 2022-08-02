package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const artifactLsDesc = `List a number of artifacts in a pipeline.`

type artifactLsOptions struct {
	json       bool
	pageNumber int64
	pageLimit  int64
}

func newArtifactLsCmd(out io.Writer) *cobra.Command {
	o := new(artifactLsOptions)
	cmd := &cobra.Command{
		Use:     "ls PIPELINE-NAME",
		Aliases: []string{"list"},
		Short:   artifactLsDesc,
		Long:    artifactLsDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorBeforePrintingUsage(cmd, "pipeline name argument is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().BoolVarP(&o.json, "json", "j", false, environmentJsonFlag)
	cmd.Flags().Int64VarP(&o.pageNumber, "page-number", "n", 1, pageNumberFlag)
	cmd.Flags().Int64VarP(&o.pageLimit, "page-limit", "l", 15, pageLimitFlag)

	return cmd
}

func (o *artifactLsOptions) run(out io.Writer, args []string) error {
	if o.pageNumber <= 0 {
		_, err := out.Write([]byte("No artifacts were requested\n"))
		if err != nil {
			return err
		}
		return nil
	}
	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%d/%d",
		global.Host, global.Owner, args[0], o.pageNumber, o.pageLimit)
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())
	if err != nil {
		return err
	}

	if o.json {
		pj, err := prettyJson(response.Body)
		if err != nil {
			return err
		}
		fmt.Println(pj)
		return nil
	}

	var artifacts []map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &artifacts)
	if err != nil {
		return err
	}

	if len(artifacts) == 0 {
		_, err := out.Write([]byte("No artifacts were found\n"))
		if err != nil {
			return err
		}
		return nil
	}

	header := []string{"COMMIT", "ARTIFACT", "STATE", "CREATED_AT"}
	rows := []string{}
	for _, artifact := range artifacts {
		evidenceMap := artifact["evidence"].(map[string]interface{})
		artifactData := evidenceMap["artifact"].(map[string]interface{})

		gitCommit := artifactData["git_commit"].(string)[:7]
		artifactName := artifactData["filename"].(string)
		// if len(artifactName) > 50 {
		// 	artifactName = artifactName[:18] + "..." + artifactName[len(artifactName)-19:]
		// }
		artifactDigest := artifactData["sha256"].(string)
		artifactState := artifact["state"].(string)
		createdAt, err := formattedTimestamp(artifact["created_at"], true)
		if err != nil {
			return err
		}

		row := fmt.Sprintf("%s\tName: %s\t%s\t%s", gitCommit, artifactName, artifactState, createdAt)
		rows = append(rows, row)
		row = fmt.Sprintf("\tSHA256: %s\t\n", artifactDigest)
		rows = append(rows, row)
	}
	printTable(out, header, rows)

	return nil
}
