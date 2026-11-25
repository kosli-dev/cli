package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kosli-dev/cli/internal/docgen"
	"github.com/stretchr/testify/require"
)

// jsonOutputAPIs maps each command that prints an API response body unchanged
// to that operation's path under /api-reference on the docs site.
var jsonOutputAPIs = map[string]string{
	"assert artifact":        "asserts/assert-artifact",
	"diff snapshots":         "envdiff/get-environment-diff",
	"get attestation-type":   "custom-attestation-types/get-custom-attestation-type",
	"get control":            "controls/get-a-control",
	"get default-org":        "user/get-default-organization",
	"get environment":        "environments/get-environment",
	"get flow":               "flows/get-flow",
	"get policy":             "policies/get-policy",
	"get repo":               "repos/get-a-repo",
	"get service-account":    "service-accounts/get-a-service-account",
	"get snapshot":           "snapshots/get-snapshot",
	"get trail":              "trails/get-trail",
	"list artifacts":         "artifacts/list-artifacts",
	"list attestation-types": "custom-attestation-types/list-custom-attestation-types",
	"list audit-logs":        "audit-log/list-membershiprole-change-history-for-an-org",
	"list controls":          "controls/list-controls",
	"list environments":      "environments/list-environments",
	"list flows":             "flows/list-flows",
	"list policies":          "policies/list-policies",
	"list rego-policies":     "rego-policies/list-rego-policies",
	"list repos":             "repos/list-repos",
	"list service-accounts":  "service-accounts/list-service-accounts-in-an-organization",
	"list snapshots":         "snapshots/list-snapshots",
	"list trails":            "trails/list-trails-for-organization",
	"log environment":        "environments/list-environment-events",
	"search":                 "search/search-artifacts-by-sha-or-fingerprint",
	"get evaluation":         "evaluations/get-an-evaluation",
	"create api-key":         "service-accounts/create-an-api-key-for-a-service-account,user/create-a-personal-api-key",
	"get api-key":            "service-accounts/get-an-api-key-for-a-service-account,user/get-a-personal-api-key",
	"list api-keys":          "service-accounts/list-api-keys-for-a-service-account,user/list-personal-api-keys",
	"get artifact":           "artifacts/get-artifact-by-fingerprint,artifacts/get-artifacts-by-commit-sha",
	"get attestation":        "attestation/get-trail-attestation,attestation/get-artifact-attestation,attestation/list-attestations",
	"list product-updates":   "product-updates/list-product-updates",
	"rotate api-key":         "service-accounts/rotate-an-api-key-for-a-service-account,user/rotate-a-personal-api-key",
	"get rego-policy":        "rego-policies/get-a-rego-policy,rego-policies/get-a-version-of-a-rego-policy",
}

// jsonOutputSources maps each command whose JSON is not an API response body
// as sent to what builds it.
var jsonOutputSources = map[string]string{
	"list product-updates": docgen.JSONOutputSourceDerived,
	"rotate api-key":       docgen.JSONOutputSourceDerived,
	"evaluate input":       docgen.JSONOutputSourceCLI,
	"evaluate policy":      docgen.JSONOutputSourceCLI,
	"evaluate trail":       docgen.JSONOutputSourceCLI,
	"evaluate trails":      docgen.JSONOutputSourceCLI,
	"publish rego-policy":  docgen.JSONOutputSourceCLI,
}

func TestJSONOutputSourceAnnotations(t *testing.T) {
	root, err := newRootCmd(&bytes.Buffer{}, &bytes.Buffer{}, nil)
	require.NoError(t, err)

	for path, source := range jsonOutputSources {
		t.Run(path, func(t *testing.T) {
			cmd, remaining, err := root.Find(strings.Fields(path))
			require.NoError(t, err)
			require.Empty(t, remaining, "command not found")
			require.Equal(t, source, cmd.Annotations[docgen.JSONOutputSourceAnnotation])
		})
	}
}

func TestJSONOutputAPIAnnotations(t *testing.T) {
	root, err := newRootCmd(&bytes.Buffer{}, &bytes.Buffer{}, nil)
	require.NoError(t, err)

	for path, operation := range jsonOutputAPIs {
		t.Run(path, func(t *testing.T) {
			cmd, remaining, err := root.Find(strings.Fields(path))
			require.NoError(t, err)
			require.Empty(t, remaining, "command not found")
			require.Equal(t, operation, cmd.Annotations[docgen.JSONOutputAPIAnnotation])
			require.NotNil(t, cmd.Flags().Lookup("output"), "annotated command has no --output flag")
		})
	}
}
