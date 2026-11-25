package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/kosli-dev/cli/internal/docgen"
	"github.com/kosli-dev/cli/internal/gitview"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// OpenAPIContractTestSuite guards against drift between the CLI's hand-written
// request payload structs and the Kosli API's OpenAPI schema. It fetches the
// live schema from the test server (the same image the integration tests run
// against) and, for each registered case, asserts that:
//
//   - every json field the CLI struct sends exists as a property in the schema
//     component (so the CLI never sends a field the server would reject — the
//     control/env inputs are declared extra="forbid"), and
//   - every property the schema marks required is present in the CLI struct (so
//     the CLI can always satisfy the server's required fields).
//
// Optional schema properties the CLI does not (yet) expose are allowed and only
// logged — that is a deliberate coverage gap, not drift.
//
// To extend coverage, add one line to the registry in TestPayloadsMatchSchema.
type OpenAPIContractTestSuite struct {
	suite.Suite
	schema openAPISchema
}

type openAPISchema struct {
	Paths      map[string]map[string]json.RawMessage `json:"paths"`
	Components struct {
		Schemas map[string]componentSchema `json:"schemas"`
	} `json:"components"`
}

type componentSchema struct {
	Properties map[string]json.RawMessage `json:"properties"`
	Required   []string                   `json:"required"`
}

// driftCase maps a CLI payload struct to the OpenAPI component it must match.
type driftCase struct {
	name      string
	payload   any
	component string
	// ignore lists json field names to skip on both sides, for deliberate
	// CLI/API divergences (none needed yet).
	ignore []string
}

func (suite *OpenAPIContractTestSuite) SetupSuite() {
	resp, err := http.Get("http://localhost:8001/api/v2/openapi.json")
	require.NoError(suite.T(), err, "should fetch the OpenAPI schema from the test server")
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	require.NoError(suite.T(), err)
	require.NoError(suite.T(), json.Unmarshal(body, &suite.schema), "OpenAPI schema should be valid JSON")
	require.NotEmpty(suite.T(), suite.schema.Components.Schemas, "OpenAPI schema should contain components")
}

func (suite *OpenAPIContractTestSuite) TestPayloadsMatchSchema() {
	registry := []driftCase{
		{name: "create control", payload: ControlPayload{}, component: "ControlPostInput"},
		{name: "create environment", payload: CreateEnvironmentPayload{}, component: "CreateEnvironmentPutInput"},
		{name: "git commit info", payload: gitview.BasicCommitInfo{}, component: "GitCommit"},
	}

	for _, c := range registry {
		suite.Run(c.name, func() {
			t := suite.T()
			component, ok := suite.schema.Components.Schemas[c.component]
			require.True(t, ok, "OpenAPI component %q not found — it may have been renamed or removed", c.component)

			structFields := jsonFieldNames(c.payload)

			// 1. Every field the CLI sends must exist in the schema.
			for _, field := range structFields {
				if contains(c.ignore, field) {
					continue
				}
				_, exists := component.Properties[field]
				require.True(t, exists,
					"CLI struct %T sends field %q which is not a property of OpenAPI component %q — schema drift",
					c.payload, field, c.component)
			}

			// 2. Every required schema property must be covered by the CLI struct.
			for _, req := range component.Required {
				if contains(c.ignore, req) {
					continue
				}
				require.Contains(t, structFields, req,
					"OpenAPI component %q requires %q but CLI struct %T does not send it — schema drift",
					c.component, req, c.payload)
			}

			// Surface (without failing) optional schema properties the CLI does
			// not expose yet, so the coverage gap is visible.
			for prop := range component.Properties {
				if !contains(structFields, prop) && !contains(c.ignore, prop) {
					t.Logf("note: OpenAPI component %q has property %q not exposed by CLI struct %T", c.component, prop, c.payload)
				}
			}
		})
	}
}

// TestJSONOutputOperationsExistInSchema guards the API reference links on the
// CLI reference pages: every operation a command names in its jsonOutputAPI
// annotation must be an operation the server publishes, and the link text the
// docs generator derives for it must be that operation's summary.
func (suite *OpenAPIContractTestSuite) TestJSONOutputOperationsExistInSchema() {
	summaries := suite.apiReferenceSummaries()
	require.NotEmpty(suite.T(), summaries, "OpenAPI schema should contain operations")

	root, err := newRootCmd(&bytes.Buffer{}, &bytes.Buffer{}, nil)
	require.NoError(suite.T(), err)

	checked := 0
	var walk func(cmd *cobra.Command)
	walk = func(cmd *cobra.Command) {
		for _, sub := range cmd.Commands() {
			walk(sub)
		}
		for _, operation := range docgen.SplitOperations(cmd.Annotations[docgen.JSONOutputAPIAnnotation]) {
			checked++
			summary, published := summaries[operation]
			require.Truef(suite.T(), published,
				"%s names API operation %q, which the server's OpenAPI schema does not publish",
				cmd.CommandPath(), operation)
			require.Equalf(suite.T(), summary, docgen.OperationTitle(operation),
				"%s: the link text for %q differs from the operation's summary; add the lost word to docgen's titleWords",
				cmd.CommandPath(), operation)
		}
	}
	walk(root)
	require.NotZero(suite.T(), checked, "no command names an API operation")
}

// apiReferenceSummaries maps the docs-site path of every operation in the
// schema to its summary. The docs site builds each endpoint page's path from
// the operation's first tag and its summary, so this must keep to the same rule.
func (suite *OpenAPIContractTestSuite) apiReferenceSummaries() map[string]string {
	summaries := map[string]string{}
	for _, item := range suite.schema.Paths {
		for method, raw := range item {
			switch method {
			case "get", "post", "put", "patch", "delete":
			default:
				continue
			}
			var operation struct {
				Tags    []string `json:"tags"`
				Summary string   `json:"summary"`
			}
			require.NoError(suite.T(), json.Unmarshal(raw, &operation))
			if len(operation.Tags) == 0 || operation.Summary == "" {
				continue
			}
			summaries[apiReferenceSlug(operation.Tags[0])+"/"+apiReferenceSlug(operation.Summary)] = operation.Summary
		}
	}
	return summaries
}

var (
	apiReferenceSlugDropped = regexp.MustCompile(`[^a-z0-9\s-]`)
	apiReferenceSlugSpaces  = regexp.MustCompile(`\s+`)
)

// apiReferenceSlug lowercases s, drops punctuation and joins words with hyphens.
func apiReferenceSlug(s string) string {
	s = apiReferenceSlugDropped.ReplaceAllString(strings.ToLower(s), "")
	return apiReferenceSlugSpaces.ReplaceAllString(strings.TrimSpace(s), "-")
}

// jsonFieldNames returns the wire names from a struct's json tags.
func jsonFieldNames(v any) []string {
	t := reflect.TypeOf(v)
	names := []string{}
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		name := strings.Split(tag, ",")[0]
		if name == "" || name == "-" {
			continue
		}
		names = append(names, name)
	}
	return names
}

func contains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

func TestOpenAPIContractTestSuite(t *testing.T) {
	suite.Run(t, new(OpenAPIContractTestSuite))
}
