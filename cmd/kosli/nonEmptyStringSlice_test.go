package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNonEmptyStringSliceRejectsAnEmptyElement pins the whole point of the
// type. pflag's own StringSlice runs each value through a CSV reader, and that
// reader yields nothing for an empty string, so an empty element is appended as
// nothing and leaves no trace behind. Refusing it at Set is the only place the
// element still exists to be seen.
func TestNonEmptyStringSliceRejectsAnEmptyElement(t *testing.T) {
	for _, value := range []string{"", "a,,b", ",a", "a,"} {
		t.Run(value, func(t *testing.T) {
			var values []string
			slice := newNonEmptyStringSlice(&values)

			require.Error(t, slice.Set(value))
		})
	}
}

// TestNonEmptyStringSliceKeepsCommaSplitting pins the behaviour that stops this
// being solved with pflag's StringArray instead. An environment variable cannot
// be repeated, so a comma list is the only way to give a multi-value flag more
// than one value from the environment.
func TestNonEmptyStringSliceKeepsCommaSplitting(t *testing.T) {
	var values []string
	slice := newNonEmptyStringSlice(&values)

	require.NoError(t, slice.Set("a,b"))

	require.Equal(t, []string{"a", "b"}, values)
}

// TestNonEmptyStringSliceReplaceRejectsAnEmptyElement pins that the refusal
// covers the pflag.SliceValue path as well as Set. bindFlags applies a config
// file list through Replace, so a type that checked only Set would leave the
// config file as a way in for exactly the values the flag refuses on the
// command line.
func TestNonEmptyStringSliceReplaceRejectsAnEmptyElement(t *testing.T) {
	var values []string
	slice := newNonEmptyStringSlice(&values)

	require.Error(t, slice.Replace([]string{"a", "", "b"}))
}

// TestNonEmptyStringSliceReplaceOverwrites pins that Replace discards whatever
// the flag held, which is what pflag documents it to do.
func TestNonEmptyStringSliceReplaceOverwrites(t *testing.T) {
	var values []string
	slice := newNonEmptyStringSlice(&values)
	require.NoError(t, slice.Set("a"))

	require.NoError(t, slice.Replace([]string{"b", "c"}))

	require.Equal(t, []string{"b", "c"}, values)
}

// TestNonEmptyStringSliceAppendRejectsAnEmptyElement pins the same refusal on
// the remaining pflag.SliceValue method.
func TestNonEmptyStringSliceAppendRejectsAnEmptyElement(t *testing.T) {
	var values []string
	slice := newNonEmptyStringSlice(&values)

	require.Error(t, slice.Append(""))
}

// TestNonEmptyStringSliceGetSliceReturnsTheValues pins the accessor that lets a
// post-parse check see the elements, rather than the bracketed String form.
func TestNonEmptyStringSliceGetSliceReturnsTheValues(t *testing.T) {
	var values []string
	slice := newNonEmptyStringSlice(&values)
	require.NoError(t, slice.Set("a,b"))

	require.Equal(t, []string{"a", "b"}, slice.GetSlice())
}

// TestNonEmptyStringSliceAppendsOnRepeatedSet pins that a flag given more than
// once accumulates its values, which is how pflag applies a repeated flag.
func TestNonEmptyStringSliceAppendsOnRepeatedSet(t *testing.T) {
	var values []string
	slice := newNonEmptyStringSlice(&values)

	require.NoError(t, slice.Set("a"))
	require.NoError(t, slice.Set("b,c"))

	require.Equal(t, []string{"a", "b", "c"}, values)
}
