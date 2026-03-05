package logging

import (
	"bytes"
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
