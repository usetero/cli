package log

import (
	"log/slog"
	"os"
)

// Re-export slog attribute constructors so callers don't need to import log/slog
var (
	String   = slog.String
	Int      = slog.Int
	Int64    = slog.Int64
	Uint64   = slog.Uint64
	Float64  = slog.Float64
	Bool     = slog.Bool
	Time     = slog.Time
	Duration = slog.Duration
	Group    = slog.Group
	Any      = slog.Any
)

// Logger is the interface for logging throughout the application
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// New creates a new logger that writes to /tmp/tero.log
// Returns *slogger.Logger which implements Logger interface
func New() Logger {
	logFile, err := os.OpenFile("/tmp/tero.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}

	return slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}
