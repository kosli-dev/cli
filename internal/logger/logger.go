package logger

import (
	"fmt"
	"io"
	"log"
	"os"
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
}

func NewStandardLogger() *Logger {
	return NewLogger(os.Stdout, os.Stderr, false)
}

func NewLogger(infoOut, errOut io.Writer, debug bool) *Logger {
	return &Logger{
		DebugEnabled: debug,
		Out:          infoOut,
		ErrOut:       errOut,
		warnLog:      log.New(errOut, "", 0),
		debugLog:     log.New(errOut, "", 0),
		errLog:       log.New(errOut, "", 0),
		infoLog:      log.New(infoOut, "", 0),
	}
}

func (l *Logger) SetErrOut(out io.Writer) {
	l.ErrOut = out
	l.errLog.SetOutput(out)
	l.warnLog.SetOutput(out)
	l.debugLog.SetOutput(out)
}

func (l *Logger) SetInfoOut(out io.Writer) {
	l.Out = out
	l.infoLog.SetOutput(out)
}

func (l *Logger) Debug(format string, v ...any) {
	if l.DebugEnabled {
		l.debugLog.Printf("[debug] %s\n", fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Warn(format string, v ...any) {
	if l.QuietEnabled {
		return
	}
	l.warnLog.Printf("[warning] %s\n", fmt.Sprintf(format, v...))
}

func (l *Logger) Error(format string, v ...any) {
	l.errLog.Fatalf("Error: %s\n", fmt.Sprintf(format, v...))
}

func (l *Logger) Info(format string, v ...any) {
	l.infoLog.Printf("%s\n", fmt.Sprintf(format, v...))
}

// Print writes to the info output without appending a trailing newline
// (log.Logger always appends one), e.g. for inline prompts.
func (l *Logger) Print(format string, v ...any) {
	_, _ = fmt.Fprintf(l.infoLog.Writer(), format, v...)
}
