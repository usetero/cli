package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestScopeChildPath(t *testing.T) {
	s := RootScope(NewWithWriter(nil, LevelDebug))
	s1 := s.Child("app")
	s2 := s1.Child("onboarding")

	if got := s1.Path(); got != "app" {
		t.Fatalf("unexpected path: %q", got)
	}
	if got := s2.Path(); got != "app/onboarding" {
		t.Fatalf("unexpected nested path: %q", got)
	}
}

func TestWithPreservesPath(t *testing.T) {
	s := RootScope(NewWithWriter(nil, LevelDebug)).Child("app")
	s2 := s.With("k", "v")

	if s2.Path() != s.Path() {
		t.Fatalf("with should preserve scope path: got %q want %q", s2.Path(), s.Path())
	}
}

func TestChildDoesNotDuplicateScopeField(t *testing.T) {
	var buf bytes.Buffer
	scope := RootScope(NewWithWriter(&buf, LevelInfo)).
		Child("app").
		Child("chat")

	scope.Info("started")

	out := buf.String()
	if strings.Count(out, "scope=") != 1 {
		t.Fatalf("expected one scope field, got output: %q", out)
	}
	if !strings.Contains(out, "scope=app/chat") {
		t.Fatalf("expected nested scope path, got output: %q", out)
	}
}

func TestWithContextSurvivesChild(t *testing.T) {
	var buf bytes.Buffer
	scope := RootScope(NewWithWriter(&buf, LevelInfo)).
		With("request_id", "r1").
		Child("runtime")

	scope.Info("event")

	out := buf.String()
	if !strings.Contains(out, "request_id=r1") {
		t.Fatalf("expected request_id context in child scope output: %q", out)
	}
	if !strings.Contains(out, "scope=runtime") {
		t.Fatalf("expected scope field in child output: %q", out)
	}
}
