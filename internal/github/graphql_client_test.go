package github

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsNullField(t *testing.T) {
	for _, tc := range []struct {
		name string
		data string
		path []any
		want bool
	}{
		{"a null field", `{"repository":{"pullRequest":{"mergedAt":null}}}`,
			[]any{"repository", "pullRequest", "mergedAt"}, true},
		{"a field with a value", `{"repository":{"pullRequest":{"title":"A PR"}}}`,
			[]any{"repository", "pullRequest", "title"}, false},
		{"a list item", `{"repository":{"pullRequest":{"commits":{"nodes":[{"x":null}]}}}}`,
			[]any{"repository", "pullRequest", "commits", "nodes", 0.0}, false},
		{"a root field", `{"repository":null}`, []any{"repository"}, false},
		{"a field two segments down", `{"repository":{"pullRequest":null}}`,
			[]any{"repository", "pullRequest"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var data any
			require.NoError(t, json.Unmarshal([]byte(tc.data), &data))
			require.Equal(t, tc.want, isNullField(data, tc.path))
		})
	}
}
