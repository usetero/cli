package powersync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/usetero/cli/internal/log/logtest"
)

func TestNDJSONStreamCapture_WritesLines(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "capture.ndjson")
	capture, err := NewNDJSONStreamCapture(path, 1024, logtest.NewScope(t))
	if err != nil {
		t.Fatalf("NewNDJSONStreamCapture() error = %v", err)
	}
	defer func() { _ = capture.Close() }()

	capture.CaptureLine([]byte(`{"op":"put"}`))
	capture.CaptureLine([]byte(`{"op":"remove"}`))

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	got := string(data)
	want := "{\"op\":\"put\"}\n{\"op\":\"remove\"}\n"
	if got != want {
		t.Fatalf("capture file = %q, want %q", got, want)
	}
}

func TestNDJSONStreamCapture_StopsAtMaxBytes(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "capture.ndjson")
	capture, err := NewNDJSONStreamCapture(path, 8, logtest.NewScope(t))
	if err != nil {
		t.Fatalf("NewNDJSONStreamCapture() error = %v", err)
	}
	defer func() { _ = capture.Close() }()

	capture.CaptureLine([]byte("abc")) // 4 bytes with newline
	capture.CaptureLine([]byte("def")) // 4 bytes with newline
	capture.CaptureLine([]byte("ghi")) // should be ignored

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	got := string(data)
	want := "abc\ndef\n"
	if got != want {
		t.Fatalf("capture file = %q, want %q", got, want)
	}
}
