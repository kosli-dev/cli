package main

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHomeDir(t *testing.T) {
	t.Run("returns HOME when it is set", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", t.TempDir())

		got, err := homeDir()
		require.NoError(t, err)
		require.Equal(t, home, got)
	})

	t.Run("falls back to the platform home directory when HOME is empty", func(t *testing.T) {
		profile := t.TempDir()
		t.Setenv("HOME", "")
		t.Setenv("USERPROFILE", profile)

		got, err := homeDir()
		if runtime.GOOS == "windows" {
			require.NoError(t, err)
			require.Equal(t, profile, got)
		} else {
			require.Error(t, err)
		}
	})
}
