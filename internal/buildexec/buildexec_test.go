//go:build !windows

package buildexec

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func run(t *testing.T, argv ...string) (Result, string, string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	res, err := Run(context.Background(), argv, []string{"KOSLI_TEST_EXTRA=extra"}, nil, &stdout, &stderr)
	return res, stdout.String(), stderr.String(), err
}

func TestRunExitCodes(t *testing.T) {
	notExecutable := filepath.Join(t.TempDir(), "build.sh")
	require.NoError(t, os.WriteFile(notExecutable, []byte("#!/bin/sh\n"), 0o644))

	for _, tt := range []struct {
		name     string
		argv     []string
		wantCode int
		wantErr  bool
	}{
		{name: "success exits 0", argv: []string{"true"}, wantCode: 0},
		{name: "exit code passes through", argv: []string{"sh", "-c", "exit 3"}, wantCode: 3, wantErr: true},
		{name: "command not on PATH exits 127", argv: []string{"kosli-no-such-command"}, wantCode: 127, wantErr: true},
		{name: "missing command path exits 127", argv: []string{"./kosli-no-such-command"}, wantCode: 127, wantErr: true},
		{name: "non-executable command exits 126", argv: []string{notExecutable}, wantCode: 126, wantErr: true},
		{name: "command killed by a signal exits 128 + signal", argv: []string{"sh", "-c", "kill -TERM $$"}, wantCode: 128 + int(syscall.SIGTERM), wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			res, _, _, err := run(t, tt.argv...)
			require.Equal(t, tt.wantErr, err != nil, "error: %v", err)
			require.Equal(t, tt.wantCode, res.ExitCode)
		})
	}
}

func TestRunPassesStreamsAndEnv(t *testing.T) {
	res, stdout, stderr, err := run(t, "sh", "-c", `printf out; printf err >&2; printf " %s %s" "$KOSLI_TEST_EXTRA" "$HOME"`)
	require.NoError(t, err)
	require.Equal(t, 0, res.ExitCode)
	require.Equal(t, "out extra "+os.Getenv("HOME"), stdout)
	require.Equal(t, "err", stderr)
}

func TestRunForwardsSIGTERMToGrandchild(t *testing.T) {
	dir := t.TempDir()
	ready := filepath.Join(dir, "ready")
	marker := filepath.Join(dir, "marker")
	// "; true" stops the outer shell from exec-ing the inner one, so the inner
	// shell is a grandchild of kosli.
	inner := `trap "touch ` + marker + `; exit 0" TERM; touch ` + ready + `; while :; do sleep 0.1; done`
	done := make(chan Result)
	go func() {
		res, _, _, _ := run(t, "sh", "-c", `sh -c '`+inner+`'; true`)
		done <- res
	}()

	require.Eventually(t, func() bool { _, err := os.Stat(ready); return err == nil }, 5*time.Second, 20*time.Millisecond)
	require.NoError(t, syscall.Kill(os.Getpid(), syscall.SIGTERM))

	select {
	case res := <-done:
		require.Equal(t, 128+int(syscall.SIGTERM), res.ExitCode)
	case <-time.After(5 * time.Second):
		t.Fatal("build command was not stopped by SIGTERM")
	}
	require.Eventually(t, func() bool { _, err := os.Stat(marker); return err == nil }, 5*time.Second, 20*time.Millisecond)
}
