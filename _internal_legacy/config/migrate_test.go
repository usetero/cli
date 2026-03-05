package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// setTestHome overrides HOME to a temp directory and returns a cleanup function.
func setTestHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	return tmp
}

// writeYAML writes a YAML file at the given path, creating parent directories.
func writeYAML(t *testing.T, path string, data map[string]any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	out, err := yaml.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, out, 0o600); err != nil {
		t.Fatal(err)
	}
}

// readYAML reads a YAML file into a map.
func readYAML(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

// assertFile checks that a file exists or doesn't exist.
func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected %s to exist", filepath.Base(path))
	}
}

func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected %s to not exist", filepath.Base(path))
	}
}

func TestMigrate_NewUser(t *testing.T) {
	home := setTestHome(t)
	env := "test"
	envDir := filepath.Join(home, ".tero", "environments", env)

	Migrate(env)

	assertFileNotExists(t, filepath.Join(envDir, "preferences.yaml"))
}

func TestMigrate_AlreadyMigrated(t *testing.T) {
	home := setTestHome(t)
	env := "test"
	envDir := filepath.Join(home, ".tero", "environments", env)

	writeYAML(t, filepath.Join(envDir, "preferences.yaml"), map[string]any{
		"active_org_id": "org-existing",
		"role":          "engineer",
	})

	// Stale active.yaml should not be touched.
	writeYAML(t, filepath.Join(envDir, "active.yaml"), map[string]any{
		"org_id": "org-stale",
	})

	Migrate(env)

	userPrefs := readYAML(t, filepath.Join(envDir, "preferences.yaml"))
	if userPrefs["active_org_id"] != "org-existing" {
		t.Errorf("active_org_id = %v, want org-existing", userPrefs["active_org_id"])
	}
	assertFileExists(t, filepath.Join(envDir, "active.yaml"))
}

func TestMigrate_PreOrgLayout(t *testing.T) {
	home := setTestHome(t)
	env := "test"
	envDir := filepath.Join(home, ".tero", "environments", env)
	orgID := "org-123"
	orgDir := filepath.Join(envDir, "orgs", orgID)

	// State B: flat config.yaml with everything in it, databases at env level.
	writeYAML(t, filepath.Join(envDir, "config.yaml"), map[string]any{
		"default_org_id":       orgID,
		"default_account_id":   "acc-456",
		"default_workspace_id": "ws-789",
		"role":                 "platform",
	})

	// Put a database file at env level.
	dbDir := filepath.Join(envDir, "databases")
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dbDir, "test.sqlite"), []byte("db"), 0o600); err != nil {
		t.Fatal(err)
	}

	Migrate(env)

	// User preferences.
	userPrefs := readYAML(t, filepath.Join(envDir, "preferences.yaml"))
	if userPrefs["active_org_id"] != orgID {
		t.Errorf("active_org_id = %v, want %v", userPrefs["active_org_id"], orgID)
	}
	if userPrefs["role"] != "platform" {
		t.Errorf("role = %v, want platform", userPrefs["role"])
	}

	// Org preferences.
	orgPrefs := readYAML(t, filepath.Join(orgDir, "preferences.yaml"))
	if orgPrefs["default_account_id"] != "acc-456" {
		t.Errorf("default_account_id = %v, want acc-456", orgPrefs["default_account_id"])
	}
	if orgPrefs["default_workspace_id"] != "ws-789" {
		t.Errorf("default_workspace_id = %v, want ws-789", orgPrefs["default_workspace_id"])
	}

	// Database moved to org directory.
	assertFileExists(t, filepath.Join(orgDir, "databases", "test.sqlite"))
	assertFileNotExists(t, filepath.Join(envDir, "databases", "test.sqlite"))

	// Old files cleaned up.
	assertFileNotExists(t, filepath.Join(envDir, "config.yaml"))
}

func TestMigrate_PostOrgLayout(t *testing.T) {
	home := setTestHome(t)
	env := "test"
	envDir := filepath.Join(home, ".tero", "environments", env)
	orgID := "org-123"
	orgDir := filepath.Join(envDir, "orgs", orgID)

	// State C: active.yaml + org config.yaml (role ended up in org config).
	writeYAML(t, filepath.Join(envDir, "active.yaml"), map[string]any{
		"org_id": orgID,
	})
	writeYAML(t, filepath.Join(orgDir, "config.yaml"), map[string]any{
		"default_account_id":   "acc-456",
		"default_workspace_id": "ws-789",
		"role":                 "engineer",
	})

	Migrate(env)

	// User preferences — role extracted from org config.
	userPrefs := readYAML(t, filepath.Join(envDir, "preferences.yaml"))
	if userPrefs["active_org_id"] != orgID {
		t.Errorf("active_org_id = %v, want %v", userPrefs["active_org_id"], orgID)
	}
	if userPrefs["role"] != "engineer" {
		t.Errorf("role = %v, want engineer", userPrefs["role"])
	}

	// Org preferences — only org-level keys.
	orgPrefs := readYAML(t, filepath.Join(orgDir, "preferences.yaml"))
	if orgPrefs["default_account_id"] != "acc-456" {
		t.Errorf("default_account_id = %v, want acc-456", orgPrefs["default_account_id"])
	}
	if orgPrefs["default_workspace_id"] != "ws-789" {
		t.Errorf("default_workspace_id = %v, want ws-789", orgPrefs["default_workspace_id"])
	}
	if _, hasRole := orgPrefs["role"]; hasRole {
		t.Error("role should not be in org preferences")
	}

	// Old files cleaned up.
	assertFileNotExists(t, filepath.Join(envDir, "active.yaml"))
	assertFileNotExists(t, filepath.Join(orgDir, "config.yaml"))
}

func TestMigrate_PostOrgLayout_RoleInEnvConfig(t *testing.T) {
	home := setTestHome(t)
	env := "test"
	envDir := filepath.Join(home, ".tero", "environments", env)
	orgDir := filepath.Join(envDir, "orgs", "org-123")

	// active.yaml + env config with role + org config without role.
	writeYAML(t, filepath.Join(envDir, "active.yaml"), map[string]any{
		"org_id": "org-123",
	})
	writeYAML(t, filepath.Join(envDir, "config.yaml"), map[string]any{
		"role": "platform",
	})
	writeYAML(t, filepath.Join(orgDir, "config.yaml"), map[string]any{
		"default_account_id": "acc-456",
	})

	Migrate(env)

	userPrefs := readYAML(t, filepath.Join(envDir, "preferences.yaml"))
	if userPrefs["role"] != "platform" {
		t.Errorf("role = %v, want platform", userPrefs["role"])
	}
}

func TestMigrate_ActiveYamlTakesPrecedence(t *testing.T) {
	home := setTestHome(t)
	env := "test"
	envDir := filepath.Join(home, ".tero", "environments", env)

	// active.yaml and config.yaml both have org IDs — active.yaml wins.
	writeYAML(t, filepath.Join(envDir, "active.yaml"), map[string]any{
		"org_id": "org-from-active",
	})
	writeYAML(t, filepath.Join(envDir, "config.yaml"), map[string]any{
		"default_org_id": "org-stale",
	})
	// orgs/ exists so State B (pre-org) path is skipped.
	if err := os.MkdirAll(filepath.Join(envDir, "orgs", "org-from-active"), 0o755); err != nil {
		t.Fatal(err)
	}

	Migrate(env)

	userPrefs := readYAML(t, filepath.Join(envDir, "preferences.yaml"))
	if userPrefs["active_org_id"] != "org-from-active" {
		t.Errorf("active_org_id = %v, want org-from-active", userPrefs["active_org_id"])
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	home := setTestHome(t)
	env := "test"
	envDir := filepath.Join(home, ".tero", "environments", env)
	orgID := "org-123"
	orgDir := filepath.Join(envDir, "orgs", orgID)

	writeYAML(t, filepath.Join(envDir, "active.yaml"), map[string]any{
		"org_id": orgID,
	})
	writeYAML(t, filepath.Join(orgDir, "config.yaml"), map[string]any{
		"default_account_id": "acc-456",
		"role":               "platform",
	})

	Migrate(env)
	Migrate(env)

	userPrefs := readYAML(t, filepath.Join(envDir, "preferences.yaml"))
	if userPrefs["active_org_id"] != orgID {
		t.Errorf("active_org_id = %v, want %v", userPrefs["active_org_id"], orgID)
	}
	if userPrefs["role"] != "platform" {
		t.Errorf("role = %v, want platform", userPrefs["role"])
	}

	orgPrefs := readYAML(t, filepath.Join(orgDir, "preferences.yaml"))
	if orgPrefs["default_account_id"] != "acc-456" {
		t.Errorf("default_account_id = %v, want acc-456", orgPrefs["default_account_id"])
	}
}
