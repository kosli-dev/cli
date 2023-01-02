package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

type Logger struct {
	DebugEnabled bool
	Out          io.Writer
	warnLog      *log.Logger
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
		warnLog:      log.New(errOut, "", 0),
		errLog:       log.New(errOut, "", 0),
		infoLog:      log.New(infoOut, "", 0),
	}
}

func (l *Logger) SetErrOut(out io.Writer) {
	l.errLog.SetOutput(out)
}

func (l *Logger) SetInfoOut(out io.Writer) {
	l.infoLog.SetOutput(out)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.DebugEnabled {
		format = fmt.Sprintf("[debug] %s\n", format)
		err := l.infoLog.Output(2, fmt.Sprintf(format, v...))
		if err != nil {
			l.Error(err.Error())
		}
	}
}

func (l *Logger) Warning(format string, v ...interface{}) {
	format = fmt.Sprintf("[warning] %s\n", format)
	l.warnLog.Printf(format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	format = fmt.Sprintf("Error: %s\n", format)
	l.errLog.Fatalf(format, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
	format = fmt.Sprintf("%s\n", format)
	l.infoLog.Printf(format, v...)
}
