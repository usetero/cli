package terotest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	domainprefs "github.com/usetero/cli/internal/domains/preferences"
)

func SeedPreferences(t testing.TB, homeDir, env string, snapshot domainprefs.Snapshot) {
	t.Helper()

	path := filepath.Join(homeDir, ".tero", "environments", env, "preferences.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create preferences dir: %v", err)
	}
	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		t.Fatalf("marshal preferences: %v", err)
	}
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatalf("seed preferences: %v", err)
	}
}
