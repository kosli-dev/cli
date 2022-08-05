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

const approvalGetDesc = `Get an approval from a specified pipeline`

type approvalGetOptions struct {
	json         bool
	pipelineName string
}

func newApprovalGetCmd(out io.Writer) *cobra.Command {
	o := new(approvalGetOptions)
	cmd := &cobra.Command{
		Use:   "get APPROVAL-ID",
		Short: approvalGetDesc,
		Long:  approvalGetDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorBeforePrintingUsage(cmd, "approval ID argument is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", pipelineNameFlag)
	cmd.Flags().BoolVarP(&o.json, "json", "j", false, jsonOutputFlag)

	err := RequireFlags(cmd, []string{"pipeline"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *approvalGetOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/approvals/%s", global.Host, global.Owner, o.pipelineName, args[0])
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

	var approval map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &approval)
	if err != nil {
		return err
	}

	rows := []string{}
	rows = append(rows, fmt.Sprintf("ID:\t%d", int64(approval["release_number"].(float64))))
	rows = append(rows, fmt.Sprintf("Artifact SHA256:\t%s", approval["base_artifact"].(string)))
	rows = append(rows, fmt.Sprintf("Artifact name:\t%s", approval["artifact_name"].(string)))
	rows = append(rows, fmt.Sprintf("State:\t%s", approval["state"].(string)))
	lastModifiedAt, err := formattedTimestamp(approval["last_modified_at"], false)
	if err != nil {
		return err
	}
	rows = append(rows, fmt.Sprintf("Last modified at:\t%s", lastModifiedAt))
	reviews := approval["approvals"].([]interface{})
	if len(reviews) > 0 {
		rows = append(rows, "Reviews:")
		for _, review := range reviews {
			convertedReview := review.(map[string]interface{})
			approvedBy := "Unknown"
			if convertedReview["approved_by"] != nil {
				approvedBy = convertedReview["approved_by"].(string)
			}
			reviewRow := fmt.Sprintf("\t%s By: %s", convertedReview["state"].(string), approvedBy)
			rows = append(rows, reviewRow)
		}
	} else {
		rows = append(rows, "Reviews:\tNone")
	}

	printTable(out, []string{}, rows)
	return nil
}
