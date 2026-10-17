package kosli

import (
	"errors"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
)

// APIError is the error for a response with a status other than 200 or 201.
type APIError = requests.APIError

// IsNotFound reports whether err is an APIError for a 404.
func IsNotFound(err error) bool {
	return hasStatus(err, http.StatusNotFound)
}

// IsConflict reports whether err is an APIError for a 409.
func IsConflict(err error) bool {
	return hasStatus(err, http.StatusConflict)
}

// IsForbidden reports whether err is an APIError for a 403, which is also what
// a beta endpoint answers to an organization without the feature enabled.
func IsForbidden(err error) bool {
	return hasStatus(err, http.StatusForbidden)
}

func hasStatus(err error, status int) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == status
}
