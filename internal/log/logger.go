package log

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
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
	With(args ...any) Logger
}

// Level re-exports slog.Level for callers.
type Level = slog.Level

// Log levels.
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// logger wraps slog.Logger to implement our Logger interface.
type logger struct {
	*slog.Logger
}

func (l *logger) With(args ...any) Logger {
	return &logger{l.Logger.With(args...)}
}

// Wrap wraps a *slog.Logger to implement the Logger interface.
func Wrap(l *slog.Logger) Logger {
	return &logger{l}
}

// LogPath returns the log file path for the given environment.
func LogPath(env string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".tero", "environments", env, "tero.log"), nil
}

const maxLogSize = 5 * 1024 * 1024 // 5MB

// New creates a new logger that writes to ~/.tero/environments/{env}/tero.log.
// Appends to existing log file, with a clear session separator.
// Truncates the file at session start if it exceeds 5MB.
// Falls back to a no-op logger if the log file cannot be opened.
func New(env string, level Level) Logger {
	logPath, err := LogPath(env)
	if err != nil {
		return &logger{slog.New(slog.NewTextHandler(io.Discard, nil))}
	}

	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return &logger{slog.New(slog.NewTextHandler(io.Discard, nil))}
	}

	// Truncate if over size limit before opening for append.
	if info, err := os.Stat(logPath); err == nil && info.Size() > maxLogSize {
		_ = os.Truncate(logPath, 0)
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return &logger{slog.New(slog.NewTextHandler(io.Discard, nil))}
	}

	// Write session separator (ignore errors - best effort logging)
	_, _ = logFile.WriteString("\n")
	_, _ = logFile.WriteString("================================================================================\n")
	_, _ = logFile.WriteString("SESSION START\n")
	_, _ = logFile.WriteString("================================================================================\n")

	return &logger{slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: level,
	}))}
}
