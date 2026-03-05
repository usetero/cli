package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// Logger is the logging contract used by the application.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
}

// Level controls logger verbosity.
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Valid reports whether l is a known level.
func (l Level) Valid() bool {
	switch l {
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
		return true
	default:
		return false
	}
}

func (l Level) toSlogLevel() slog.Level {
	switch l {
	case LevelDebug:
		return slog.LevelDebug
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

const maxLogSize = 5 * 1024 * 1024 // 5MB

type logger struct {
	*slog.Logger
}

// Wrap adapts slog.Logger to the local Logger interface.
func Wrap(l *slog.Logger) Logger {
	return &logger{Logger: l}
}

func (l *logger) With(args ...any) Logger {
	return &logger{Logger: l.Logger.With(args...)}
}

// NewWithWriter creates a logger writing to the given writer.
func NewWithWriter(w io.Writer, level Level) Logger {
	if w == nil {
		w = io.Discard
	}
	return Wrap(slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level.toSlogLevel()})))
}

// LogPath returns the log file path for the given environment.
func LogPath(env string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".tero", "environments", env, "tero.log"), nil
}

// New creates a file-backed logger under ~/.tero/environments/{env}/tero.log.
// If the logger cannot be initialized, it falls back to a discard logger.
func New(env string, level Level) Logger {
	logPath, err := LogPath(env)
	if err != nil {
		return NewWithWriter(io.Discard, level)
	}

	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return NewWithWriter(io.Discard, level)
	}

	if info, err := os.Stat(logPath); err == nil && info.Size() > maxLogSize {
		_ = os.Truncate(logPath, 0)
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return NewWithWriter(io.Discard, level)
	}

	_, _ = logFile.WriteString("\n")
	_, _ = logFile.WriteString("================================================================================\n")
	_, _ = logFile.WriteString("SESSION START\n")
	_, _ = logFile.WriteString("================================================================================\n")

	return NewWithWriter(logFile, level)
}
