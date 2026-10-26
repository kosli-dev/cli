package github

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
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
		{"a connection", `{"repository":{"pullRequest":{"reviews":null}}}`,
			[]any{"repository", "pullRequest", "reviews"}, false},
		{"a non-null connection", `{"repository":{"pullRequest":{"commits":null}}}`,
			[]any{"repository", "pullRequest", "commits"}, false},
		{"a connection's nodes", `{"repository":{"pullRequest":{"commits":{"nodes":null}}}}`,
			[]any{"repository", "pullRequest", "commits", "nodes"}, false},
		{"a commit's pull requests", `{"repository":{"object":{"associatedPullRequests":null}}}`,
			[]any{"repository", "object", "associatedPullRequests"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var data any
			require.NoError(t, json.Unmarshal([]byte(tc.data), &data))
			require.Equal(t, tc.want, isNullField(data, tc.path))
		})
	}
}

type replyTransport string

func (r replyTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK, ContentLength: int64(len(r)),
		Header: http.Header{"Content-Length": {strconv.Itoa(len(r))}},
		Body:   io.NopCloser(strings.NewReader(string(r)))}, nil
}

// TestNullFieldErrorTransport_LengthMatchesTheNewBody stops a reader that
// trusts the length from reading the old body's size.
func TestNullFieldErrorTransport_LengthMatchesTheNewBody(t *testing.T) {
	reply := `{"data":{"repository":{"pullRequest":{"mergedAt":null}}},` +
		`"errors":[{"message":"x","path":["repository","pullRequest","mergedAt"]}]}`
	resp, err := (&nullFieldErrorTransport{base: replyTransport(reply)}).RoundTrip(&http.Request{})
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Less(t, len(body), len(reply), "the error should have been dropped")
	require.Equal(t, int64(len(body)), resp.ContentLength)
	require.Empty(t, resp.Header.Get("Content-Length"))
}
