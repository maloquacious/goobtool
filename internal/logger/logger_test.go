package logger

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestStdLogger(t *testing.T) {
	var buf bytes.Buffer
	l := &StdLogger{
		logger: log.New(&buf, "", 0),
	}

	tests := []struct {
		name     string
		fn       func()
		expected string
	}{
		{
			name:     "Info",
			fn:       func() { l.Info("test message") },
			expected: "[INFO] test message",
		},
		{
			name:     "Warn",
			fn:       func() { l.Warn("warning message") },
			expected: "[WARN] warning message",
		},
		{
			name:     "Error",
			fn:       func() { l.Error("error message") },
			expected: "[ERROR] error message",
		},
		{
			name:     "Debug",
			fn:       func() { l.Debug("debug message") },
			expected: "[DEBUG] debug message",
		},
		{
			name:     "Info with args",
			fn:       func() { l.Info("test %s=%d", "count", 42) },
			expected: "[INFO] test count=42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.fn()
			got := strings.TrimSpace(buf.String())
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestDefault(t *testing.T) {
	if Default == nil {
		t.Error("Default logger should not be nil")
	}

	Default.Info("test")
}
