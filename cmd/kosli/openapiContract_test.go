package main

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

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
	payload   interface{}
	component string
	// ignore lists json field names to skip on both sides, for deliberate
	// CLI/API divergences (none needed yet).
	ignore []string
}

func (suite *OpenAPIContractTestSuite) SetupSuite() {
	resp, err := http.Get("http://localhost:8001/api/v2/openapi.json")
	require.NoError(suite.T(), err, "should fetch the OpenAPI schema from the test server")
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(suite.T(), err)
	require.NoError(suite.T(), json.Unmarshal(body, &suite.schema), "OpenAPI schema should be valid JSON")
	require.NotEmpty(suite.T(), suite.schema.Components.Schemas, "OpenAPI schema should contain components")
}

func (suite *OpenAPIContractTestSuite) TestPayloadsMatchSchema() {
	registry := []driftCase{
		{name: "create control", payload: ControlPayload{}, component: "ControlPostInput"},
		{name: "create environment", payload: CreateEnvironmentPayload{}, component: "CreateEnvironmentPutInput"},
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

// jsonFieldNames returns the wire names from a struct's json tags.
func jsonFieldNames(v interface{}) []string {
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
