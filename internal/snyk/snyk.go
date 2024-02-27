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
		Tool: SnykTool{
			Name:    report.Runs[0].Tool.Driver.Name,
			Version: *report.Runs[0].Tool.Driver.Version,
		},
		Results: []SnykResult{},
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
		if l.PhysicalLocation != nil {
			lines := ""
			if l.PhysicalLocation.Region != nil {
				lines = strconv.Itoa(*l.PhysicalLocation.Region.StartLine)
				if l.PhysicalLocation.Region.EndLine != nil && *l.PhysicalLocation.Region.EndLine != *l.PhysicalLocation.Region.StartLine {
					lines += fmt.Sprintf("-%d", *l.PhysicalLocation.Region.EndLine)
				}
			}
			locations = append(locations, Location{
				URI:   *l.PhysicalLocation.ArtifactLocation.URI,
				Lines: lines,
			})
		}
	}
	return Vulnerability{
		ID:            *r.RuleID,
		Message:       *r.Message.Text,
		Locations:     locations,
		PriorityScore: r.Properties["priorityScore"].(float64),
	}
}
