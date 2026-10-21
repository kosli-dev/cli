package main

import (
	"bytes"
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

	self, err := os.Executable()
	require.NoError(t, err)
	cmd := exec.Command(self, "-test.run=^TestMainErrorKeepsPercentEncoding$")
	cmd.Env = append(os.Environ(), "KOSLI_TEST_RUN_MAIN=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr, "expected main to exit non-zero")
	require.Contains(t, stderr.String(), "feature%2Fx")
	require.NotContains(t, stderr.String(), "%!")
}
