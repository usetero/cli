package logtest

import (
	"log/slog"
	"testing"

	"github.com/usetero/cli/internal/infrastructure/logging"
)

type testWriter struct {
	t *testing.T
}

func (w *testWriter) Write(p []byte) (n int, err error) {
	w.t.Log(string(p))
	return len(p), nil
}

// New creates a debug logger writing to testing.T.
func New(t *testing.T) logging.Logger {
	return logging.Wrap(slog.New(slog.NewTextHandler(&testWriter{t: t}, &slog.HandlerOptions{Level: slog.LevelDebug})))
}

// NewScope creates a root scope writing to testing.T.
func NewScope(t *testing.T) logging.Scope {
	return logging.RootScope(New(t))
}
