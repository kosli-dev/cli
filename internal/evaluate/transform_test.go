package evaluate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransformTrail(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		result := TransformTrail(nil)
		require.Nil(t, result)
	})
}
