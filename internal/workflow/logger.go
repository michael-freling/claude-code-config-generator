package workflow

import "fmt"

// LogLevel represents the verbosity level for logging
type LogLevel int

const (
	// LogLevelNormal is the default log level showing only important messages
	LogLevelNormal LogLevel = iota
	// LogLevelVerbose enables verbose output with additional details
	LogLevelVerbose
	// LogLevelDebug enables debug output (for future use)
	LogLevelDebug
)

// Logger provides structured logging with different verbosity levels
type Logger interface {
	// Info outputs important messages that are always shown
	Info(format string, args ...interface{})
	// Verbose outputs detailed messages only when LogLevel >= LogLevelVerbose
	Verbose(format string, args ...interface{})
	// Debug outputs debug messages only when LogLevel >= LogLevelDebug
	Debug(format string, args ...interface{})
	// IsVerbose returns true if verbose mode is enabled
	IsVerbose() bool
}

// defaultLogger implements Logger with thread-safe output.
// Note: fmt.Printf is safe for concurrent use as it synchronizes writes to stdout.
type defaultLogger struct {
	level LogLevel
}

// NewLogger creates a new Logger with the specified log level
func NewLogger(level LogLevel) Logger {
	return &defaultLogger{
		level: level,
	}
}

// Info outputs important messages that are always shown
func (l *defaultLogger) Info(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// Verbose outputs detailed messages only when LogLevel >= LogLevelVerbose
func (l *defaultLogger) Verbose(format string, args ...interface{}) {
	if l.level >= LogLevelVerbose {
		fmt.Printf("%s %s\n", Cyan("→"), fmt.Sprintf(format, args...))
	}
}

// Debug outputs debug messages only when LogLevel >= LogLevelDebug
func (l *defaultLogger) Debug(format string, args ...interface{}) {
	if l.level >= LogLevelDebug {
		fmt.Printf("%s %s\n", Yellow("[DEBUG]"), fmt.Sprintf(format, args...))
	}
}

// IsVerbose returns true if verbose mode is enabled
func (l *defaultLogger) IsVerbose() bool {
	return l.level >= LogLevelVerbose
}
