package logging

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScopedStructuredLog(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithWriter(&buf, LevelDebug)

	scope := RootScope(logger).Child("app").Child("chat")
	scope.Info("started", "conversation_id", "c1")

	out := buf.String()
	if !strings.Contains(out, "scope=app/chat") {
		t.Fatalf("expected scope field in output, got %q", out)
	}
	if !strings.Contains(out, "conversation_id=c1") {
		t.Fatalf("expected structured field in output, got %q", out)
	}
	if !strings.Contains(out, "msg=started") {
		t.Fatalf("expected message in output, got %q", out)
	}
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithWriter(&buf, LevelInfo)

	logger.Debug("debug-msg")
	logger.Info("info-msg")

	out := buf.String()
	if strings.Contains(out, "debug-msg") {
		t.Fatalf("debug log should be filtered at info level: %q", out)
	}
	if !strings.Contains(out, "info-msg") {
		t.Fatalf("info log should be present: %q", out)
	}
}

func TestNewCreatesLogFileInEnvironmentDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	logger := New("dev", LevelInfo)
	logger.Info("hello")

	path := filepath.Join(home, ".tero", "environments", "dev", "tero.log")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected log file to be created: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "SESSION START") {
		t.Fatalf("expected session marker in log file, got %q", text)
	}
	if !strings.Contains(text, "hello") {
		t.Fatalf("expected logged message in file, got %q", text)
	}
}
