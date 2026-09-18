// Package evaluations talks to the Kosli server-side policy evaluation API:
// it asks for an evaluation of one or more trails against an inline Rego
// bundle, and reads the verdict back. It compiles and runs nothing itself.
package evaluations

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/requests"
)

// The statuses an evaluation reports. Only Completed and Failed are terminal;
// any other value, including one this client has never heard of, means the
// evaluation is still running. The server is free to add one, and reading an
// unknown status as terminal would turn a running evaluation into a verdict.
const (
	StatusPending   = "pending"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

// TrailRef names one trail to evaluate, spelled as the API expects it.
type TrailRef struct {
	Flow  string `json:"flow"`
	Trail string `json:"trail"`
}

// CreateRequest describes one evaluation. Every trail in it resolves at a
// single instant, which is why several trails belong in one request rather
// than in one request each.
type CreateRequest struct {
	Trails   []TrailRef
	Files    map[string]string
	Params   map[string]any
	Decision *Decision
}

// Decision is where an evaluation records its outcome, and against which
// control. The evaluation writes it itself, so the verdict never travels back
// through a caller that could assert a different one.
type Decision struct {
	Control string `json:"control"`
	Name    string `json:"name"`
	Flow    string `json:"flow"`
	Trail   string `json:"trail"`
	// Absent, the decision is about the trail rather than an artifact in it.
	Fingerprint string `json:"fingerprint,omitempty"`
}

// Result is the policy's verdict. A denial is a result, not a failure.
type Result struct {
	Allow      bool     `json:"allow"`
	Violations []string `json:"violations"`
}

// Failure is a defect the evaluator classified, such as a policy that does not
// compile. Kind is deliberately a plain string: the set is agreed with the
// evaluator rather than owned here, and an unrecognised kind must still reach
// the user rather than be flattened into a generic error.
type Failure struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

// Evaluation carries at most one of Result and Failure. Which one is present
// follows from Status, so callers branch on Status rather than on presence.
type Evaluation struct {
	ID          string   `json:"id"`
	Status      string   `json:"status"`
	RequestedAt float64  `json:"requested_at"`
	RecordedAt  float64  `json:"recorded_at"`
	Result      *Result  `json:"result"`
	Failure     *Failure `json:"error"`
	// Empty until a decision that was asked for has been written.
	DecisionAttestationID string `json:"decision_attestation_id"`
}

// IsTerminal reports whether the evaluation has finished and will not change.
func (e *Evaluation) IsTerminal() bool {
	return e.Status == StatusCompleted || e.Status == StatusFailed
}

// Client is a Kosli API client scoped to the evaluations endpoints.
type Client struct {
	http   *requests.Client
	host   string
	token  string
	dryRun bool
}

func NewClient(httpClient *requests.Client, host, token string, dryRun bool) *Client {
	return &Client{http: httpClient, host: host, token: token, dryRun: dryRun}
}

// Create asks for an evaluation and returns it pending; the verdict is read
// back separately. A dry run sends nothing and returns no evaluation.
//
// Not idempotent, and the shared client retries a server error or a network
// timeout: a create that succeeded but whose answer was lost is sent again, so
// one command can leave more than one evaluation behind. They are duplicates
// of each other rather than disagreements, since each is deterministic for the
// same policy and instant, so the cost is wasted work. Preventing it needs a
// key the caller supplies, which the API does not take.
func (c *Client) Create(org string, request CreateRequest) (*Evaluation, error) {
	endpoint, err := url.JoinPath(c.host, "api/v2/evaluations", org)
	if err != nil {
		return nil, err
	}

	params := request.Params
	if params == nil {
		// The server's field is a plain object defaulting to empty, so null
		// fails its validation where an empty object is accepted.
		params = map[string]any{}
	}

	response, err := c.http.Do(&requests.RequestParams{
		Method: http.MethodPost,
		URL:    endpoint,
		Token:  c.token,
		DryRun: c.dryRun,
		Payload: createPayload{
			Context:  createContext{Trails: request.Trails},
			Policy:   inlinePolicy{Files: request.Files},
			Params:   params,
			Decision: request.Decision,
		},
	})
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil
	}
	return decodeEvaluation(response.Body)
}

// The payload types are unexported and separate from CreateRequest because
// every model behind this endpoint forbids unknown fields: the wire shape has
// to be stated exactly here rather than inherited from a caller's struct.
type createPayload struct {
	Context  createContext  `json:"context"`
	Policy   inlinePolicy   `json:"policy"`
	Params   map[string]any `json:"params"`
	Decision *Decision      `json:"decision,omitempty"`
}

type createContext struct {
	Trails []TrailRef `json:"trails"`
}

type inlinePolicy struct {
	Files map[string]string `json:"files"`
}

func decodeEvaluation(body string) (*Evaluation, error) {
	var evaluation Evaluation
	if err := json.Unmarshal([]byte(body), &evaluation); err != nil {
		return nil, fmt.Errorf("failed to parse the evaluation: %w", err)
	}
	return &evaluation, nil
}
