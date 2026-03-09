package config

import (
	"testing"
)

func TestLoadCLIConfig_DefaultsToPrd(t *testing.T) {
	t.Setenv("TERO_ENV", "")
	t.Setenv("TERO_API_ORIGIN", "")
	t.Setenv("TERO_POWERSYNC_ORIGIN", "")
	t.Setenv("TERO_CHAT_ORIGIN", "")
	t.Setenv("WORKOS_CLIENT_ID", "")
	t.Setenv("TERO_DEBUG", "")

	cfg := LoadCLIConfig()

	if cfg.Env != "prd" {
		t.Errorf("Env = %q, want %q", cfg.Env, "prd")
	}
	if cfg.Environment() != "prd" {
		t.Errorf("Environment() = %q, want %q", cfg.Environment(), "prd")
	}
	if cfg.APIOrigin != "https://api.usetero.com" {
		t.Errorf("APIOrigin = %q, want production", cfg.APIOrigin)
	}
	if cfg.PowerSyncOrigin != "https://powersync.usetero.com" {
		t.Errorf("PowerSyncOrigin = %q, want production", cfg.PowerSyncOrigin)
	}
	if cfg.ChatOrigin != "https://chat.usetero.com" {
		t.Errorf("ChatOrigin = %q, want production", cfg.ChatOrigin)
	}
	if cfg.WorkOSClientID != "client_01JQCC2D06JF9ASFA6GRHMFA3N" {
		t.Errorf("WorkOSClientID = %q, want production client", cfg.WorkOSClientID)
	}
}

func TestLoadCLIConfig_Local(t *testing.T) {
	t.Setenv("TERO_ENV", "local")
	t.Setenv("TERO_API_ORIGIN", "")
	t.Setenv("TERO_POWERSYNC_ORIGIN", "")
	t.Setenv("TERO_CHAT_ORIGIN", "")
	t.Setenv("WORKOS_CLIENT_ID", "")

	cfg := LoadCLIConfig()

	if cfg.Env != "local" {
		t.Errorf("Env = %q, want %q", cfg.Env, "local")
	}
	if cfg.APIOrigin != "http://localhost:18081" {
		t.Errorf("APIOrigin = %q, want localhost", cfg.APIOrigin)
	}
	if cfg.PowerSyncOrigin != "http://localhost:18084" {
		t.Errorf("PowerSyncOrigin = %q, want localhost", cfg.PowerSyncOrigin)
	}
	if cfg.ChatOrigin != "http://localhost:18083" {
		t.Errorf("ChatOrigin = %q, want localhost", cfg.ChatOrigin)
	}
	if cfg.WorkOSClientID != "client_01JQCC2CJMTB8AY2JRMZXFY9R1" {
		t.Errorf("WorkOSClientID = %q, want local/dev client", cfg.WorkOSClientID)
	}
}

func TestLoadCLIConfig_Dev(t *testing.T) {
	t.Setenv("TERO_ENV", "dev")
	t.Setenv("TERO_API_ORIGIN", "")
	t.Setenv("TERO_POWERSYNC_ORIGIN", "")
	t.Setenv("TERO_CHAT_ORIGIN", "")
	t.Setenv("WORKOS_CLIENT_ID", "")

	cfg := LoadCLIConfig()

	if cfg.Env != "dev" {
		t.Errorf("Env = %q, want %q", cfg.Env, "dev")
	}
	if cfg.APIOrigin != "https://api.usetero.dev" {
		t.Errorf("APIOrigin = %q, want dev", cfg.APIOrigin)
	}
	if cfg.PowerSyncOrigin != "https://powersync.usetero.dev" {
		t.Errorf("PowerSyncOrigin = %q, want dev", cfg.PowerSyncOrigin)
	}
	if cfg.ChatOrigin != "https://chat.usetero.dev" {
		t.Errorf("ChatOrigin = %q, want dev", cfg.ChatOrigin)
	}
}

func TestLoadCLIConfig_Prd(t *testing.T) {
	t.Setenv("TERO_ENV", "prd")
	t.Setenv("TERO_API_ORIGIN", "")
	t.Setenv("TERO_POWERSYNC_ORIGIN", "")
	t.Setenv("TERO_CHAT_ORIGIN", "")
	t.Setenv("WORKOS_CLIENT_ID", "")

	cfg := LoadCLIConfig()

	if cfg.Env != "prd" {
		t.Errorf("Env = %q, want %q", cfg.Env, "prd")
	}
	if cfg.APIOrigin != "https://api.usetero.com" {
		t.Errorf("APIOrigin = %q, want production", cfg.APIOrigin)
	}
}

func TestLoadCLIConfig_UnknownEnvFallsToPrd(t *testing.T) {
	t.Setenv("TERO_ENV", "staging")
	t.Setenv("TERO_API_ORIGIN", "")
	t.Setenv("TERO_POWERSYNC_ORIGIN", "")
	t.Setenv("TERO_CHAT_ORIGIN", "")
	t.Setenv("WORKOS_CLIENT_ID", "")

	cfg := LoadCLIConfig()

	if cfg.Env != "staging" {
		t.Errorf("Env = %q, want %q", cfg.Env, "staging")
	}
	if cfg.Environment() != "staging" {
		t.Errorf("Environment() = %q, want %q", cfg.Environment(), "staging")
	}
	if cfg.APIOrigin != "https://api.usetero.com" {
		t.Errorf("APIOrigin = %q, want production fallback", cfg.APIOrigin)
	}
}

func TestLoadCLIConfig_EnvVarOverridesDefaults(t *testing.T) {
	t.Setenv("TERO_ENV", "dev")
	t.Setenv("TERO_API_ORIGIN", "http://custom:9999")
	t.Setenv("TERO_POWERSYNC_ORIGIN", "")
	t.Setenv("TERO_CHAT_ORIGIN", "")
	t.Setenv("WORKOS_CLIENT_ID", "")

	cfg := LoadCLIConfig()

	if cfg.APIOrigin != "http://custom:9999" {
		t.Errorf("APIOrigin = %q, want custom override", cfg.APIOrigin)
	}
	// Other fields still use dev defaults
	if cfg.PowerSyncOrigin != "https://powersync.usetero.dev" {
		t.Errorf("PowerSyncOrigin = %q, want dev default", cfg.PowerSyncOrigin)
	}
}

func TestLoadCLIConfig_Debug(t *testing.T) {
	t.Setenv("TERO_ENV", "prd")
	t.Setenv("TERO_API_ORIGIN", "")
	t.Setenv("TERO_POWERSYNC_ORIGIN", "")
	t.Setenv("TERO_CHAT_ORIGIN", "")
	t.Setenv("WORKOS_CLIENT_ID", "")

	t.Setenv("TERO_DEBUG", "true")
	cfg := LoadCLIConfig()
	if !cfg.Debug {
		t.Error("Debug should be true when TERO_DEBUG=true")
	}

	t.Setenv("TERO_DEBUG", "1")
	cfg = LoadCLIConfig()
	if !cfg.Debug {
		t.Error("Debug should be true when TERO_DEBUG=1")
	}

	t.Setenv("TERO_DEBUG", "")
	cfg = LoadCLIConfig()
	if cfg.Debug {
		t.Error("Debug should be false when TERO_DEBUG is empty")
	}
}
