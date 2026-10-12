// Package regopolicies talks to the Kosli rego policy API: it publishes a
// versioned Rego bundle under an identifier, and reads its versions back. It
// compiles and validates nothing itself; the server does that asynchronously.
package regopolicies

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/kosli-dev/cli/internal/requests"
)

// The statuses a version reports. A version leaves pending exactly once:
// rejected means the policy is at fault, failed means the server is.
const (
	StatusPending  = "pending"
	StatusValid    = "valid"
	StatusRejected = "rejected"
	StatusFailed   = "failed"
)

// PublishRequest describes one publish of a policy. An empty Name or
// Description is left out, so the server keeps the value it has.
type PublishRequest struct {
	Identifier  string
	Name        string
	Description string
	InputSchema string
	Comment     string
	Files       map[string]string
}

// Version is one published version of a policy, as the server reports it.
type Version struct {
	Identifier  string  `json:"identifier"`
	Version     int     `json:"version"`
	Status      string  `json:"status"`
	Digest      string  `json:"digest"`
	InputSchema string  `json:"input_schema"`
	Comment     string  `json:"comment"`
	CreatedAt   float64 `json:"created_at"`
	CreatedBy   string  `json:"created_by"`
	// Nil while the version is pending.
	ValidatedAt *float64 `json:"validated_at,omitempty"`
	OPAVersion  string   `json:"opa_version,omitempty"`
	// Set only on a valid version; ParamsSchema only when the policy declares one.
	Entrypoint   string         `json:"entrypoint,omitempty"`
	ParamsSchema map[string]any `json:"params_schema,omitempty"`
	// Set only on a rejected or failed version.
	Error *ValidationError `json:"error,omitempty"`
}

// IsTerminal reports whether the version has left pending, which it does once.
func (v *Version) IsTerminal() bool {
	return v.Status == StatusValid || v.Status == StatusRejected || v.Status == StatusFailed
}

// ValidationError says why a version is not valid. Kind is a plain string: the
// set belongs to the evaluator, and an unknown kind must still reach the user.
type ValidationError struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

// Client is a Kosli API client scoped to the rego policy endpoints.
type Client struct {
	http   *requests.Client
	host   string
	token  string
	dryRun bool
}

func NewClient(httpClient *requests.Client, host, token string, dryRun bool) *Client {
	return &Client{http: httpClient, host: host, token: token, dryRun: dryRun}
}

// Publish stores the next version of a policy and returns it pending. A dry run
// sends nothing and returns no version.
func (c *Client) Publish(org string, request PublishRequest) (*Version, error) {
	route, err := policyRoute(org, request.Identifier)
	if err != nil {
		return nil, err
	}
	endpoint, err := url.JoinPath(c.host, route)
	if err != nil {
		return nil, err
	}

	response, err := c.http.Do(&requests.RequestParams{
		Method: http.MethodPut,
		URL:    endpoint,
		Token:  c.token,
		DryRun: c.dryRun,
		Payload: publishPayload{
			Name:        request.Name,
			Description: request.Description,
			InputSchema: request.InputSchema,
			Comment:     request.Comment,
			Files:       request.Files,
		},
	})
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil
	}
	return decodeVersion(response.Body)
}

// GetVersion reads one version of a policy. A pending one is a normal answer,
// not an error.
func (c *Client) GetVersion(org, identifier string, version int) (*Version, error) {
	route, err := policyRoute(org, identifier)
	if err != nil {
		return nil, err
	}
	endpoint, err := url.JoinPath(c.host, route, "versions", strconv.Itoa(version))
	if err != nil {
		return nil, err
	}

	// DryRun is deliberately not set: a read writes nothing.
	response, err := c.http.Do(&requests.RequestParams{
		Method: http.MethodGet,
		URL:    endpoint,
		Token:  c.token,
	})
	if err != nil {
		return nil, err
	}
	if response == nil {
		// The shared client answers a suppressed request with nothing at all,
		// so a later change that suppresses this read meets this, not a panic.
		return nil, errors.New("no rego policy version was read back")
	}
	return decodeVersion(response.Body)
}

// policyRoute is the escaped API path of one policy. The identifier is a single
// segment, so a slash in it is escaped rather than read as a separator, and the
// names a URL path resolves on its own are refused: escaping cannot keep "."
// or ".." in their segment.
func policyRoute(org, identifier string) (string, error) {
	switch identifier {
	case "", ".", "..":
		return "", fmt.Errorf("%q is not a valid rego policy identifier", identifier)
	}
	return "/api/v2/rego-policies/" + url.PathEscape(org) + "/" + url.PathEscape(identifier), nil
}

// The payload type is separate from PublishRequest because the server forbids
// unknown fields: the wire shape has to be stated exactly here.
type publishPayload struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	InputSchema string            `json:"input_schema"`
	Comment     string            `json:"comment"`
	Files       map[string]string `json:"files"`
}

func decodeVersion(body string) (*Version, error) {
	var version Version
	if err := json.Unmarshal([]byte(body), &version); err != nil {
		return nil, fmt.Errorf("failed to parse the rego policy version: %w", err)
	}
	return &version, nil
}
