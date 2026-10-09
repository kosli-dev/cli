//go:build !windows

package buildexec

import (
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/term"
)

// setProcessGroup starts the command in its own process group, so signals can
// reach its children too, unless stdin is a terminal: a background group
// would be stopped when it reads from the terminal.
func setProcessGroup(cmd *exec.Cmd, stdin io.Reader) bool {
	if f, ok := stdin.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		return false
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return true
}

func notifySignals(c chan<- os.Signal) {
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
}

func relaySignals(signals <-chan os.Signal, done <-chan struct{}, pid int, newGroup bool) {
	target := pid
	if newGroup {
		target = -pid
	}
	var hardKill *time.Timer
	for {
		select {
		case sig := <-signals:
			// Without its own group the command shares kosli's terminal, which
			// has already sent it the SIGINT.
			if newGroup || sig != syscall.SIGINT {
				_ = syscall.Kill(target, sig.(syscall.Signal))
			}
			if hardKill == nil {
				hardKill = time.AfterFunc(killDelay, func() { _ = syscall.Kill(target, syscall.SIGKILL) })
			}
		case <-done:
			if hardKill != nil {
				hardKill.Stop()
			}
			return
		}
	}
}

func exitCode(state *os.ProcessState) int {
	if status, ok := state.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		return 128 + int(status.Signal())
	}
	return state.ExitCode()
}
