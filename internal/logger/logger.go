package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

type Logger struct {
	DebugEnabled bool
	QuietEnabled bool
	Out          io.Writer
	ErrOut       io.Writer
	warnLog      *log.Logger
	debugLog     *log.Logger
	infoLog      *log.Logger
	errLog       *log.Logger
	mu           *sync.Mutex
	// label names the host a line is about when a command talks to several.
	label string
	// infoAsDebug shows info lines only with --debug, for a secondary host.
	infoAsDebug bool
}

func NewStandardLogger() *Logger {
	return NewLogger(os.Stdout, os.Stderr, false)
}

func NewLogger(infoOut, errOut io.Writer, debug bool) *Logger {
	mu := &sync.Mutex{}
	return &Logger{
		DebugEnabled: debug,
		Out:          infoOut,
		ErrOut:       errOut,
		mu:           mu,
		warnLog:      log.New(lockedWriter{mu, errOut}, "", 0),
		debugLog:     log.New(lockedWriter{mu, errOut}, "", 0),
		errLog:       log.New(lockedWriter{mu, errOut}, "", 0),
		infoLog:      log.New(lockedWriter{mu, infoOut}, "", 0),
	}
}

// lockedWriter makes the info and error streams take turns, so lines logged
// from several goroutines, such as one per Kosli host, never interleave
// within a write even when both streams lead to one place.
type lockedWriter struct {
	mu *sync.Mutex
	w  io.Writer
}

func (lw lockedWriter) Write(p []byte) (int, error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()
	return lw.w.Write(p)
}

func (l *Logger) SetErrOut(out io.Writer) {
	l.ErrOut = out
	l.errLog.SetOutput(lockedWriter{l.mu, out})
	l.warnLog.SetOutput(lockedWriter{l.mu, out})
	l.debugLog.SetOutput(lockedWriter{l.mu, out})
}

func (l *Logger) SetInfoOut(out io.Writer) {
	l.Out = out
	l.infoLog.SetOutput(lockedWriter{l.mu, out})
}

// ForHost returns a logger that names label on its debug and warning lines.
// A secondary host's info lines show only with --debug. It writes through the
// same outputs as l, so lines from concurrent hosts do not tear.
func (l *Logger) ForHost(label string, secondary bool) *Logger {
	c := *l
	c.label = label
	c.infoAsDebug = secondary
	return &c
}

func (l *Logger) Debug(format string, v ...any) {
	if l.DebugEnabled {
		l.debugLog.Printf("[debug] %s%s\n", l.label, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Warn(format string, v ...any) {
	if l.QuietEnabled {
		return
	}
	l.warnLog.Printf("[warning] %s%s\n", l.label, fmt.Sprintf(format, v...))
}

func (l *Logger) Error(format string, v ...any) {
	l.errLog.Fatalf("Error: %s\n", fmt.Sprintf(format, v...))
}

func (l *Logger) Info(format string, v ...any) {
	if l.infoAsDebug {
		l.Debug(format, v...)
		return
	}
	l.infoLog.Printf("%s\n", fmt.Sprintf(format, v...))
}

// Print writes to the info output without appending a trailing newline
// (log.Logger always appends one), e.g. for inline prompts.
func (l *Logger) Print(format string, v ...any) {
	_, _ = fmt.Fprintf(l.infoLog.Writer(), format, v...)
}
