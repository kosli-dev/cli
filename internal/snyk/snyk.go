package snyk

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/owenrumney/go-sarif/v2/sarif"
)

type SnykTool struct {
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

type SnykResult struct {
	HighCount   int             `json:"high_count"`
	MediumCount int             `json:"medium_count"`
	LowCount    int             `json:"low_count"`
	High        []Vulnerability `json:"high,omitempty"`
	Medium      []Vulnerability `json:"medium,omitempty"`
	LOW         []Vulnerability `json:"low,omitempty"`
}

type SnykData struct {
	SchemaVersion int          `json:"schema_version"`
	Tool          SnykTool     `json:"tool"`
	Results       []SnykResult `json:"results"`
}

// ProcessSnykResultFile takes a path to a Snyk scan results file
// and returns a processed SnykData object from it
func ProcessSnykResultFile(file string) (*SnykData, error) {
	report, err := sarif.Open(file)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(report.Schema, "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-") {
		return nil, fmt.Errorf("invalid sarif file")
	}
	data := &SnykData{
		SchemaVersion: 1,
		Tool:          SnykTool{},
		Results:       []SnykResult{},
	}

	if len(report.Runs) > 0 {
		data.Tool.Name = report.Runs[0].Tool.Driver.Name
		if report.Runs[0].Tool.Driver.Version != nil {
			data.Tool.Version = *report.Runs[0].Tool.Driver.Version
		}
	}

	for _, run := range report.Runs {
		result := SnykResult{}
		for _, r := range run.Results {
			switch *r.Level {
			case "error":
				result.HighCount++
				result.High = append(result.High, createVulnerability(r))
			case "warning":
				result.MediumCount++
				result.Medium = append(result.Medium, createVulnerability(r))
			case "info":
				result.LowCount++
				result.Medium = append(result.Medium, createVulnerability(r))
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
