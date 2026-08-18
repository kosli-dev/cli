package main

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

// wrappedSliceFlag returns a stringSlice flag's value with the refusal wrapped
// around it, as the walk over the command tree does, reached through the
// interface a caller working with elements rather than a rendered string uses.
func wrappedSliceFlag(t *testing.T) (pflag.SliceValue, *[]string) {
	t.Helper()
	var values []string
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.StringSliceVar(&values, "exclude", []string{}, "")

	wrapped := newNonEmptyValue(flags.Lookup("exclude").Value)
	slice, ok := wrapped.(pflag.SliceValue)
	require.True(t, ok, "a wrapped multi-value flag should still hold elements")
	return slice, &values
}

// TestWrappedFlagAppendRejectsAnEmptyElement pins that the refusal covers
// Append as well as Set. Nothing calls Append today, so this is what makes the
// first caller safe rather than a way round the check.
func TestWrappedFlagAppendRejectsAnEmptyElement(t *testing.T) {
	slice, _ := wrappedSliceFlag(t)

	require.Error(t, slice.Append(""))
}

// TestWrappedFlagAppendKeepsAValue pins that Append still adds a real element,
// so wrapping refuses without changing what the flag does.
func TestWrappedFlagAppendKeepsAValue(t *testing.T) {
	slice, values := wrappedSliceFlag(t)

	require.NoError(t, slice.Append("vendor"))

	require.Equal(t, []string{"vendor"}, *values)
	require.Equal(t, []string{"vendor"}, slice.GetSlice())
}

// TestWrappedFlagReplaceRejectsAnEmptyElement pins that a caller applying a
// whole list cannot put in an element the flag refuses on the command line.
func TestWrappedFlagReplaceRejectsAnEmptyElement(t *testing.T) {
	slice, _ := wrappedSliceFlag(t)

	require.Error(t, slice.Replace([]string{"node_modules", "", "vendor"}))
}

// TestWrappedFlagReplaceOverwrites pins that Replace discards whatever the flag
// held, which is what pflag documents it to do.
func TestWrappedFlagReplaceOverwrites(t *testing.T) {
	slice, values := wrappedSliceFlag(t)
	require.NoError(t, slice.Append("node_modules"))

	require.NoError(t, slice.Replace([]string{"vendor", "build"}))

	require.Equal(t, []string{"vendor", "build"}, *values)
}
