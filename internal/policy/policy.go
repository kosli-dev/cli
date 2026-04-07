package policy

import "gopkg.in/yaml.v3"

const SchemaURL = "https://kosli.mintlify.app/schemas/policy/v1"

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
	Name string `yaml:"name,omitempty"`
	If   string `yaml:"if,omitempty"`
}

func NewPolicy() *Policy {
	return &Policy{
		Schema: SchemaURL,
	}
}

func (p *Policy) ToYAML() ([]byte, error) {
	out := p.normalized()
	return yaml.Marshal(out)
}

// normalized returns a copy with wildcard "*" names cleared so they are omitted from YAML.
func (p *Policy) normalized() *Policy {
	cp := *p
	if cp.Artifacts != nil {
		arts := *cp.Artifacts
		if len(arts.Attestations) > 0 {
			normalized := make([]AttestationRule, len(arts.Attestations))
			for i, a := range arts.Attestations {
				if a.Name == "*" {
					a.Name = ""
				}
				normalized[i] = a
			}
			arts.Attestations = normalized
		}
		cp.Artifacts = &arts
	}
	return &cp
}
