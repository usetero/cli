package logtest

import (
	"log/slog"
	"testing"

	"github.com/usetero/cli/internal/log"
)

// testWriter wraps testing.T to implement io.Writer
type testWriter struct {
	t *testing.T
}

func (w *testWriter) Write(p []byte) (n int, err error) {
	w.t.Log(string(p))
	return len(p), nil
}

// New creates a logger that writes to testing.T
// Logs are only shown when the test fails or with -v flag
func New(t *testing.T) log.Logger {
	return log.Wrap(slog.New(slog.NewTextHandler(&testWriter{t: t}, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))
}

// NewScope creates a root scope that writes to testing.T
// Logs are only shown when the test fails or with -v flag
func NewScope(t *testing.T) log.Scope {
	return log.RootScope(New(t))
}
