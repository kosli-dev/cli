package main

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogCommandAliases(t *testing.T) {
	root, err := newRootCmd(io.Discard, io.Discard, nil)
	require.NoError(t, err)

	for _, args := range [][]string{{"logs", "environment"}, {"logs", "env"}, {"lo", "env"}} {
		cmd, _, err := root.Find(args)
		require.NoError(t, err)
		require.Equal(t, "kosli log environment", cmd.CommandPath())
	}
}
