package gentest

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/gobuffalo/genny"
	"github.com/markbates/safe"
)

const (
	DEBUG string = "DEBU"
	INFO         = "INFO"
	WARN         = "WARN"
	ERROR        = "ERRO"
	FATAL        = "FATA"
	PANIC        = "PANI"
	PRINT        = "PRIN"
)

// compile-time assertion to guarantee Logger conforms to genny.Logger
var _ genny.Logger = &Logger{}

// NewLogger produces an initialized Logger
func NewLogger() *Logger {
	l := &Logger{
		Stream: &bytes.Buffer{},
		Log:    map[string][]string{},
		moot:   &sync.Mutex{},
	}
	return l
}

// Logger is a repository of log messages emitted by Generators suitable for
// examination in tests.
type Logger struct {
	// Stream contains the raw byte stream from all logging activity. This is
	// useful for output with testing.(*T).Log in concert with checking for
	// testing.Verbose()
	Stream *bytes.Buffer

	// Log contains all log messages, categorized by log level. The keys of this
	// map are the log levels found in the const definitions.
	Log map[string][]string

	// PrintFn is an optional func which will be invoked with all log messages
	// when they are emitted
	PrintFn func(...interface{})

	// CloseFn is an optional cleanup function invoked at the end of the runner's
	// execution
	CloseFn func() error
	moot    *sync.Mutex
}

// Close is invoked at the end of a genny#Runner's execution. It will invoke a
// CloseFn if set.
func (l *Logger) Close() error {
	if l.CloseFn == nil {
		return nil
	}
	return l.CloseFn()
}

func (l *Logger) logf(lvl string, s string, args ...interface{}) {
	l.log(lvl, fmt.Sprintf(s, args...))
}

func (l *Logger) log(lvl string, args ...interface{}) {
	l.moot.Lock()
	m := l.Log[lvl]
	s := fmt.Sprint(args...)
	m = append(m, s)
	l.Stream.WriteString(fmt.Sprintf("[%s] %s\n", lvl, s))
	l.Log[lvl] = m
	l.moot.Unlock()
	if l.PrintFn != nil {
		safe.Run(func() {
			l.PrintFn(args...)
		})
	}
}

// Debugf processes a format string and produces a debug message to the test
// logger
func (l *Logger) Debugf(s string, args ...interface{}) {
	l.logf(DEBUG, s, args...)
}

// Debug combines a variadic number of strings into a single debug log message
func (l *Logger) Debug(args ...interface{}) {
	l.log(DEBUG, args...)
}

// Infof processes a format string and produces a log message at the Info level
func (l *Logger) Infof(s string, args ...interface{}) {
	l.logf(INFO, s, args...)
}

// Info combines a variadic number of strings into a log message at the Info
// level.
func (l *Logger) Info(args ...interface{}) {
	l.log(INFO, args...)
}

// Printf logs messages at the Print level after processing the string and its
// arguments as a format string
func (l *Logger) Printf(s string, args ...interface{}) {
	l.logf(PRINT, s, args...)
}

// Print logs messages at the Print level after combining all arguments into a
// single string
func (l *Logger) Print(args ...interface{}) {
	l.log(PRINT, args...)
}

// Warnf logs messages at the Warn level after processing the string as a
// format string against the provided arguments
func (l *Logger) Warnf(s string, args ...interface{}) {
	l.logf(WARN, s, args...)
}

// Warn logs messages at the Warn level after combining the provided arguments
// into a single string
func (l *Logger) Warn(args ...interface{}) {
	l.log(WARN, args...)
}

// Errorf logs the provided string at the warning level. Prior to logging, the
// string is processed as a format string against the provided arguments
func (l *Logger) Errorf(s string, args ...interface{}) {
	l.logf(ERROR, s, args...)
}

// Error logs a message formed by combining the arguments into a single string
// at the Error level
func (l *Logger) Error(args ...interface{}) {
	l.log(ERROR, args...)
}

// Fatalf logs the provided string at the Fatal level. Prior to logging, the
// string is processed as a format string against the provided arguments
func (l *Logger) Fatalf(s string, args ...interface{}) {
	l.logf(FATAL, s, args...)
}

// Fatal logs a message formed by combining the arguments into a single string
// at the Fatal level
func (l *Logger) Fatal(args ...interface{}) {
	l.log(FATAL, args...)
}

// Panicf logs the provided string at the Panic level. Prior to logging, the
// string is processed as a format string against the provided arguments
func (l *Logger) Panicf(s string, args ...interface{}) {
	l.logf(PANIC, s, args...)
}

// Panic logs a message formed by combining the arguments into a single string
// at the Panic level
func (l *Logger) Panic(args ...interface{}) {
	l.log(PANIC, args...)
}
