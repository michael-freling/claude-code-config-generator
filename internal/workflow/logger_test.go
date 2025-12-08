package workflow

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name  string
		level LogLevel
	}{
		{
			name:  "creates logger with normal level",
			level: LogLevelNormal,
		},
		{
			name:  "creates logger with verbose level",
			level: LogLevelVerbose,
		},
		{
			name:  "creates logger with debug level",
			level: LogLevelDebug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.level)
			assert.NotNil(t, logger)
		})
	}
}

func TestLogger_Info(t *testing.T) {
	tests := []struct {
		name   string
		level  LogLevel
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "outputs info message at normal level",
			level:  LogLevelNormal,
			format: "test message",
			args:   nil,
			want:   "test message\n",
		},
		{
			name:   "outputs info message at verbose level",
			level:  LogLevelVerbose,
			format: "test message",
			args:   nil,
			want:   "test message\n",
		},
		{
			name:   "outputs info message at debug level",
			level:  LogLevelDebug,
			format: "test message",
			args:   nil,
			want:   "test message\n",
		},
		{
			name:   "formats info message with arguments",
			level:  LogLevelNormal,
			format: "test %s with %d",
			args:   []interface{}{"message", 42},
			want:   "test message with 42\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				logger := NewLogger(tt.level)
				logger.Info(tt.format, tt.args...)
			})

			assert.Equal(t, tt.want, output)
		})
	}
}

func TestLogger_Verbose(t *testing.T) {
	tests := []struct {
		name         string
		level        LogLevel
		format       string
		args         []interface{}
		expectOutput bool
		wantContains []string
	}{
		{
			name:         "does not output at normal level",
			level:        LogLevelNormal,
			format:       "verbose message",
			args:         nil,
			expectOutput: false,
		},
		{
			name:         "outputs at verbose level",
			level:        LogLevelVerbose,
			format:       "verbose message",
			args:         nil,
			expectOutput: true,
			wantContains: []string{"→", "verbose message"},
		},
		{
			name:         "outputs at debug level",
			level:        LogLevelDebug,
			format:       "verbose message",
			args:         nil,
			expectOutput: true,
			wantContains: []string{"→", "verbose message"},
		},
		{
			name:         "formats verbose message with arguments",
			level:        LogLevelVerbose,
			format:       "verbose %s with %d",
			args:         []interface{}{"message", 42},
			expectOutput: true,
			wantContains: []string{"→", "verbose message with 42"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				logger := NewLogger(tt.level)
				logger.Verbose(tt.format, tt.args...)
			})

			if tt.expectOutput {
				for _, want := range tt.wantContains {
					assert.Contains(t, output, want)
				}
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestLogger_Debug(t *testing.T) {
	tests := []struct {
		name         string
		level        LogLevel
		format       string
		args         []interface{}
		expectOutput bool
		wantContains []string
	}{
		{
			name:         "does not output at normal level",
			level:        LogLevelNormal,
			format:       "debug message",
			args:         nil,
			expectOutput: false,
		},
		{
			name:         "does not output at verbose level",
			level:        LogLevelVerbose,
			format:       "debug message",
			args:         nil,
			expectOutput: false,
		},
		{
			name:         "outputs at debug level",
			level:        LogLevelDebug,
			format:       "debug message",
			args:         nil,
			expectOutput: true,
			wantContains: []string{"[DEBUG]", "debug message"},
		},
		{
			name:         "formats debug message with arguments",
			level:        LogLevelDebug,
			format:       "debug %s with %d",
			args:         []interface{}{"message", 42},
			expectOutput: true,
			wantContains: []string{"[DEBUG]", "debug message with 42"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				logger := NewLogger(tt.level)
				logger.Debug(tt.format, tt.args...)
			})

			if tt.expectOutput {
				for _, want := range tt.wantContains {
					assert.Contains(t, output, want)
				}
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestLogger_IsVerbose(t *testing.T) {
	tests := []struct {
		name  string
		level LogLevel
		want  bool
	}{
		{
			name:  "returns false at normal level",
			level: LogLevelNormal,
			want:  false,
		},
		{
			name:  "returns true at verbose level",
			level: LogLevelVerbose,
			want:  true,
		},
		{
			name:  "returns true at debug level",
			level: LogLevelDebug,
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.level)
			got := logger.IsVerbose()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLogger_OutputBehavior(t *testing.T) {
	tests := []struct {
		name           string
		level          LogLevel
		infoOutputs    bool
		verboseOutputs bool
		debugOutputs   bool
	}{
		{
			name:           "at normal level: only info outputs",
			level:          LogLevelNormal,
			infoOutputs:    true,
			verboseOutputs: false,
			debugOutputs:   false,
		},
		{
			name:           "at verbose level: info and verbose output",
			level:          LogLevelVerbose,
			infoOutputs:    true,
			verboseOutputs: true,
			debugOutputs:   false,
		},
		{
			name:           "at debug level: all output",
			level:          LogLevelDebug,
			infoOutputs:    true,
			verboseOutputs: true,
			debugOutputs:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.level)

			infoOutput := captureOutput(func() {
				logger.Info("info message")
			})
			if tt.infoOutputs {
				assert.NotEmpty(t, infoOutput, "Info should produce output")
				assert.Contains(t, infoOutput, "info message")
			} else {
				assert.Empty(t, infoOutput, "Info should not produce output")
			}

			verboseOutput := captureOutput(func() {
				logger.Verbose("verbose message")
			})
			if tt.verboseOutputs {
				assert.NotEmpty(t, verboseOutput, "Verbose should produce output")
				assert.Contains(t, verboseOutput, "verbose message")
			} else {
				assert.Empty(t, verboseOutput, "Verbose should not produce output")
			}

			debugOutput := captureOutput(func() {
				logger.Debug("debug message")
			})
			if tt.debugOutputs {
				assert.NotEmpty(t, debugOutput, "Debug should produce output")
				assert.Contains(t, debugOutput, "debug message")
			} else {
				assert.Empty(t, debugOutput, "Debug should not produce output")
			}
		})
	}
}

// captureOutput captures stdout during the execution of a function
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Error ignored in test helper - io.Copy from pipe is reliable
	return buf.String()
}
