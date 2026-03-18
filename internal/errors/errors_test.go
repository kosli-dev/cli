package errors_test

import (
	"fmt"
	"testing"

	kosliErrors "github.com/kosli-dev/cli/internal/errors"
	"github.com/stretchr/testify/assert"
)

func TestExitCodeFor(t *testing.T) {
	t.Run("nil returns 0", func(t *testing.T) {
		assert.Equal(t, 0, kosliErrors.ExitCodeFor(nil))
	})

	t.Run("ErrCompliance returns 1", func(t *testing.T) {
		err := kosliErrors.NewErrCompliance("artifact is not compliant")
		assert.Equal(t, 1, kosliErrors.ExitCodeFor(err))
	})

	t.Run("wrapped ErrCompliance returns 1", func(t *testing.T) {
		inner := kosliErrors.NewErrCompliance("artifact is not compliant")
		err := fmt.Errorf("outer: %w", inner)
		assert.Equal(t, 1, kosliErrors.ExitCodeFor(err))
	})

	t.Run("ErrServer returns 2", func(t *testing.T) {
		err := kosliErrors.NewErrServer("server unreachable")
		assert.Equal(t, 2, kosliErrors.ExitCodeFor(err))
	})

	t.Run("wrapped ErrServer returns 2", func(t *testing.T) {
		inner := kosliErrors.NewErrServer("server unreachable")
		err := fmt.Errorf("outer: %w", inner)
		assert.Equal(t, 2, kosliErrors.ExitCodeFor(err))
	})

	t.Run("ErrConfig returns 3", func(t *testing.T) {
		err := kosliErrors.NewErrConfig("bad api token")
		assert.Equal(t, 3, kosliErrors.ExitCodeFor(err))
	})

	t.Run("wrapped ErrConfig returns 3", func(t *testing.T) {
		inner := kosliErrors.NewErrConfig("bad api token")
		err := fmt.Errorf("outer: %w", inner)
		assert.Equal(t, 3, kosliErrors.ExitCodeFor(err))
	})

	t.Run("ErrUsage returns 4", func(t *testing.T) {
		err := kosliErrors.NewErrUsage("missing required flag")
		assert.Equal(t, 4, kosliErrors.ExitCodeFor(err))
	})

	t.Run("wrapped ErrUsage returns 4", func(t *testing.T) {
		inner := kosliErrors.NewErrUsage("missing required flag")
		err := fmt.Errorf("outer: %w", inner)
		assert.Equal(t, 4, kosliErrors.ExitCodeFor(err))
	})

	t.Run("unknown error returns 1", func(t *testing.T) {
		err := fmt.Errorf("some unexpected error")
		assert.Equal(t, 1, kosliErrors.ExitCodeFor(err))
	})
}
