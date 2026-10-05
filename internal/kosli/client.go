// Package kosli is a client for the Kosli API, with one file per tag of the
// API's OpenAPI spec.
package kosli

import (
	"context"
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

func (c *Client) endpoint(segments ...string) (string, error) {
	return url.JoinPath(c.host, append([]string{"api/v2"}, segments...)...)
}

func (c *Client) send(ctx context.Context, params *requests.RequestParams) (*Result, error) {
	params.Token = c.token
	response, err := c.sender.Do(params)
	if err != nil {
		return nil, err
	}
	return &Result{
		Raw:     []byte(response.Body),
		Created: response.Resp.StatusCode == http.StatusCreated,
	}, nil
}
