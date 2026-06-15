package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
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
			"markdown": o.printTrailAsMarkdown,
		})
}

// printTrailAsMarkdown renders a trail as GitHub-Flavored Markdown, suitable for
// piping into a CI job summary (e.g. GitHub's $GITHUB_STEP_SUMMARY or a GitLab
// summary.md artifact). It is a method so the trail heading can link to the
// trail page in the Kosli app, which needs the flow name.
func (o *getTrailOptions) printTrailAsMarkdown(raw string, out io.Writer, page int) error {
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
	trailURL, err := url.JoinPath(global.Host, global.Org, "flows", o.flowName, "trails", fmt.Sprintf("%v", trail["name"]))
	if err != nil {
		trailURL = ""
	}
	heading := mdCell(trail["name"])
	if trailURL != "" {
		heading = fmt.Sprintf("[%s](%s)", heading, trailURL)
	}
	fmt.Fprintf(&b, "## Trail: %s\n\n", heading)
	b.WriteString("|  |  |\n")
	b.WriteString("| --- | --- |\n")
	fmt.Fprintf(&b, "| Name | %s |\n", mdCell(trail["name"]))
	fmt.Fprintf(&b, "| Description | %s |\n", mdCell(trail["description"]))
	fmt.Fprintf(&b, "| Compliance | %s |\n", mdComplianceState(trail["compliance_state"]))
	fmt.Fprintf(&b, "| Last modified at | %s |\n", mdCell(lastModifiedAt))
	if originURL, ok := trail["origin_url"].(string); ok && originURL != "" {
		fmt.Fprintf(&b, "| Origin | %s |\n", mdCell(originURL))
	}

	if commitInfo, ok := trail["git_commit_info"].(map[string]interface{}); ok {
		commitTimestamp, err := formattedTimestamp(commitInfo["timestamp"], false)
		if err != nil {
			return err
		}
		sha := mdCell(commitInfo["sha1"])
		if commitURL, ok := commitInfo["url"].(string); ok && commitURL != "" {
			sha = fmt.Sprintf("[%s](%s)", sha, commitURL)
		}
		b.WriteString("\n### Git commit\n\n")
		b.WriteString("|  |  |\n")
		b.WriteString("| --- | --- |\n")
		fmt.Fprintf(&b, "| Sha1 | %s |\n", sha)
		fmt.Fprintf(&b, "| Author | %s |\n", mdCell(commitInfo["author"]))
		fmt.Fprintf(&b, "| Timestamp | %s |\n", mdCell(commitTimestamp))
		fmt.Fprintf(&b, "| Message | %s |\n", mdCell(firstLine(commitInfo["message"])))
	}

	writeAttestationStatuses(&b, trail["compliance_status"], trailURL)

	b.WriteString("\n### Events\n\n")
	if events, ok := trail["events"].([]interface{}); ok && len(events) > 0 {
		b.WriteString("| Time | Description | Git commit | Compliance |\n")
		b.WriteString("| --- | --- | --- | --- |\n")
		for _, event := range events {
			e, err := eventFields(event)
			if err != nil {
				return err
			}
			commit := mdCell(e.commitSHA)
			if commit != "" && e.commitURL != "" {
				commit = fmt.Sprintf("[%s](%s)", commit, e.commitURL)
			}
			fmt.Fprintf(&b, "| %s | %s | %s | %s |\n",
				mdCell(e.timestamp), mdEventDescription(e, trailURL), commit, mdEventCompliance(e.compliance))
		}
	} else {
		b.WriteString("_No events._\n")
	}

	_, err = fmt.Fprint(out, b.String())
	return err
}

// mdCell renders a value as a single markdown table cell, escaping characters
// that would otherwise break the table layout or be swallowed as HTML (e.g.
// "<email>" in a commit author). CR and CRLF count as line endings in
// CommonMark, so they must be normalized along with LF.
func mdCell(v interface{}) string {
	if v == nil {
		return ""
	}
	s := fmt.Sprintf("%v", v)
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.ReplaceAll(s, "\n", "<br>")
	return s
}

// firstLine returns the first line of a multi-line value, e.g. a git commit
// message subject. A full commit message would dominate a CI summary table.
func firstLine(v interface{}) string {
	if v == nil {
		return ""
	}
	s := fmt.Sprintf("%v", v)
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		return s[:i]
	}
	return s
}

// mdComplianceState prefixes a trail or artifact compliance state with a
// glanceable emoji. Values come from the server: COMPLIANT / NON-COMPLIANT /
// INCOMPLETE for trails, plus MISSING for artifacts.
func mdComplianceState(v interface{}) string {
	s := mdCell(v)
	switch s {
	case "COMPLIANT":
		return "✅ " + s
	case "NON_COMPLIANT", "NON-COMPLIANT":
		return "❌ " + s
	case "INCOMPLETE", "MISSING":
		return "⏳ " + s
	default:
		return s
	}
}

// writeAttestationStatuses renders the trail's attestation compliance statuses
// as headerless two-column tables (attestation name → compliance), grouped by
// the trail and by each artifact. The attestation name links to the attestation
// on the trail page when an attestation_id is present. The section is omitted
// when the trail has no attestation statuses.
func writeAttestationStatuses(b *strings.Builder, complianceStatus interface{}, trailURL string) {
	cs, ok := complianceStatus.(map[string]interface{})
	if !ok {
		return
	}
	trailAtts, _ := cs["attestations_statuses"].([]interface{})
	artifactsStatuses, _ := cs["artifacts_statuses"].(map[string]interface{})

	artifactNames := make([]string, 0, len(artifactsStatuses))
	for name := range artifactsStatuses {
		artifactNames = append(artifactNames, name)
	}
	sort.Strings(artifactNames)

	total := len(trailAtts)
	for _, name := range artifactNames {
		if artifact, ok := artifactsStatuses[name].(map[string]interface{}); ok {
			if atts, ok := artifact["attestations_statuses"].([]interface{}); ok {
				total += len(atts)
			}
		}
	}
	if total == 0 {
		return
	}

	b.WriteString("\n### Attestations\n")

	if len(trailAtts) > 0 {
		b.WriteString("\n**Trail**\n\n")
		writeAttestationTable(b, trailAtts, trailURL)
	}

	for _, name := range artifactNames {
		artifact, ok := artifactsStatuses[name].(map[string]interface{})
		if !ok {
			continue
		}
		fmt.Fprintf(b, "\n**%s** — %s\n\n", mdCell(name), mdComplianceState(artifact["status"]))
		atts, _ := artifact["attestations_statuses"].([]interface{})
		writeAttestationTable(b, atts, trailURL)
	}
}

// writeAttestationTable writes a headerless two-column table of attestation
// name (linked when possible) and compliance status.
func writeAttestationTable(b *strings.Builder, attestations []interface{}, trailURL string) {
	b.WriteString("|  |  |\n")
	b.WriteString("| --- | --- |\n")
	for _, a := range attestations {
		att, ok := a.(map[string]interface{})
		if !ok {
			continue
		}
		name := mdCell(att["attestation_name"])
		if id, ok := att["attestation_id"].(string); ok && id != "" && trailURL != "" {
			name = fmt.Sprintf("[%s](%s?attestation_id=%s)", name, trailURL, id)
		}
		status, _ := att["status"].(string)
		unexpected, _ := att["unexpected"].(bool)
		fmt.Fprintf(b, "| %s | %s |\n", name, mdAttestationCompliance(status, att["is_compliant"], unexpected))
	}
}

// mdAttestationCompliance maps an attestation's compliance to a glanceable emoji
// label, covering every status the server produces: MISSING (not yet reported),
// COMPLETE with is_compliant true/false, and the unexpected flag (reported but
// not expected by the template).
func mdAttestationCompliance(status string, isCompliant interface{}, unexpected bool) string {
	var label string
	switch status {
	case "MISSING":
		label = "⏳ missing"
	default:
		if compliant, ok := isCompliant.(bool); ok {
			if compliant {
				label = "✅ compliant"
			} else {
				label = "❌ non-compliant"
			}
		} else {
			label = "⏳ pending"
		}
	}
	if unexpected {
		label += " (+)"
	}
	return label
}

// mdEventDescription renders an event description as a markdown cell, linking
// the parts of the description that have a page in the Kosli app:
//   - the environment name of started/stopped running events links to the
//     environment snapshot ({host}/{org}/environments/{env}/{snapshot-index}),
//     or to the environment page when no snapshot index is available
//   - the attestation reference of attestation events links to the attestation
//     on the trail page ({trail-url}?attestation_id={id})
func mdEventDescription(e trailEventFields, trailURL string) string {
	description := mdCell(e.description)

	if e.environmentName != "" {
		envURL, err := url.JoinPath(global.Host, global.Org, "environments", e.environmentName, e.snapshotIndex)
		if err == nil {
			if e.snapshotIndex == "" {
				envURL += "/"
			}
			quoted := "'" + mdCell(e.environmentName) + "'"
			description = strings.Replace(description, quoted, fmt.Sprintf("[%s](%s)", quoted, envURL), 1)
		}
	}

	if e.attestationID != "" && e.attestationRef != "" && trailURL != "" {
		ref := mdCell(e.attestationRef)
		// anchor on "for " to avoid linking an attestation type that happens
		// to share its name with the reference
		anchored := "for " + ref
		link := fmt.Sprintf("for [%s](%s?attestation_id=%s)", ref, trailURL, e.attestationID)
		description = strings.Replace(description, anchored, link, 1)
	}

	return description
}

// mdEventCompliance prefixes an event compliance value with a glanceable emoji.
func mdEventCompliance(compliance string) string {
	switch compliance {
	case "compliant":
		return "✅ " + compliance
	case "non-compliant":
		return "❌ " + compliance
	default:
		return mdCell(compliance)
	}
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
	e, err := eventFields(event)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("\t%s\t%s\t%s\t%s", e.timestamp, e.description, e.commitSHA, e.compliance), nil
}

// trailEventFields holds the displayable fields of a trail event so they can be
// rendered in any output format (table, markdown).
type trailEventFields struct {
	timestamp       string
	description     string
	commitSHA       string
	commitURL       string
	compliance      string
	environmentName string
	snapshotIndex   string
	attestationID   string
	attestationRef  string // the attestation reference as it appears in the description, e.g. "artifact.snyk-scan"
}

func eventFields(event interface{}) (trailEventFields, error) {
	eventMap := event.(map[string]interface{})
	eventTimestamp, err := formattedTimestamp(eventMap["timestamp"].(float64), true)
	if err != nil {
		return trailEventFields{}, err
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
	eventCommitURL := ""
	if commitInfo, ok := eventMap["git_commit_info"].(map[string]interface{}); ok {
		if sha1, ok := commitInfo["sha1"].(string); ok {
			eventCommit = sha1[0:7]
		}
		if commitURL, ok := commitInfo["url"].(string); ok {
			eventCommitURL = commitURL
		}
	}

	eventEnvironment := ""
	eventSnapshotIndex := ""
	eventAttestationID := ""
	eventAttestationRef := ""

	eventType := eventMap["type"].(string)
	switch eventType {
	case "trail_reported":
		eventDescription = "trail started"
	case "trail_updated":
		eventDescription = "trail updated"
	case "trail_attestation_reported":
		eventDescription = fmt.Sprintf("'%s' attestation reported for %s on the trail", eventMap["attestation_type"], eventMap["template_reference_name"])
		eventAttestationRef = fmt.Sprintf("%v", eventMap["template_reference_name"])
		if id, ok := eventMap["attestation_id"].(string); ok {
			eventAttestationID = id
		}
	case "artifact_creation_reported":
		eventDescription = fmt.Sprintf("artifact '%s' created for template name '%s'", eventMap["artifact_name"], eventMap["template_reference_name"])
	case "artifact_attestation_reported", "trail_attestation_for_artifact_reported":
		eventDescription = fmt.Sprintf("'%s' attestation reported for %s.%s", eventMap["attestation_type"], eventMap["target_artifact"], eventMap["template_reference_name"])
		eventAttestationRef = fmt.Sprintf("%v.%v", eventMap["target_artifact"], eventMap["template_reference_name"])
		if id, ok := eventMap["attestation_id"].(string); ok {
			eventAttestationID = id
		}
	case "artifact_approval_reported":
		if eventMap["state"].(string) != "PENDING" {
			eventDescription = fmt.Sprintf("approval #%.0f created by '%s'", eventMap["approval_number"].(float64), eventMap["reviewer"])
		} else {
			eventDescription = fmt.Sprintf("approval #%.0f requested", eventMap["approval_number"].(float64))
		}
	case "artifact_started_running", "artifact_stopped_running":
		verb := "started"
		if eventType == "artifact_stopped_running" {
			verb = "stopped"
		}
		eventDescription = fmt.Sprintf("artifact '%s' %s running in '%s'", eventMap["template_reference_name"], verb, eventMap["environment_name"])
		if envName, ok := eventMap["environment_name"].(string); ok {
			eventEnvironment = envName
		}
		if snapshotIndex, ok := eventMap["snapshot_index"].(float64); ok {
			eventSnapshotIndex = fmt.Sprintf("%.0f", snapshotIndex)
		}
	default:
		eventDescription = eventType
	}
	return trailEventFields{
		timestamp:       eventTimestamp,
		description:     eventDescription,
		commitSHA:       eventCommit,
		commitURL:       eventCommitURL,
		compliance:      eventCompliance,
		environmentName: eventEnvironment,
		snapshotIndex:   eventSnapshotIndex,
		attestationID:   eventAttestationID,
		attestationRef:  eventAttestationRef,
	}, nil
}
