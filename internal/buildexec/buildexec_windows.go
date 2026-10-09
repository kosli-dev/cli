//go:build windows

package buildexec

import (
	"io"
	"os"
	"os/exec"
)

func setProcessGroup(*exec.Cmd, io.Reader) bool { return false }

func notifySignals(chan<- os.Signal) {}

func relaySignals(_ <-chan os.Signal, done <-chan struct{}, _ int, _ bool) { <-done }

func exitCode(state *os.ProcessState) int { return state.ExitCode() }
