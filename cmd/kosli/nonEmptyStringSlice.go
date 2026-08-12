package main

import (
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

// pflag.SliceValue is the interface anything working with a flag's elements
// rather than its rendered string asserts to - reading them with GetSlice, or
// applying a list with Replace. Nothing in this CLI does so today, so these
// three methods and this assertion are future-proofing: a caller reaches the
// type by assertion, which fails silently into the string path rather than
// failing to compile, and the assertion turns that into a build error.
//
// Values from a config file or the environment do not arrive that way. bindFlags
// applies them through Set, stringified with %v, so the refusal covers them by
// the same path as the command line, and a refused value fails the command
// rather than being logged and skipped.
//
// pflag.Value needs no such assertion - the Flags().Var call registering the
// flag already requires it.
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
// pflag.SliceValue. The refusal holds here as well as in Set so that a caller
// applying a list this way cannot put in values the flag refuses on the command
// line.
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
