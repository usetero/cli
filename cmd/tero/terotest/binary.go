package terotest

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func Build(t testing.TB) string {
	t.Helper()

	outputPath := filepath.Join(t.TempDir(), "tero")
	cmd := exec.CommandContext(context.Background(), "go", "build", "-o", outputPath, ".")
	cmd.Dir = cmdDir(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build tero binary: %v\n%s", err, string(output))
	}
	return outputPath
}

func cmdDir(t testing.TB) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve cmd/tero directory")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
}
