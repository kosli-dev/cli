//go:build unix

package main

import (
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// The only guard in loadSbom without a test, and it could not have one while the
// check ran on an already-open handle: open(2) on a fifo with no writer blocks,
// so a reversed order hangs rather than failing. The timeout turns that hang
// into a named failure, which makes the ordering itself something a red run can
// report, not just the branch.
func TestSbomRejectsAFifoWithoutBlocking(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bom.json")
	require.NoError(t, syscall.Mkfifo(path, 0644))

	o := &attestSbomOptions{
		CommonAttestationOptions: &CommonAttestationOptions{fingerprintOptions: &fingerprintOptions{}},
		sbomFilePath:             path,
		payload:                  SbomAttestationPayload{CommonAttestationPayload: &CommonAttestationPayload{}},
	}
	done := make(chan error, 1)
	go func() { _, err := o.loadSbom(); done <- err }()

	select {
	case err := <-done:
		require.ErrorContains(t, err, "is not a regular file")
	case <-time.After(5 * time.Second):
		t.Fatal("loadSbom blocked on a fifo: the stat must precede the open")
	}
}
