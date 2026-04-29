package sarif

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/owenrumney/go-sarif/v2/sarif"
)

type SarifTool struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Location struct {
	URI   string `json:"uri"`
	Lines string `json:"lines,omitempty"`
}

type Vulnerability struct {
	ID            string     `json:"id"`
	Message       string     `json:"message"`
	Locations     []Location `json:"locations,omitempty"`
	PriorityScore float64    `json:"priority_score,omitempty"`
}

type SarifResult struct {
	HighCount   int             `json:"high_count"`
	MediumCount int             `json:"medium_count"`
	LowCount    int             `json:"low_count"`
	High        []Vulnerability `json:"high,omitempty"`
	Medium      []Vulnerability `json:"medium,omitempty"`
	Low         []Vulnerability `json:"low,omitempty"`
}

type SarifData struct {
	SchemaVersion int           `json:"schema_version"`
	Tool          SarifTool     `json:"tool"`
	Results       []SarifResult `json:"results"`
}

// ProcessSarifResultFile takes a path to a SARIF scan results file
// and returns a processed SarifData object from it. The parser is
// generic over SARIF v2.1.0 producers (Snyk, Checkov, Trivy, Semgrep, etc.)
// and uses Snyk-specific property fallbacks only when the standard SARIF
// `level` field is absent.
func ProcessSarifResultFile(file string) (*SarifData, error) {
	report, err := sarif.Open(file)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(report.Schema, "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/") && !strings.HasPrefix(report.Schema, "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-") && !strings.HasPrefix(report.Schema, "https://docs.oasis-open.org/sarif/sarif/v2.1.0/errata01/os/schemas/sarif-schema-2.1.0.json") {
		return nil, fmt.Errorf("invalid sarif file")
	}
	data := &SarifData{
		SchemaVersion: 1,
		Tool:          SarifTool{},
		Results:       []SarifResult{},
	}

	if len(report.Runs) > 0 {
		data.Tool.Name = report.Runs[0].Tool.Driver.Name
		if report.Runs[0].Tool.Driver.Version != nil {
			data.Tool.Version = *report.Runs[0].Tool.Driver.Version
		}
	}

	for _, run := range report.Runs {
		result := SarifResult{}
		for _, r := range run.Results {
			level := r.Level
			vulnerability := createVulnerability(r)
			if level == nil {
				ruleLevel, err := findLevel(run, vulnerability.ID)
				if err != nil {
					return nil, err
				}
				level = &ruleLevel
			}
			// levels in sarif and JSON snyk output files differ
			// mapping can be found here: https://docs.snyk.io/snyk-cli/scan-and-maintain-projects-using-the-cli/snyk-cli-for-snyk-code/view-snyk-code-cli-results#severity-levels-in-json-and-sarif-files
			switch *level {
			case "error", "high", "critical":
				result.HighCount++
				result.High = append(result.High, vulnerability)
			case "warning", "medium":
				result.MediumCount++
				result.Medium = append(result.Medium, vulnerability)
			case "info", "low", "note":
				result.LowCount++
				result.Low = append(result.Low, vulnerability)
			}

		}
		data.Results = append(data.Results, result)
	}

	return data, nil
}

func createVulnerability(r *sarif.Result) Vulnerability {
	locations := []Location{}
	for _, l := range r.Locations {
		if l != nil && l.PhysicalLocation != nil {
			lines := ""
			if l.PhysicalLocation.Region != nil && l.PhysicalLocation.Region.StartLine != nil {
				lines = strconv.Itoa(*l.PhysicalLocation.Region.StartLine)
				if l.PhysicalLocation.Region.EndLine != nil && *l.PhysicalLocation.Region.EndLine != *l.PhysicalLocation.Region.StartLine {
					lines += fmt.Sprintf("-%d", *l.PhysicalLocation.Region.EndLine)
				}
			}
			uri := ""
			if l.PhysicalLocation.ArtifactLocation != nil && l.PhysicalLocation.ArtifactLocation.URI != nil {
				uri = *l.PhysicalLocation.ArtifactLocation.URI
			}

			locations = append(locations, Location{
				URI:   uri,
				Lines: lines,
			})
		}
	}
	vul := Vulnerability{
		Locations: locations,
	}
	if r.RuleID != nil {
		vul.ID = *r.RuleID
	}
	if r.Message.Text != nil {
		vul.Message = *r.Message.Text
	}
	if value, exists := r.Properties["priorityScore"]; exists {
		if value != nil {
			vul.PriorityScore = value.(float64)
		}
	}
	return vul
}

func findLevel(r *sarif.Run, id string) (string, error) {
	ruleDesc, err := r.GetRuleById(id)
	if err != nil {
		return "", fmt.Errorf("could not find rule ID: %s. %s", id, err)
	}
	// defaultConfig := ruleDesc.DefaultConfiguration
	// if defaultConfig != nil {
	// 	return defaultConfig.Level, nil
	// }
	problem, problem_exists := ruleDesc.Properties["problem"]
	if problem_exists && problem != nil {
		severity, severity_exists := problem.(map[string]interface{})["severity"]
		if severity_exists {
			return severity.(string), nil
		}
	}
	return "", fmt.Errorf("could not find level for rule ID: %s", id)
}
