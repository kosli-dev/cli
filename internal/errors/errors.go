package errors

import (
	"errors"
	"strings"
)

// Exit codes
const (
	ExitOK         = 0
	ExitCompliance = 1
	ExitServer     = 2
	ExitConfig     = 3
	ExitUsage      = 4
)

// ErrCompliance indicates a compliance or assertion violation (exit 1).
type ErrCompliance struct{ msg string }

func (e *ErrCompliance) Error() string { return e.msg }

func NewErrCompliance(msg string) error { return &ErrCompliance{msg: msg} }

// ErrServer indicates the Kosli server is unreachable or returned a 5xx (exit 2).
type ErrServer struct{ msg string }

func (e *ErrServer) Error() string { return e.msg }

func NewErrServer(msg string) error { return &ErrServer{msg: msg} }

// ErrConfig indicates a configuration or authentication error (exit 3).
type ErrConfig struct{ msg string }

func (e *ErrConfig) Error() string { return e.msg }

func NewErrConfig(msg string) error { return &ErrConfig{msg: msg} }

// ErrUsage indicates a CLI usage error such as a missing flag or unknown command (exit 4).
type ErrUsage struct{ msg string }

func (e *ErrUsage) Error() string { return e.msg }

func NewErrUsage(msg string) error { return &ErrUsage{msg: msg} }

// ExitCodeFor returns the appropriate process exit code for err.
// Returns 0 for nil, and 1 for any unrecognised error.
func ExitCodeFor(err error) int {
	if err == nil {
		return ExitOK
	}
	var ec *ErrCompliance
	if errors.As(err, &ec) {
		return ExitCompliance
	}
	var es *ErrServer
	if errors.As(err, &es) {
		return ExitServer
	}
	var cfg *ErrConfig
	if errors.As(err, &cfg) {
		return ExitConfig
	}
	var eu *ErrUsage
	if errors.As(err, &eu) {
		return ExitUsage
	}
	if isCobraUsageError(err) {
		return ExitUsage
	}
	return ExitCompliance // default non-zero; use exit 1 for unclassified errors
}

// isCobraUsageError returns true for plain errors produced by Cobra's built-in
// flag and argument validation that were not already wrapped as ErrUsage.
// This covers cases where innerMain's wrapping is bypassed (e.g. in tests).
func isCobraUsageError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "unknown flag:") ||
		strings.Contains(msg, "unknown shorthand flag:") ||
		strings.Contains(msg, "required flag(s)") ||
		strings.Contains(msg, "accepts") && strings.Contains(msg, "arg(s)") ||
		strings.Contains(msg, "requires at least") && strings.Contains(msg, "arg(s)")
}
