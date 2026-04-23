package policy

import "gopkg.in/yaml.v3"

const SchemaURL = "https://docs.kosli.com/schemas/policy/v1"

type Policy struct {
	Schema    string         `yaml:"_schema"`
	Artifacts *ArtifactRules `yaml:"artifacts,omitempty"`
}

type ArtifactRules struct {
	Provenance      *BooleanRule      `yaml:"provenance,omitempty"`
	TrailCompliance *BooleanRule      `yaml:"trail-compliance,omitempty"`
	Attestations    []AttestationRule `yaml:"attestations,omitempty"`
}

type BooleanRule struct {
	Required   bool            `yaml:"required"`
	Exceptions []ExceptionRule `yaml:"exceptions,omitempty"`
}

type ExceptionRule struct {
	If string `yaml:"if"`
}

type AttestationRule struct {
	Type string `yaml:"type"`
	Name string `yaml:"name"`
	If   string `yaml:"if,omitempty"`
}

func NewPolicy() *Policy {
	return &Policy{
		Schema: SchemaURL,
	}
}

func (p *Policy) ToYAML() ([]byte, error) {
	return yaml.Marshal(p)
}
