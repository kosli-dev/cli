// Package buildexec runs a build command in the foreground, passing its
// standard streams and termination signals through.
package buildexec

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

// Exit codes for a command that could not start, as POSIX shells report them.
const (
	exitNotExecutable = 126
	exitNotFound      = 127
)

// killDelay is how long a command gets to exit after a forwarded signal
// before it is killed.
const killDelay = 10 * time.Second

// Result describes a finished build command.
type Result struct {
	ExitCode int
	Duration time.Duration
}

// Run runs argv with the current environment plus extraEnv and waits for it to
// finish. The returned error is non-nil when the command could not start or
// exited non-zero; Result.ExitCode then holds the code a shell would report.
func Run(ctx context.Context, argv []string, extraEnv []string, stdin io.Reader, stdout, stderr io.Writer) (Result, error) {
	if len(argv) == 0 {
		return Result{ExitCode: exitNotFound}, errors.New("no command given")
	}
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Env = append(os.Environ(), extraEnv...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.WaitDelay = killDelay
	newGroup := setProcessGroup(cmd, stdin)

	// Registered before Start so a signal arriving during startup is relayed
	// instead of terminating kosli.
	signals := make(chan os.Signal, 1)
	notifySignals(signals)
	defer signal.Stop(signals)

	start := time.Now()
	if err := cmd.Start(); err != nil {
		return Result{ExitCode: startErrorExitCode(err), Duration: time.Since(start)}, err
	}
	done := make(chan struct{})
	go relaySignals(signals, done, cmd.Process.Pid, newGroup)
	err := cmd.Wait()
	close(done)
	result := Result{Duration: time.Since(start)}

	// The command succeeded, but a process it left behind still holds its output.
	if errors.Is(err, exec.ErrWaitDelay) {
		return result, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.ExitCode = exitCode(exitErr.ProcessState)
	} else if err != nil {
		result.ExitCode = 1
	}
	return result, err
}

func startErrorExitCode(err error) int {
	switch {
	case errors.Is(err, exec.ErrNotFound), errors.Is(err, fs.ErrNotExist):
		return exitNotFound
	case errors.Is(err, fs.ErrPermission):
		return exitNotExecutable
	default:
		return 1
	}
}
