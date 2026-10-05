// Package kosli is a client for the Kosli API, with one file per tag of the
// API's OpenAPI spec.
package kosli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/requests"
)

// Sender sends one request to the Kosli API. *requests.Client satisfies it.
type Sender interface {
	Do(*requests.RequestParams) (*requests.HTTPResponse, error)
}

// Client calls the Kosli API for one organization on one host.
type Client struct {
	sender Sender
	host   string
	org    string
	token  string
}

// New returns a Client that sends its requests through sender.
func New(sender Sender, host, org, token string) *Client {
	return &Client{sender: sender, host: host, org: org, token: token}
}

// Result is what a call to the Kosli API returned.
type Result struct {
	// Raw is the response body exactly as the server sent it, so a command
	// can print it for --output json without re-encoding.
	Raw []byte
	// Created is true when the server created the resource (201) rather than
	// updating one that already existed (200).
	Created bool
}

// Response is a Result with its body decoded.
type Response[T any] struct {
	Result
	Value T
}

func decode[T any](result *Result) (*Response[T], error) {
	response := &Response[T]{Result: *result}
	if len(result.Raw) == 0 {
		return response, nil
	}
	if err := json.Unmarshal(result.Raw, &response.Value); err != nil {
		return nil, fmt.Errorf("failed to decode the Kosli API response: %w", err)
	}
	return response, nil
}

func (c *Client) endpoint(segments ...string) (string, error) {
	return url.JoinPath(c.host, append([]string{"api/v2"}, segments...)...)
}

func (c *Client) send(ctx context.Context, params *requests.RequestParams) (*Result, error) {
	params.Context = ctx
	params.Token = c.token
	response, err := c.sender.Do(params)
	if err != nil {
		return nil, err
	}
	if response == nil {
		// A dry run logs the request instead of sending it.
		return &Result{}, nil
	}
	return &Result{
		Raw:     []byte(response.Body),
		Created: response.Resp.StatusCode == http.StatusCreated,
	}, nil
}
