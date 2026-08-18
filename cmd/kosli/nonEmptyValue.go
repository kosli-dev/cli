package main

import (
	"encoding/csv"
	"errors"
	"strings"

	"github.com/spf13/pflag"
)

// errEmptyFlagValue marks a value a flag was told to take that carries nothing.
// pflag wraps a Set error in *pflag.InvalidValueError, which unwraps, so the
// root command's flag-error function recognises this with errors.Is and reports
// it in one wording for every flag.
var errEmptyFlagValue = errors.New("empty flag value")

// A multi-value flag keeps the interface it had before wrapping, so anything
// asking for a flag's elements rather than its rendered string still reaches
// them. The assertion turns a wrapper that stops satisfying that into a build
// error rather than a silent fall back to the string path.
var _ pflag.SliceValue = (*nonEmptySliceValue)(nil)

// nonEmptyValue wraps a flag's value and refuses an empty one, delegating
// everything else to what it wraps.
//
// Set is where the refusal has to happen. pflag's own StringSlice splits each
// value with a CSV reader, which yields nothing for an empty string, so
// `--exclude "" --exclude vendor` reaches the end of parsing as a list holding
// only vendor: the empty element is gone and no later check can see it. Set is
// the last point at which it still exists.
//
// Values from a config file or the environment arrive the same way, because
// bindFlags applies them through Set, so one path covers every source.
type nonEmptyValue struct {
	wrapped pflag.Value
}

// nonEmptySliceValue is the wrapper for a flag holding elements. It refuses an
// empty element as well as an empty value, and keeps the element-level methods
// reachable.
//
// splitsOnCommas says whether one value becomes several elements. Only pflag's
// stringSlice does that, by reading the value as CSV. A stringArray keeps each
// value whole, so an empty element can only arrive as an empty value, and its
// values are not CSV at all: reading `--jq '.name | startswith("B")'` that way
// rejects a legitimate expression over a bare quote.
type nonEmptySliceValue struct {
	nonEmptyValue
	// elements is the same value the embedded nonEmptyValue wraps, seen through
	// the interface that reaches its elements. It is named apart from that
	// field so neither shadows the other.
	elements       pflag.SliceValue
	splitsOnCommas bool
}

// newNonEmptyValue returns wrapped with a refusal of empty values around it,
// keeping the element-level methods of a multi-value flag.
func newNonEmptyValue(wrapped pflag.Value) pflag.Value {
	if slice, ok := wrapped.(pflag.SliceValue); ok {
		return &nonEmptySliceValue{
			nonEmptyValue:  nonEmptyValue{wrapped: wrapped},
			elements:       slice,
			splitsOnCommas: wrapped.Type() == "stringSlice",
		}
	}
	return &nonEmptyValue{wrapped: wrapped}
}

// Set refuses an empty value and otherwise passes it to the wrapped value.
func (v *nonEmptyValue) Set(value string) error {
	if value == "" {
		return errEmptyFlagValue
	}
	return v.wrapped.Set(value)
}

// String renders the wrapped value, so help text and defaults are unchanged by
// the wrapping.
func (v *nonEmptyValue) String() string {
	return v.wrapped.String()
}

// Type names the wrapped value's type, so help output is unchanged by the
// wrapping.
func (v *nonEmptyValue) Type() string {
	return v.wrapped.Type()
}

// Set refuses a value holding an empty element before handing the whole value
// on, because the CSV split that produces the elements discards an empty one.
func (v *nonEmptySliceValue) Set(value string) error {
	if v.splitsOnCommas {
		if err := refuseEmptyElements(value); err != nil {
			return err
		}
	}
	return v.nonEmptyValue.Set(value)
}

// Append adds one element, refusing an empty one. It is part of
// pflag.SliceValue.
func (v *nonEmptySliceValue) Append(value string) error {
	if value == "" {
		return errEmptyFlagValue
	}
	return v.elements.Append(value)
}

// Replace overwrites every element, refusing an empty one. It is part of
// pflag.SliceValue, and it refuses here as well as in Set so that a caller
// applying a list this way cannot put in what the command line cannot.
func (v *nonEmptySliceValue) Replace(values []string) error {
	for _, value := range values {
		if value == "" {
			return errEmptyFlagValue
		}
	}
	return v.elements.Replace(values)
}

// GetSlice returns the elements. It is part of pflag.SliceValue.
func (v *nonEmptySliceValue) GetSlice() []string {
	return v.elements.GetSlice()
}

// refuseEmptyElements reports an empty element in a comma-separated value. An
// empty value carries no elements at all and is left to the caller, which
// refuses it whatever the flag's type.
func refuseEmptyElements(value string) error {
	if value == "" {
		return nil
	}
	elements, err := csv.NewReader(strings.NewReader(value)).Read()
	if err != nil {
		return err
	}
	for _, element := range elements {
		if element == "" {
			return errEmptyFlagValue
		}
	}
	return nil
}
