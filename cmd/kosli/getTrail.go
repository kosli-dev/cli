package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const getTrailDesc = `Get the metadata of a specific trail.`

type getTrailOptions struct {
	flowName string
	output   string
}

func newGetTrailCmd(out io.Writer) *cobra.Command {
	o := new(getTrailOptions)
	cmd := &cobra.Command{
		Use:   "trail TRAIL-NAME",
		Short: getTrailDesc,
		Long:  getTrailDesc,
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
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlagWithMarkdown)

	err := RequireFlags(cmd, []string{"flow"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *getTrailOptions) run(out io.Writer, args []string) error {
	url, err := url.JoinPath(global.Host, "api/v2/trails", global.Org, o.flowName, args[0])
	if err != nil {
		return err
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    url,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table":    printTrailAsTable,
			"json":     output.PrintJson,
			"markdown": printTrailAsMarkdown,
		})
}

// printTrailAsMarkdown renders a trail as GitHub-Flavored Markdown, suitable for
// piping into a CI job summary (e.g. GitHub's $GITHUB_STEP_SUMMARY or a GitLab
// summary.md artifact).
func printTrailAsMarkdown(raw string, out io.Writer, page int) error {
	var trail map[string]interface{}
	err := json.Unmarshal([]byte(raw), &trail)
	if err != nil {
		return err
	}

	lastModifiedAt, err := formattedTimestamp(trail["last_modified_at"], false)
	if err != nil {
		return err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "## Trail: %s\n\n", mdCell(trail["name"]))
	b.WriteString("| Field | Value |\n")
	b.WriteString("| --- | --- |\n")
	fmt.Fprintf(&b, "| Name | %s |\n", mdCell(trail["name"]))
	fmt.Fprintf(&b, "| Description | %s |\n", mdCell(trail["description"]))
	fmt.Fprintf(&b, "| Compliance | %s |\n", mdCell(trail["compliance_state"]))
	fmt.Fprintf(&b, "| Last modified at | %s |\n", mdCell(lastModifiedAt))

	if commitInfo, ok := trail["git_commit_info"].(map[string]interface{}); ok {
		commitTimestamp, err := formattedTimestamp(commitInfo["timestamp"], false)
		if err != nil {
			return err
		}
		b.WriteString("\n### Git commit\n\n")
		b.WriteString("| Field | Value |\n")
		b.WriteString("| --- | --- |\n")
		fmt.Fprintf(&b, "| Sha1 | %s |\n", mdCell(commitInfo["sha1"]))
		fmt.Fprintf(&b, "| Author | %s |\n", mdCell(commitInfo["author"]))
		fmt.Fprintf(&b, "| Timestamp | %s |\n", mdCell(commitTimestamp))
		if url, ok := commitInfo["url"]; ok {
			fmt.Fprintf(&b, "| URL | %s |\n", mdCell(url))
		}
		fmt.Fprintf(&b, "| Message | %s |\n", mdCell(commitInfo["message"]))
	}

	b.WriteString("\n### Events\n\n")
	if events, ok := trail["events"].([]interface{}); ok && len(events) > 0 {
		b.WriteString("| Time | Description | Git commit | Compliance |\n")
		b.WriteString("| --- | --- | --- | --- |\n")
		for _, event := range events {
			timestamp, description, commit, compliance, err := eventFields(event)
			if err != nil {
				return err
			}
			fmt.Fprintf(&b, "| %s | %s | %s | %s |\n",
				mdCell(timestamp), mdCell(description), mdCell(commit), mdCell(compliance))
		}
	} else {
		b.WriteString("_No events._\n")
	}

	_, err = fmt.Fprint(out, b.String())
	return err
}

// mdCell renders a value as a single markdown table cell, escaping characters
// that would otherwise break the table layout. CR and CRLF count as line
// endings in CommonMark, so they must be normalized along with LF.
func mdCell(v interface{}) string {
	if v == nil {
		return ""
	}
	s := fmt.Sprintf("%v", v)
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.ReplaceAll(s, "\n", "<br>")
	return s
}

func printTrailAsTable(raw string, out io.Writer, page int) error {
	var trail map[string]interface{}
	err := json.Unmarshal([]byte(raw), &trail)
	if err != nil {
		return err
	}

	header := []string{}
	rows := []string{}

	lastModifiedAt, err := formattedTimestamp(trail["last_modified_at"], false)
	if err != nil {
		return err
	}

	rows = append(rows, fmt.Sprintf("Name:\t%s", trail["name"]))
	rows = append(rows, fmt.Sprintf("Description:\t%s", trail["description"]))
	rows = append(rows, fmt.Sprintf("Compliance:\t%s", trail["compliance_state"]))
	rows = append(rows, fmt.Sprintf("Last modified at:\t%s", lastModifiedAt))
	if commitInfo, ok := trail["git_commit_info"].(map[string]interface{}); ok {
		rows = append(rows, "Git commit:\t")
		rows = append(rows, fmt.Sprintf("  Sha1:\t%s", commitInfo["sha1"].(string)))
		rows = append(rows, fmt.Sprintf("  Author:\t%s", commitInfo["author"].(string)))
		commitTimestamp, err := formattedTimestamp(commitInfo["timestamp"].(float64), false)
		if err != nil {
			return err
		}
		rows = append(rows, fmt.Sprintf("  Timestamp:\t%s", commitTimestamp))
		if url, ok := commitInfo["url"]; ok {
			rows = append(rows, fmt.Sprintf("  url:\t%s", url))
		}
		rows = append(rows, fmt.Sprintf("  message:\t%s", prefixEachLine(commitInfo["message"].(string), "\t")))
	}
	rows = append(rows, "Events:\n")

	tabFormattedPrint(out, header, rows)

	if events, ok := trail["events"].([]interface{}); ok {
		eventsHeader := []string{"\tTIME", "DESCRIPTION", "GIT-COMMIT", "COMPLIANCE"}
		eventsRows := []string{}
		for _, event := range events {
			row, err := eventRow(event)
			if err != nil {
				return err
			}
			eventsRows = append(eventsRows, row)
		}
		tabFormattedPrint(out, eventsHeader, eventsRows)
	}

	return nil
}

func eventRow(event interface{}) (string, error) {
	eventTimestamp, eventDescription, eventCommit, eventCompliance, err := eventFields(event)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("\t%s\t%s\t%s\t%s", eventTimestamp, eventDescription, eventCommit, eventCompliance), nil
}

// eventFields extracts the displayable fields of a trail event so they can be
// rendered in any output format (table, markdown).
func eventFields(event interface{}) (timestamp, description, commit, compliance string, err error) {
	eventMap := event.(map[string]interface{})
	eventTimestamp, err := formattedTimestamp(eventMap["timestamp"].(float64), true)
	if err != nil {
		return "", "", "", "", err
	}

	eventDescription := ""
	eventCompliance := ""
	if isCompliant, ok := eventMap["is_compliant"].(bool); ok {
		if isCompliant {
			eventCompliance = "compliant"
		} else if !isCompliant {
			eventCompliance = "non-compliant"
		}
	}

	eventCommit := ""
	if commitInfo, ok := eventMap["git_commit_info"].(map[string]interface{}); ok {
		if sha1, ok := commitInfo["sha1"].(string); ok {
			eventCommit = sha1[0:7]
		}
	}

	eventType := eventMap["type"].(string)
	switch eventType {
	case "trail_reported":
		eventDescription = "trail started"
	case "trail_updated":
		eventDescription = "trail updated"
	case "trail_attestation_reported":
		eventDescription = fmt.Sprintf("'%s' attestation reported for %s on the trail", eventMap["attestation_type"], eventMap["template_reference_name"])
	case "artifact_creation_reported":
		eventDescription = fmt.Sprintf("artifact '%s' created for template name '%s'", eventMap["artifact_name"], eventMap["template_reference_name"])
	case "artifact_attestation_reported", "trail_attestation_for_artifact_reported":
		eventDescription = fmt.Sprintf("'%s' attestation reported for %s.%s", eventMap["attestation_type"], eventMap["target_artifact"], eventMap["template_reference_name"])
	case "artifact_approval_reported":
		if eventMap["state"].(string) != "PENDING" {
			eventDescription = fmt.Sprintf("approval #%.0f created by '%s'", eventMap["approval_number"].(float64), eventMap["reviewer"])
		} else {
			eventDescription = fmt.Sprintf("approval #%.0f requested", eventMap["approval_number"].(float64))
		}
	case "artifact_started_running":
		eventDescription = fmt.Sprintf("artifact '%s' started running in '%s'", eventMap["template_reference_name"], eventMap["environment_name"])
	case "artifact_stopped_running":
		eventDescription = fmt.Sprintf("artifact '%s' stopped running in '%s'", eventMap["template_reference_name"], eventMap["environment_name"])
	default:
		eventDescription = eventType
	}
	return eventTimestamp, eventDescription, eventCommit, eventCompliance, nil
}
