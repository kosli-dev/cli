package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToYAML_EmptyPolicy(t *testing.T) {
	p := NewPolicy()
	out, err := p.ToYAML()
	require.NoError(t, err)

	expected := "_schema: https://docs.kosli.com/schemas/policy/v1\n"
	assert.Equal(t, expected, string(out))
}

func TestToYAML_ProvenanceRequired(t *testing.T) {
	p := NewPolicy()
	p.Artifacts = &ArtifactRules{
		Provenance: &BooleanRule{Required: true},
	}
	out, err := p.ToYAML()
	require.NoError(t, err)

	expected := `_schema: https://docs.kosli.com/schemas/policy/v1
artifacts:
    provenance:
        required: true
`
	assert.Equal(t, expected, string(out))
}

func TestToYAML_ProvenanceWithExceptions(t *testing.T) {
	p := NewPolicy()
	p.Artifacts = &ArtifactRules{
		Provenance: &BooleanRule{
			Required: true,
			Exceptions: []ExceptionRule{
				{If: `${{ matches(artifact.name, "^datadog:.*") }}`},
			},
		},
	}
	out, err := p.ToYAML()
	require.NoError(t, err)

	expected := `_schema: https://docs.kosli.com/schemas/policy/v1
artifacts:
    provenance:
        required: true
        exceptions:
            - if: ${{ matches(artifact.name, "^datadog:.*") }}
`
	assert.Equal(t, expected, string(out))
}

func TestToYAML_TrailComplianceWithExceptions(t *testing.T) {
	p := NewPolicy()
	p.Artifacts = &ArtifactRules{
		TrailCompliance: &BooleanRule{
			Required: true,
			Exceptions: []ExceptionRule{
				{If: `${{ flow.name == "legacy" }}`},
			},
		},
	}
	out, err := p.ToYAML()
	require.NoError(t, err)

	expected := `_schema: https://docs.kosli.com/schemas/policy/v1
artifacts:
    trail-compliance:
        required: true
        exceptions:
            - if: ${{ flow.name == "legacy" }}
`
	assert.Equal(t, expected, string(out))
}

func TestToYAML_SingleAttestation(t *testing.T) {
	p := NewPolicy()
	p.Artifacts = &ArtifactRules{
		Attestations: []AttestationRule{
			{Type: "snyk", Name: "*"},
		},
	}
	out, err := p.ToYAML()
	require.NoError(t, err)

	expected := `_schema: https://docs.kosli.com/schemas/policy/v1
artifacts:
    attestations:
        - type: snyk
          name: '*'
`
	assert.Equal(t, expected, string(out))
}

func TestToYAML_AttestationWithNameAndIf(t *testing.T) {
	p := NewPolicy()
	p.Artifacts = &ArtifactRules{
		Attestations: []AttestationRule{
			{
				Type: "pull_request",
				Name: "pr-check",
				If:   `${{ flow.tags.risk-level == "high" }}`,
			},
		},
	}
	out, err := p.ToYAML()
	require.NoError(t, err)

	expected := `_schema: https://docs.kosli.com/schemas/policy/v1
artifacts:
    attestations:
        - type: pull_request
          name: pr-check
          if: ${{ flow.tags.risk-level == "high" }}
`
	assert.Equal(t, expected, string(out))
}

func TestToYAML_MultipleAttestations(t *testing.T) {
	p := NewPolicy()
	p.Artifacts = &ArtifactRules{
		Attestations: []AttestationRule{
			{Type: "snyk", Name: "security-scan"},
			{Type: "junit", Name: "*"},
			{Type: "custom:coverage-metrics", Name: "coverage"},
		},
	}
	out, err := p.ToYAML()
	require.NoError(t, err)

	expected := `_schema: https://docs.kosli.com/schemas/policy/v1
artifacts:
    attestations:
        - type: snyk
          name: security-scan
        - type: junit
          name: '*'
        - type: custom:coverage-metrics
          name: coverage
`
	assert.Equal(t, expected, string(out))
}

func TestToYAML_FullPolicy(t *testing.T) {
	p := NewPolicy()
	p.Artifacts = &ArtifactRules{
		Provenance: &BooleanRule{
			Required: true,
			Exceptions: []ExceptionRule{
				{If: `${{ matches(artifact.name, "^datadog:.*") }}`},
			},
		},
		TrailCompliance: &BooleanRule{
			Required: true,
		},
		Attestations: []AttestationRule{
			{Type: "snyk", Name: "security-scan"},
			{
				Type: "pull_request",
				Name: "pull-request",
				If:   `${{ flow.tags.risk-level == "high" }}`,
			},
			{Type: "custom:coverage-metrics", Name: "coverage"},
		},
	}
	out, err := p.ToYAML()
	require.NoError(t, err)

	expected := `_schema: https://docs.kosli.com/schemas/policy/v1
artifacts:
    provenance:
        required: true
        exceptions:
            - if: ${{ matches(artifact.name, "^datadog:.*") }}
    trail-compliance:
        required: true
    attestations:
        - type: snyk
          name: security-scan
        - type: pull_request
          name: pull-request
          if: ${{ flow.tags.risk-level == "high" }}
        - type: custom:coverage-metrics
          name: coverage
`
	assert.Equal(t, expected, string(out))
}

func TestToYAML_WildcardNameExplicit(t *testing.T) {
	p := NewPolicy()
	p.Artifacts = &ArtifactRules{
		Attestations: []AttestationRule{
			{Type: "snyk", Name: "*"},
		},
	}
	out, err := p.ToYAML()
	require.NoError(t, err)

	// name: "*" should always be explicit in output
	expected := `_schema: https://docs.kosli.com/schemas/policy/v1
artifacts:
    attestations:
        - type: snyk
          name: '*'
`
	assert.Equal(t, expected, string(out))
}

func TestToYAML_WildcardTypeRequiresNonWildcardName(t *testing.T) {
	// When type is "*", name must not be "*" per the schema
	p := NewPolicy()
	p.Artifacts = &ArtifactRules{
		Attestations: []AttestationRule{
			{Type: "*", Name: "security-scan"},
		},
	}
	out, err := p.ToYAML()
	require.NoError(t, err)

	expected := `_schema: https://docs.kosli.com/schemas/policy/v1
artifacts:
    attestations:
        - type: '*'
          name: security-scan
`
	assert.Equal(t, expected, string(out))
}
