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

const approvalLsDesc = `List a number of approvals in a pipeline.`

type approvalLsOptions struct {
	json       bool
	pageNumber int64
	pageLimit  int64
}

func newApprovalLsCmd(out io.Writer) *cobra.Command {
	o := new(approvalLsOptions)
	cmd := &cobra.Command{
		Use:     "ls PIPELINE-NAME",
		Aliases: []string{"list"},
		Short:   approvalLsDesc,
		Long:    approvalLsDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorBeforePrintingUsage(cmd, "pipeline name argument is required")
			}
			if o.pageNumber <= 0 {
				return ErrorBeforePrintingUsage(cmd, "page number must be a positive integer")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().BoolVarP(&o.json, "json", "j", false, jsonOutputFlag)
	cmd.Flags().Int64VarP(&o.pageNumber, "page-number", "n", 1, pageNumberFlag)
	cmd.Flags().Int64VarP(&o.pageLimit, "page-limit", "l", 15, pageLimitFlag)

	return cmd
}

func (o *approvalLsOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/approvals/%d/%d",
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

	var approvals []map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &approvals)
	if err != nil {
		return err
	}

	if len(approvals) == 0 {
		msg := "No approvals were found"
		if o.pageNumber != 1 {
			msg = fmt.Sprintf("%s at page number %d", msg, o.pageNumber)
		}
		_, err := out.Write([]byte(msg + ".\n"))
		if err != nil {
			return err
		}
		return nil
	}

	header := []string{"ID", "ARTIFACT", "STATE", "LAST_MODIFIED_AT"}
	rows := []string{}
	for _, approval := range approvals {
		approvalId := int(approval["release_number"].(float64))
		artifactName := approval["artifact_name"].(string)
		approvalState := approval["state"].(string)
		artifactDigest := approval["base_artifact"].(string)
		lastModifiedAt, err := formattedTimestamp(approval["last_modified_at"], true)
		if err != nil {
			return err
		}
		row := fmt.Sprintf("%d\tName: %s\t%s\t%s", approvalId, artifactName, approvalState, lastModifiedAt)
		rows = append(rows, row)
		row = fmt.Sprintf("\tSHA256: %s\t\n", artifactDigest)
		rows = append(rows, row)
	}
	printTable(out, header, rows)

	return nil
}
