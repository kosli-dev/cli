package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSortedTagPairs(t *testing.T) {
	for _, tc := range []struct {
		name string
		tags any
		want []string
	}{
		{
			name: "pairs are ordered by key, not by map iteration order",
			tags: map[string]any{"team": "platform", "env": "prod", "app": "api"},
			want: []string{"app=api", "env=prod", "team=platform"},
		},
		{
			name: "non-string values are rendered as values, not as %!s verbs",
			tags: map[string]any{"replicas": float64(3), "critical": true},
			want: []string{"critical=true", "replicas=3"},
		},
		{
			name: "nil tags yield no pairs",
			tags: nil,
			want: nil,
		},
		{
			name: "a non-map yields no pairs",
			tags: "team=platform",
			want: nil,
		},
		{
			name: "an empty map yields no pairs",
			tags: map[string]any{},
			want: nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, sortedTagPairs(tc.tags))
		})
	}
}

func TestFormatTags(t *testing.T) {
	for _, tc := range []struct {
		name string
		tags any
		want string
	}{
		{
			name: "tags are bracketed, comma separated and ordered by key",
			tags: map[string]any{"team": "platform", "env": "prod"},
			want: "[env=prod], [team=platform]",
		},
		{
			name: "no tags yields the empty string",
			tags: nil,
			want: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, formatTags(tc.tags))
		})
	}
}

// The table printers receive tags as a map, so without an explicit sort their
// output ordering was whatever Go's map iteration gave that run.
func TestTagRenderingIsDeterministicAcrossPrinters(t *testing.T) {
	t.Run("printFlowAsTable orders tags by key", func(t *testing.T) {
		raw := `{"name":"backend","description":"Backend service","template":"artifact","last_deployment_at":null,"tags":{"team":"platform","env":"prod","app":"api"}}`
		var buf bytes.Buffer
		require.NoError(t, printFlowAsTable(raw, &buf, 0))
		require.Contains(t, buf.String(), "Tags:                [app=api], [env=prod], [team=platform]\n")
	})

	t.Run("printEnvironmentAsTable orders tags by key", func(t *testing.T) {
		raw := `{"name":"prod","type":"K8S","description":"","state":true,"last_reported_at":null,"tags":{"team":"platform","env":"prod","app":"api"}}`
		var buf bytes.Buffer
		require.NoError(t, printEnvironmentAsTable(raw, &buf, 0))
		require.Contains(t, buf.String(), "[app=api], [env=prod], [team=platform]")
	})

	t.Run("printFlowsListAsTable orders tags by key", func(t *testing.T) {
		raw := `[{"name":"backend","description":"Backend service","tags":{"team":"platform","env":"prod","app":"api"}}]`
		var buf bytes.Buffer
		require.NoError(t, printFlowsListAsTable(raw, &buf, 1))
		require.Contains(t, buf.String(), "[app=api], [env=prod], [team=platform]")
	})

	t.Run("printEnvListAsTable orders tags by key", func(t *testing.T) {
		raw := `[{"name":"prod","type":"K8S","last_reported_at":null,"last_modified_at":null,"tags":{"team":"platform","env":"prod","app":"api"}}]`
		var buf bytes.Buffer
		require.NoError(t, printEnvListAsTable(raw, &buf, 1))
		require.Contains(t, buf.String(), "[app=api], [env=prod], [team=platform]")
	})
}

// The API omits "tags" entirely for an environment without tags, which used to
// panic on an unchecked type assertion.
func TestPrintEnvironmentAsTableWithoutTags(t *testing.T) {
	raw := `{"name":"prod","type":"K8S","description":"","state":true,"last_reported_at":null}`
	var buf bytes.Buffer
	require.NoError(t, printEnvironmentAsTable(raw, &buf, 0))
	require.Contains(t, buf.String(), "Tags:")
	require.Contains(t, buf.String(), "None")
}
