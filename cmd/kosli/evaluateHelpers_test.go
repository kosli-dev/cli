package main

import (
	"bytes"
	"errors"
	"testing"

	kosliErrors "github.com/kosli-dev/cli/internal/errors"
	"github.com/stretchr/testify/assert"
)

func TestPolicyDenialIsErrCompliance(t *testing.T) {
	deniedJSON := `{"allow":false,"violations":["missing approval"]}`

	t.Run("printEvaluateResultAsJson returns ErrCompliance on policy denial", func(t *testing.T) {
		err := printEvaluateResultAsJson(deniedJSON, &bytes.Buffer{}, 0)
		var ec *kosliErrors.ErrCompliance
		assert.True(t, errors.As(err, &ec), "expected ErrCompliance, got %T: %v", err, err)
	})

	t.Run("printEvaluateResultAsTable returns ErrCompliance on policy denial", func(t *testing.T) {
		err := printEvaluateResultAsTable(deniedJSON, &bytes.Buffer{}, 0)
		var ec *kosliErrors.ErrCompliance
		assert.True(t, errors.As(err, &ec), "expected ErrCompliance, got %T: %v", err, err)
	})

	t.Run("printEvaluateResultAsJson returns nil on policy allow", func(t *testing.T) {
		err := printEvaluateResultAsJson(`{"allow":true,"violations":[]}`, &bytes.Buffer{}, 0)
		assert.NoError(t, err)
	})

	t.Run("printEvaluateResultAsTable returns nil on policy allow", func(t *testing.T) {
		err := printEvaluateResultAsTable(`{"allow":true,"violations":[]}`, &bytes.Buffer{}, 0)
		assert.NoError(t, err)
	})
}
