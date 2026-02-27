package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const evaluateTrailDesc = `Evaluate a trail against a policy.`

type evaluateTrailOptions struct {
	flowName string
}

func newEvaluateTrailCmd(out io.Writer) *cobra.Command {
	o := new(evaluateTrailOptions)
	cmd := &cobra.Command{
		Use:   "trail TRAIL-NAME",
		Short: evaluateTrailDesc,
		Long:  evaluateTrailDesc,
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)

	err := RequireFlags(cmd, []string{"flow"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *evaluateTrailOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v2/trails/%s/%s/%s", global.Host, global.Org, o.flowName, args[0])

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    url,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	var trailData interface{}
	err = json.Unmarshal([]byte(response.Body), &trailData)
	if err != nil {
		return fmt.Errorf("failed to parse trail response: %v", err)
	}

	wrapped := map[string]interface{}{
		"trail": trailData,
	}

	output, err := json.MarshalIndent(wrapped, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal output: %v", err)
	}

	_, err = fmt.Fprintln(out, string(output))
	return err
}
