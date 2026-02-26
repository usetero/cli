package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeLine_DeterministicAndSafe(t *testing.T) {
	s := newSanitizer()

	line := `{"checkpoint":{"last_op_id":"42","buckets":[{"bucket":"29#account_data[\"acc-123\"]"}]},"data":{"op":"PUT","object_id":"user@example.com","data":"{\"account_id\":\"acc-123\",\"url\":\"https://api.usetero.com/v1\"}"}}`
	got1, err := s.sanitizeLine(line)
	if err != nil {
		t.Fatalf("sanitizeLine() error = %v", err)
	}
	got2, err := s.sanitizeLine(line)
	if err != nil {
		t.Fatalf("sanitizeLine() second error = %v", err)
	}
	if got1 != got2 {
		t.Fatalf("sanitize output must be deterministic\ngot1=%s\ngot2=%s", got1, got2)
	}

	if !strings.Contains(got1, `"last_op_id":"42"`) {
		t.Fatalf("expected preserved numeric checkpoint field, got: %s", got1)
	}
	if strings.Contains(got1, "user@example.com") {
		t.Fatalf("expected sensitive value to be redacted, got: %s", got1)
	}
	if strings.Contains(got1, "api.usetero.com") {
		t.Fatalf("expected nested JSON string content to be redacted, got: %s", got1)
	}
}

func TestSanitizeFixtureFile_MaxLines(t *testing.T) {
	in := filepath.Join(t.TempDir(), "in.ndjson")
	out := filepath.Join(t.TempDir(), "out.ndjson")

	content := strings.Join([]string{
		`{"token_expires_in":3600}`,
		`{"checkpoint":{"last_op_id":"0","buckets":[]}}`,
		`{"checkpoint":{"last_op_id":"1","buckets":[]}}`,
	}, "\n") + "\n"
	if err := os.WriteFile(in, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	n, err := sanitizeFixtureFile(in, out, 2)
	if err != nil {
		t.Fatalf("sanitizeFixtureFile() error = %v", err)
	}
	if n != 2 {
		t.Fatalf("line count = %d, want 2", n)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("output lines = %d, want 2", len(lines))
	}
}

func TestSanitizeLine_DropsNonReplayMessages(t *testing.T) {
	s := newSanitizer()
	got, err := s.sanitizeLine(`{"token_expires_in":3600}`)
	if err != nil {
		t.Fatalf("sanitizeLine() error = %v", err)
	}
	if got != "" {
		t.Fatalf("expected dropped line, got %q", got)
	}
}

func TestSanitizeLine_ArrayAndNestedJSONDeterministic(t *testing.T) {
	s := newSanitizer()

	line := `{"data":{"op":"PUT","object_id":"user-123","tags":["env:prod","env:prod","service:payments"],"payload":"{\"owner\":\"user-123\",\"emails\":[\"a@example.com\",\"a@example.com\"],\"tokens\":[\"tok_abcdef\",\"tok_abcdef\"]}"}}`
	got, err := s.sanitizeLine(line)
	if err != nil {
		t.Fatalf("sanitizeLine() error = %v", err)
	}

	// Repeated sensitive values should map to the same token.
	if strings.Count(got, "redacted_") < 3 {
		t.Fatalf("expected multiple redactions, got: %s", got)
	}
	if strings.Contains(got, "a@example.com") || strings.Contains(got, "tok_abcdef") || strings.Contains(got, "user-123") {
		t.Fatalf("expected sensitive values redacted, got: %s", got)
	}

	got2, err := s.sanitizeLine(line)
	if err != nil {
		t.Fatalf("sanitizeLine() second error = %v", err)
	}
	if got != got2 {
		t.Fatalf("non-deterministic output\ngot=%s\ngot2=%s", got, got2)
	}
}
