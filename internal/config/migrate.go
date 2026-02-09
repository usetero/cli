package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Target layout:
//
//	~/.tero/environments/{env}/preferences.yaml           user prefs (active_org_id, role, theme)
//	~/.tero/environments/{env}/orgs/{id}/preferences.yaml org prefs (default_account_id, default_workspace_id)
//	~/.tero/environments/{env}/orgs/{id}/databases/       SQLite files

// Migrate brings any prior disk layout to the target layout. Idempotent.
func Migrate(env string) {
	dir, err := envDir(env)
	if err != nil {
		return
	}

	// Already at target layout.
	if fileExists(filepath.Join(dir, "preferences.yaml")) {
		return
	}

	// Detect state and migrate directly to target.
	switch {
	case hasV1Layout(env):
		migrateFromV1(env, dir)
	case hasV2Layout(dir):
		migrateFromV2(dir)
	case hasV3Layout(dir):
		migrateFromV3(dir)
	}
	// New user: nothing exists, nothing to do.
}

// Layout detection.
//
// V1: ~/.tero/config.yaml (prd) or ~/.tero/{host}/config.yaml
// V2: ~/.tero/environments/{env}/config.yaml (flat, no orgs/ dir)
// V3: ~/.tero/environments/{env}/active.yaml + orgs/{id}/config.yaml

func hasV1Layout(env string) bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	teroDir := filepath.Join(homeDir, ".tero")
	if env == "prd" {
		return fileExists(filepath.Join(teroDir, "config.yaml"))
	}
	return fileExists(filepath.Join(teroDir, env, "config.yaml"))
}

func hasV2Layout(dir string) bool {
	return fileExists(filepath.Join(dir, "config.yaml")) && !fileExists(filepath.Join(dir, "orgs"))
}

func hasV3Layout(dir string) bool {
	return fileExists(filepath.Join(dir, "active.yaml")) || fileExists(filepath.Join(dir, "orgs"))
}

// V1 → target.
//
// Old: ~/.tero/config.yaml + ~/.tero/data/*.sqlite
// Does everything: moves files into env dir, splits into org dir, writes preferences.
func migrateFromV1(env, dir string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	teroDir := filepath.Join(homeDir, ".tero")

	var oldDir string
	if env == "prd" {
		oldDir = teroDir
	} else {
		oldDir = filepath.Join(teroDir, env)
	}

	// Read the old config.
	cfgData := readYAMLFile(filepath.Join(oldDir, "config.yaml"))
	if cfgData == nil {
		return
	}

	// Move config.yaml into env dir so we can reuse V2 logic.
	_ = os.MkdirAll(dir, 0o755)
	_ = os.Rename(filepath.Join(oldDir, "config.yaml"), filepath.Join(dir, "config.yaml"))

	// Move databases (old: data/, new: databases/).
	moveDBFiles(filepath.Join(oldDir, "data"), filepath.Join(dir, "databases"))

	// Clean up stale V1 files.
	_ = os.Remove(filepath.Join(teroDir, "tero.db"))
	if env != "prd" {
		_ = os.Remove(oldDir)
	}

	// Now we're in V2 state — finish the migration.
	migrateFromV2(dir)
}

// V2 → target.
//
// Old: environments/{env}/config.yaml (flat, all keys in one file, databases/ at env level)
func migrateFromV2(dir string) {
	envConfigPath := filepath.Join(dir, "config.yaml")
	cfgData := readYAMLFile(envConfigPath)
	if cfgData == nil {
		return
	}

	userPrefs := make(map[string]any)

	// Extract user-level keys.
	if role, ok := cfgData["role"].(string); ok && role != "" {
		userPrefs["role"] = role
	}

	orgID, _ := cfgData["default_org_id"].(string)
	if orgID != "" {
		userPrefs["active_org_id"] = orgID

		// Move databases/ → orgs/{id}/databases/
		orgDir := filepath.Join(dir, "orgs", orgID)
		_ = os.MkdirAll(orgDir, 0o755)
		moveDBFiles(filepath.Join(dir, "databases"), filepath.Join(orgDir, "databases"))

		// Write org preferences.
		writeOrgPrefs(cfgData, filepath.Join(orgDir, "preferences.yaml"))
	}

	// Write user preferences.
	writeYAMLFile(filepath.Join(dir, "preferences.yaml"), userPrefs)

	// Clean up.
	_ = os.Remove(envConfigPath)
}

// V3 → target.
//
// Old: active.yaml + orgs/{id}/config.yaml (role may be in org config)
func migrateFromV3(dir string) {
	activePath := filepath.Join(dir, "active.yaml")
	envConfigPath := filepath.Join(dir, "config.yaml")

	userPrefs := make(map[string]any)

	// Read active org ID.
	if data := readYAMLFile(activePath); data != nil {
		if orgID, ok := data["org_id"].(string); ok && orgID != "" {
			userPrefs["active_org_id"] = orgID
		}
	}

	// Read stale env-level config (may have role).
	if data := readYAMLFile(envConfigPath); data != nil {
		if role, ok := data["role"].(string); ok && role != "" {
			userPrefs["role"] = role
		}
	}

	// Convert each org's config.yaml → preferences.yaml.
	orgsDir := filepath.Join(dir, "orgs")
	entries, _ := os.ReadDir(orgsDir)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		orgDir := filepath.Join(orgsDir, entry.Name())
		orgConfigPath := filepath.Join(orgDir, "config.yaml")
		orgPrefsPath := filepath.Join(orgDir, "preferences.yaml")

		if fileExists(orgPrefsPath) {
			continue
		}

		cfgData := readYAMLFile(orgConfigPath)
		if cfgData == nil {
			continue
		}

		// Role may have ended up in org config (V2 moved the whole file).
		if _, hasRole := userPrefs["role"]; !hasRole {
			if role, ok := cfgData["role"].(string); ok && role != "" {
				userPrefs["role"] = role
			}
		}

		writeOrgPrefs(cfgData, orgPrefsPath)
		_ = os.Remove(orgConfigPath)
	}

	// Write user preferences.
	writeYAMLFile(filepath.Join(dir, "preferences.yaml"), userPrefs)

	// Clean up.
	_ = os.Remove(activePath)
	_ = os.Remove(envConfigPath)
}

// Helpers.

func writeOrgPrefs(cfgData map[string]any, path string) {
	orgPrefs := make(map[string]any)
	for _, key := range []string{"default_account_id", "default_workspace_id"} {
		if v, ok := cfgData[key]; ok {
			orgPrefs[key] = v
		}
	}
	writeYAMLFile(path, orgPrefs)
}

func moveDBFiles(src, dst string) {
	files, _ := filepath.Glob(filepath.Join(src, "*.sqlite"))
	if len(files) == 0 {
		return
	}
	_ = os.MkdirAll(dst, 0o755)
	for _, f := range files {
		_ = os.Rename(f, filepath.Join(dst, filepath.Base(f)))
	}
	_ = os.Remove(src)
}

func readYAMLFile(path string) map[string]any {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var result map[string]any
	if yaml.Unmarshal(data, &result) != nil {
		return nil
	}
	return result
}

func writeYAMLFile(path string, data map[string]any) {
	out, err := yaml.Marshal(data)
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, out, 0o600)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
