package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/shurcooL/graphql"
)

// newGraphQLClient is the one GraphQL client every evidence query uses.
func (c *GithubConfig) newGraphQLClient(ctx context.Context) (*graphql.Client, error) {
	ghClient, err := NewGithubClientFromToken(ctx, c.Token, c.BaseURL, c.Debug)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{Transport: &nullFieldErrorTransport{
		base:  ghClient.Client().Transport,
		debug: c.debugOut(),
	}}
	return graphql.NewClient(graphqlEndpoint(c.BaseURL), httpClient), nil
}

// nullFieldErrorTransport drops each GraphQL error that only explains why one
// field came back null, so a null field reads the same with or without an
// error. The client returns any error at all as a failure and drops its path,
// so this is the only place the path can still be checked.
type nullFieldErrorTransport struct {
	base  http.RoundTripper
	debug io.Writer
}

func (t *nullFieldErrorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return resp, err
	}
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(t.withoutNullFieldErrors(body)))
	return resp, nil
}

// withoutNullFieldErrors returns body minus the errors isNullField accepts,
// or body unchanged when it cannot be read or nothing is dropped.
func (t *nullFieldErrorTransport) withoutNullFieldErrors(body []byte) []byte {
	var reply map[string]json.RawMessage
	if json.Unmarshal(body, &reply) != nil {
		return body
	}
	var errs []json.RawMessage
	var data any
	if json.Unmarshal(reply["errors"], &errs) != nil || json.Unmarshal(reply["data"], &data) != nil {
		return body
	}
	kept := []json.RawMessage{}
	for _, e := range errs {
		var parsed struct{ Path []any }
		if json.Unmarshal(e, &parsed) == nil && isNullField(data, parsed.Path) {
			if t.debug != nil {
				_, _ = fmt.Fprintf(t.debug, "[debug-github] ignoring error on null field: %s\n", e)
			}
			continue
		}
		kept = append(kept, e)
	}
	if len(kept) == len(errs) {
		return body
	}
	if len(kept) == 0 {
		delete(reply, "errors")
	} else {
		reply["errors"], _ = json.Marshal(kept)
	}
	out, err := json.Marshal(reply)
	if err != nil {
		return body
	}
	return out
}

// isNullField reports whether path names an object field that is null in
// data while everything above it is not. An error anywhere else (a list item,
// a field under a null parent, no path) is more than one missing field.
func isNullField(data any, path []any) bool {
	if len(path) == 0 {
		return false
	}
	field, ok := path[len(path)-1].(string)
	if !ok {
		return false
	}
	cur := data
	for _, seg := range path[:len(path)-1] {
		switch s := seg.(type) {
		case string:
			obj, ok := cur.(map[string]any)
			if !ok {
				return false
			}
			cur = obj[s]
		case float64:
			list, ok := cur.([]any)
			if !ok || s < 0 || int(s) >= len(list) {
				return false
			}
			cur = list[int(s)]
		default:
			return false
		}
	}
	obj, ok := cur.(map[string]any)
	if !ok {
		return false
	}
	value, present := obj[field]
	return present && value == nil
}
