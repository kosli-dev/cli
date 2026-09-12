package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

// byteSizeUnits maps a lower-cased unit suffix to its size in bytes. Sizes are
// binary, as Lambda's /tmp and most disk figures are. The trailing "b" or "ib"
// is stripped before lookup, so "M", "MB" and "MiB" all land on the same entry.
var byteSizeUnits = map[string]int64{
	"":  1 << 20, // a bare number is megabytes
	"b": 1,
	"k": 1 << 10,
	"m": 1 << 20,
	"g": 1 << 30,
	"t": 1 << 40,
}

// parseByteSize reads a size such as "512", "512M", "8GB" or "1.5G". A bare
// number is megabytes; a unit suffix, case-insensitive and with an optional B,
// selects kilobytes, megabytes, gigabytes or terabytes; "B" alone means bytes.
func parseByteSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("size is empty")
	}

	digits := 0
	for digits < len(s) && (s[digits] >= '0' && s[digits] <= '9' || s[digits] == '.') {
		digits++
	}
	number, unit := s[:digits], strings.TrimSpace(s[digits:])
	if number == "" || strings.ContainsFunc(unit, func(r rune) bool { return !unicode.IsLetter(r) }) {
		return 0, fmt.Errorf("%q is not a size: expected a number with an optional K, M, G or T unit, e.g. 512M", s)
	}
	value, err := strconv.ParseFloat(number, 64)
	if err != nil {
		return 0, fmt.Errorf("%q is not a size: expected a number with an optional K, M, G or T unit, e.g. 512M", s)
	}

	key := strings.ToLower(unit)
	if key != "b" {
		key = strings.TrimSuffix(strings.TrimSuffix(key, "ib"), "b")
	}
	multiplier, ok := byteSizeUnits[key]
	if !ok {
		return 0, fmt.Errorf("unknown unit %q in size %q: use K, M, G or T, optionally followed by B", unit, s)
	}

	bytes := value * float64(multiplier)
	if bytes >= math.MaxInt64 {
		return 0, fmt.Errorf("size %q is too large", s)
	}
	if bytes < 1 {
		return 0, fmt.Errorf("size %q must be at least 1 byte", s)
	}
	return int64(bytes), nil
}
