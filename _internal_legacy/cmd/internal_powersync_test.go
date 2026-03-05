package cmd

import (
	"path/filepath"
	"testing"
)

func TestResolveCaptureOutputPath_Absolute(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	abs := "/tmp/capture.ndjson"

	got, err := resolveCaptureOutputPath("prd", abs)
	if err != nil {
		t.Fatalf("resolveCaptureOutputPath() error = %v", err)
	}
	if got != abs {
		t.Fatalf("path = %q, want %q", got, abs)
	}
}

func TestResolveCaptureOutputPath_Relative(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got, err := resolveCaptureOutputPath("dev", filepath.Join("fixtures", "capture.ndjson"))
	if err != nil {
		t.Fatalf("resolveCaptureOutputPath() error = %v", err)
	}

	want := filepath.Join(home, ".tero", "environments", "dev", "fixtures", "capture.ndjson")
	if got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}
