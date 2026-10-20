package kosli

import (
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
)

// DryRun wraps next so that reads still reach the server but writes are only
// logged, with the payload that would have been sent.
func DryRun(next Sender) Sender {
	return dryRunSender{next: next}
}

type dryRunSender struct {
	next Sender
}

func (s dryRunSender) Do(params *requests.RequestParams) (*requests.HTTPResponse, error) {
	switch params.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return s.next.Do(params)
	}
	write := *params
	write.DryRun = true
	return s.next.Do(&write)
}
