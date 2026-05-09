package logger_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kosli-dev/cli/internal/logger"
)

func TestPlaceholder(t *testing.T) {
	// We use the tool cover for coverage, but if there is no _test.go file, then
	// Go use the tool covdata. At the same time, they removed covdata as a precompiled
	// binary in the distribution. This made the coverage calculation fail for some of us.
}

func TestWarnSuppressedWhenQuietEnabled(t *testing.T) {
	var out, errOut bytes.Buffer
	l := logger.NewLogger(&out, &errOut, false)
	l.QuietEnabled = true

	l.Warn("something happened")

	if errOut.Len() != 0 {
		t.Fatalf("expected no stderr output when quiet, got %q", errOut.String())
	}
}

func TestWarnEmittedWhenQuietDisabled(t *testing.T) {
	var out, errOut bytes.Buffer
	l := logger.NewLogger(&out, &errOut, false)

	l.Warn("something happened")

	if !strings.Contains(errOut.String(), "[warning] something happened") {
		t.Fatalf("expected warning on stderr, got %q", errOut.String())
	}
}

func TestQuietDoesNotSuppressInfoOrDebug(t *testing.T) {
	var out, errOut bytes.Buffer
	l := logger.NewLogger(&out, &errOut, true)
	l.QuietEnabled = true

	l.Info("hello")
	l.Debug("debug-line")

	if !strings.Contains(out.String(), "hello") {
		t.Fatalf("expected info preserved, got %q", out.String())
	}
	if !strings.Contains(errOut.String(), "[debug] debug-line") {
		t.Fatalf("expected debug preserved, got %q", errOut.String())
	}
}
