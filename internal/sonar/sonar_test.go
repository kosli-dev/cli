package sonar_test

import "testing"

func TestPlaceholder(t *testing.T) {
	// We use the tool cover for coverage, but if there is no _test.go file, then
	// Go use the tool covdata. At the same time, they removed covdata as a precompiled
	// binary in the distribution. This made the coverage calculation fail for some of us.
}
