package main

import (
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

// SliceValue is satisfied by type assertion at the point of use: bindFlags
// applies a config file list through Replace, and a post-parse check reads
// GetSlice. Nothing enforces it, so a drifting signature would silently fall
// back to the string path instead of failing. This assertion makes that a
// compile error. pflag.Value needs no such assertion - the Flags().VarP call
// registering the flag already requires it.
var _ pflag.SliceValue = (*nonEmptyStringSlice)(nil)

// nonEmptyStringSlice holds the values of a multi-value flag and refuses an
// empty one.
//
// pflag's own StringSlice splits each value with a CSV reader, which yields
// nothing at all for an empty string. An empty element is therefore appended as
// nothing and leaves no trace: after parsing, `--attachments "" --attachments
// provenance.json` cannot be told apart from `--attachments provenance.json`.
// Set is the last point at which the element still exists, so it is where the
// refusal has to happen.
//
// The comma splitting is kept rather than switching to pflag's StringArray,
// which stores values verbatim. An environment variable cannot be repeated, so
// a comma list is the only way to give a multi-value flag more than one value
// from the environment.
type nonEmptyStringSlice struct {
	values  *[]string
	changed bool
}

// newNonEmptyStringSlice returns a flag value writing through to values.
func newNonEmptyStringSlice(values *[]string) *nonEmptyStringSlice {
	return &nonEmptyStringSlice{values: values}
}

// Append adds one value, refusing an empty one. It is part of pflag.SliceValue.
func (s *nonEmptyStringSlice) Append(value string) error {
	if value == "" {
		return fmt.Errorf("empty values are not allowed")
	}
	*s.values = append(*s.values, value)
	s.changed = true
	return nil
}

// Replace overwrites every value, refusing an empty one. It is part of
// pflag.SliceValue, and it is the path a config file list arrives by, so the
// refusal has to hold here as well as in Set. Checking only Set would leave the
// config file as a way in for the values the flag refuses on the command line.
func (s *nonEmptyStringSlice) Replace(values []string) error {
	for _, value := range values {
		if value == "" {
			return fmt.Errorf("empty values are not allowed")
		}
	}
	*s.values = values
	s.changed = true
	return nil
}

// GetSlice returns the values as separate elements. It is part of
// pflag.SliceValue, and it is what lets a post-parse check see the elements
// rather than the bracketed String form.
func (s *nonEmptyStringSlice) GetSlice() []string {
	return *s.values
}

// String renders the values the way pflag's own StringSlice does, so help text
// and defaults read identically for a flag that switches to this type.
func (s *nonEmptyStringSlice) String() string {
	return "[" + strings.Join(*s.values, ",") + "]"
}

// Type names the type shown in help text. It matches pflag's StringSlice so a
// flag that switches to this type keeps the same help output.
func (s *nonEmptyStringSlice) Type() string {
	return "stringSlice"
}

// Set splits value on commas and stores the result, replacing whatever the flag
// held on the first call and appending on later ones, which is how pflag
// accumulates a repeated flag. It returns an error when value is empty or
// contains an empty element. pflag prefixes that error with the flag name.
func (s *nonEmptyStringSlice) Set(value string) error {
	if value == "" {
		return fmt.Errorf("empty values are not allowed")
	}

	elements, err := csv.NewReader(strings.NewReader(value)).Read()
	if err != nil {
		return err
	}
	for _, element := range elements {
		if element == "" {
			return fmt.Errorf("empty values are not allowed")
		}
	}

	if s.changed {
		*s.values = append(*s.values, elements...)
	} else {
		*s.values = elements
		s.changed = true
	}

	return nil
}
