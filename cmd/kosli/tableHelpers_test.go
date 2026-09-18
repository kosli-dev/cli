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
// output ordering was whatever Go's map iteration gave that run: the released
// CLI returned three distinct orderings for these three keys across 20 runs.
// Each subtest pins the full rendering, tab padding included, so a column-width
// change cannot pass silently.
func TestTagRenderingIsSortedAcrossPrinters(t *testing.T) {
	const tags = `{"team":"platform","env":"prod","app":"api"}`

	t.Run("printFlowAsTable", func(t *testing.T) {
		raw := `{"name":"backend","description":"Backend service","template":"artifact","last_deployment_at":null,"tags":` + tags + `}`
		var buf bytes.Buffer
		require.NoError(t, printFlowAsTable(raw, &buf, 0))
		require.Equal(t, "Name:                backend\n"+
			"Description:         Backend service\n"+
			"Template:            artifact\n"+
			"Last Deployment At:  N/A\n"+
			"Tags:                [app=api], [env=prod], [team=platform]\n", buf.String())
	})

	t.Run("printEnvironmentAsTable", func(t *testing.T) {
		raw := `{"name":"prod","type":"K8S","description":"","state":true,"last_reported_at":null,"tags":` + tags + `}`
		var buf bytes.Buffer
		require.NoError(t, printEnvironmentAsTable(raw, &buf, 0))
		require.Equal(t, "Name:              prod\n"+
			"Type:              K8S\n"+
			"Description:       \n"+
			"State:             COMPLIANT\n"+
			"Last Reported At:  N/A\n"+
			"Tags:              [app=api], [env=prod], [team=platform]\n"+
			"Policies:          []\n", buf.String())
	})

	t.Run("printFlowsListAsTable", func(t *testing.T) {
		raw := `[{"name":"backend","description":"Backend service","tags":` + tags + `}]`
		var buf bytes.Buffer
		require.NoError(t, printFlowsListAsTable(raw, &buf, 1))
		require.Equal(t, "NAME     DESCRIPTION      TAGS\n"+
			"backend  Backend service  [app=api], [env=prod], [team=platform]\n", buf.String())
	})

	t.Run("printEnvListAsTable", func(t *testing.T) {
		raw := `[{"name":"prod","type":"K8S","last_reported_at":null,"last_modified_at":null,"tags":` + tags + `}]`
		var buf bytes.Buffer
		require.NoError(t, printEnvListAsTable(raw, &buf, 1))
		require.Equal(t, "NAME  TYPE  LAST REPORT  LAST MODIFIED  TAGS                                    POLICIES\n"+
			"prod  K8S                               [app=api], [env=prod], [team=platform]  []\n", buf.String())
	})
}

// Responses carry "tags": {} for an untagged resource, but get environment
// asserted the value to a map unchecked, so an absent or null key would have
// panicked. The other printers already tolerated both shapes.
func TestPrintEnvironmentAsTableWithoutTags(t *testing.T) {
	raw := `{"name":"prod","type":"K8S","description":"","state":true,"last_reported_at":null}`
	var buf bytes.Buffer
	require.NoError(t, printEnvironmentAsTable(raw, &buf, 0))
	require.Contains(t, buf.String(), "Tags:              None\n")
}

func TestFormatPlainTags(t *testing.T) {
	require.Equal(t, "", formatPlainTags(nil))
	require.Equal(t, "", formatPlainTags(map[string]any{}))
	require.Equal(t, "", formatPlainTags("not-a-map"))
	require.Equal(t, "a=1, b=x", formatPlainTags(map[string]any{"b": "x", "a": float64(1)}))
}
