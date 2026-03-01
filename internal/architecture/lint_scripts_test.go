package architecture_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEventNamingLint(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	script := filepath.Join(root, "scripts", "lint", "check-event-naming.sh")

	t.Run("passes for valid event suffixes", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "events.go"), `package events
type ThemeChangeRequested struct{}
type ErrorToastPublished struct{}
type SyncStateChanged struct{}
`)

		out, err := runScript(t, script, "TERO_EVENT_DIR="+dir)
		if err != nil {
			t.Fatalf("expected success, got error: %v\noutput:\n%s", err, out)
		}
		if !strings.Contains(out, "event naming lint passed") {
			t.Fatalf("expected pass output, got:\n%s", out)
		}
	})

	t.Run("fails for invalid event suffix", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "events.go"), `package events
type SomethingHappened struct{}
`)

		out, err := runScript(t, script, "TERO_EVENT_DIR="+dir)
		if err == nil {
			t.Fatalf("expected failure, got success:\n%s", out)
		}
		if !strings.Contains(out, "event naming lint failed") {
			t.Fatalf("expected failure output, got:\n%s", out)
		}
	})
}

func TestLocalMsgNamingLint(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	script := filepath.Join(root, "scripts", "lint", "check-local-msg-naming.sh")

	t.Run("passes for valid local msg suffixes", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "model.go"), `package sample
type detailLoadedMsg struct{}
type refreshTickMsg struct{}
type authCompletedMsg struct{}
`)

		out, err := runScript(t, script, "TERO_LOCAL_MSG_ROOT="+dir)
		if err != nil {
			t.Fatalf("expected success, got error: %v\noutput:\n%s", err, out)
		}
		if !strings.Contains(out, "local msg naming lint passed") {
			t.Fatalf("expected pass output, got:\n%s", out)
		}
	})

	t.Run("fails for ambiguous local msg name", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "model.go"), `package sample
type detailMsg struct{}
`)

		out, err := runScript(t, script, "TERO_LOCAL_MSG_ROOT="+dir)
		if err == nil {
			t.Fatalf("expected failure, got success:\n%s", out)
		}
		if !strings.Contains(out, "local msg naming lint failed") {
			t.Fatalf("expected failure output, got:\n%s", out)
		}
	})
}

func TestEventLoopSafetyLint(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	script := filepath.Join(root, "scripts", "lint", "check-event-loop-safety.sh")

	t.Run("passes when Update/View have no direct blocking calls", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "model.go"), `package sample
import "context"
func helper() {
	_, _ = context.WithTimeout(context.Background(), 0)
}
type Model struct{}
func (m *Model) Update(_ any) any { return nil }
func (m *Model) View() string { return "" }
`)

		out, err := runScript(t, script, "TERO_EVENT_LOOP_ROOT="+dir)
		if err != nil {
			t.Fatalf("expected success, got error: %v\noutput:\n%s", err, out)
		}
		if !strings.Contains(out, "event-loop safety lint passed") {
			t.Fatalf("expected pass output, got:\n%s", out)
		}
	})

	t.Run("fails when Update performs blocking work", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "model.go"), `package sample
import "context"
type Model struct{}
func (m *Model) Update(_ any) any {
	_, _ = context.WithTimeout(context.Background(), 0)
	return nil
}
`)

		out, err := runScript(t, script, "TERO_EVENT_LOOP_ROOT="+dir)
		if err == nil {
			t.Fatalf("expected failure, got success:\n%s", out)
		}
		if !strings.Contains(out, "event-loop safety lint failed") {
			t.Fatalf("expected failure output, got:\n%s", out)
		}
	})
}

func TestEventOwnershipLint(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	script := filepath.Join(root, "scripts", "lint", "check-event-ownership.sh")

	t.Run("passes for owner and app root imports", func(t *testing.T) {
		t.Parallel()

		scanRoot := t.TempDir()
		writeFile(t, filepath.Join(scanRoot, "internal", "app", "chat", "events", "events.go"), `package events
type ChatUpdated struct{}
`)
		writeFile(t, filepath.Join(scanRoot, "internal", "app", "chat", "use.go"), `package chat
import _ "github.com/usetero/cli/internal/app/chat/events"
`)
		writeFile(t, filepath.Join(scanRoot, "internal", "app", "app.go"), `package app
import _ "github.com/usetero/cli/internal/app/chat/events"
`)

		out, err := runScript(t, script,
			"TERO_SCAN_ROOT="+scanRoot,
			"TERO_MODULE_ROOT="+scanRoot,
			"TERO_APP_ROOT="+filepath.Join(scanRoot, "internal", "app"),
		)
		if err != nil {
			t.Fatalf("expected success, got error: %v\noutput:\n%s", err, out)
		}
		if !strings.Contains(out, "event ownership lint passed") {
			t.Fatalf("expected pass output, got:\n%s", out)
		}
	})

	t.Run("fails for cross-feature import", func(t *testing.T) {
		t.Parallel()

		scanRoot := t.TempDir()
		writeFile(t, filepath.Join(scanRoot, "internal", "app", "chat", "events", "events.go"), `package events
type ChatUpdated struct{}
`)
		writeFile(t, filepath.Join(scanRoot, "internal", "app", "statusbar", "bad.go"), `package statusbar
import _ "github.com/usetero/cli/internal/app/chat/events"
`)

		out, err := runScript(t, script,
			"TERO_SCAN_ROOT="+scanRoot,
			"TERO_MODULE_ROOT="+scanRoot,
			"TERO_APP_ROOT="+filepath.Join(scanRoot, "internal", "app"),
		)
		if err == nil {
			t.Fatalf("expected failure, got success:\n%s", out)
		}
		if !strings.Contains(out, "event ownership lint failed") {
			t.Fatalf("expected failure output, got:\n%s", out)
		}
	})
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	root, err := findModuleRoot(wd)
	if err != nil {
		t.Fatalf("find module root: %v", err)
	}
	return root
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func runScript(t *testing.T, scriptPath string, env ...string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", scriptPath)
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
