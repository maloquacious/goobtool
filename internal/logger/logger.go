package logger

import (
	"log"
	"os"
)

// Logger defines the Goob logging contract.
// Implementations should support standard log levels and be safe for concurrent use.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
}

// StdLogger wraps Go's standard logger to implement the Goob logging contract.
type StdLogger struct {
	logger *log.Logger
}

// NewStdLogger creates a new StdLogger using Go's standard logger.
func NewStdLogger() *StdLogger {
	return &StdLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (l *StdLogger) Info(msg string, args ...any) {
	l.logger.Printf("[INFO] "+msg, args...)
}

func (l *StdLogger) Warn(msg string, args ...any) {
	l.logger.Printf("[WARN] "+msg, args...)
}

func (l *StdLogger) Error(msg string, args ...any) {
	l.logger.Printf("[ERROR] "+msg, args...)
}

func (l *StdLogger) Debug(msg string, args ...any) {
	l.logger.Printf("[DEBUG] "+msg, args...)
}

// Default provides a global default logger instance using Go's standard logger.
var Default Logger = NewStdLogger()
