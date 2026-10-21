package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

// main() exits through log.Fatalf, so the test re-runs itself as a child
// process and checks what the child printed.
func TestMainErrorKeepsPercentEncoding(t *testing.T) {
	if os.Getenv("KOSLI_TEST_RUN_MAIN") == "1" {
		os.Args = []string{"kosli", "version", "--short=feature%2Fx"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainErrorKeepsPercentEncoding$")
	cmd.Env = append(os.Environ(), "KOSLI_TEST_RUN_MAIN=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	var exitErr *exec.ExitError
	require.True(t, errors.As(err, &exitErr), "expected main to exit non-zero, got %v", err)
	require.Contains(t, stderr.String(), "feature%2Fx")
	require.NotContains(t, stderr.String(), "%!")
}
