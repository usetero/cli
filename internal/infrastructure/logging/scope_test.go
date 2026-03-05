package logging

import "testing"

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
