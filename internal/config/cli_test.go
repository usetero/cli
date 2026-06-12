package config

import (
	"testing"
)

func TestLoadCLIConfig_DefaultsToPrd(t *testing.T) {
	t.Setenv("TERO_ENV", "")
	t.Setenv("TERO_API_ENDPOINT", "")
	t.Setenv("TERO_POWERSYNC_ENDPOINT", "")
	t.Setenv("TERO_CHAT_ENDPOINT", "")
	t.Setenv("WORKOS_CLIENT_ID", "")
	t.Setenv("TERO_DEBUG", "")

	cfg := LoadCLIConfig()

	if cfg.Env != "prd" {
		t.Errorf("Env = %q, want %q", cfg.Env, "prd")
	}
	if cfg.Environment() != "prd" {
		t.Errorf("Environment() = %q, want %q", cfg.Environment(), "prd")
	}
	if cfg.APIEndpoint != "https://api.usetero.com" {
		t.Errorf("APIEndpoint = %q, want production", cfg.APIEndpoint)
	}
	if cfg.PowerSyncEndpoint != "https://powersync.usetero.com" {
		t.Errorf("PowerSyncEndpoint = %q, want production", cfg.PowerSyncEndpoint)
	}
	if cfg.WorkOSClientID != "client_01JQCC2D06JF9ASFA6GRHMFA3N" {
		t.Errorf("WorkOSClientID = %q, want production client", cfg.WorkOSClientID)
	}
}

func TestLoadCLIConfig_Local(t *testing.T) {
	t.Setenv("TERO_ENV", "local")
	t.Setenv("TERO_API_ENDPOINT", "")
	t.Setenv("TERO_POWERSYNC_ENDPOINT", "")
	t.Setenv("TERO_CHAT_ENDPOINT", "")
	t.Setenv("WORKOS_CLIENT_ID", "")

	cfg := LoadCLIConfig()

	if cfg.Env != "local" {
		t.Errorf("Env = %q, want %q", cfg.Env, "local")
	}
	if cfg.APIEndpoint != "http://localhost:18081" {
		t.Errorf("APIEndpoint = %q, want localhost", cfg.APIEndpoint)
	}
	if cfg.PowerSyncEndpoint != "http://localhost:18084" {
		t.Errorf("PowerSyncEndpoint = %q, want localhost", cfg.PowerSyncEndpoint)
	}
	if cfg.WorkOSClientID != "client_01JQCC2CJMTB8AY2JRMZXFY9R1" {
		t.Errorf("WorkOSClientID = %q, want local/dev client", cfg.WorkOSClientID)
	}
}

func TestLoadCLIConfig_Dev(t *testing.T) {
	t.Setenv("TERO_ENV", "dev")
	t.Setenv("TERO_API_ENDPOINT", "")
	t.Setenv("TERO_POWERSYNC_ENDPOINT", "")
	t.Setenv("TERO_CHAT_ENDPOINT", "")
	t.Setenv("WORKOS_CLIENT_ID", "")

	cfg := LoadCLIConfig()

	if cfg.Env != "dev" {
		t.Errorf("Env = %q, want %q", cfg.Env, "dev")
	}
	if cfg.APIEndpoint != "https://api.usetero.dev" {
		t.Errorf("APIEndpoint = %q, want dev", cfg.APIEndpoint)
	}
	if cfg.PowerSyncEndpoint != "https://powersync.usetero.dev" {
		t.Errorf("PowerSyncEndpoint = %q, want dev", cfg.PowerSyncEndpoint)
	}
}

func TestLoadCLIConfig_Prd(t *testing.T) {
	t.Setenv("TERO_ENV", "prd")
	t.Setenv("TERO_API_ENDPOINT", "")
	t.Setenv("TERO_POWERSYNC_ENDPOINT", "")
	t.Setenv("TERO_CHAT_ENDPOINT", "")
	t.Setenv("WORKOS_CLIENT_ID", "")

	cfg := LoadCLIConfig()

	if cfg.Env != "prd" {
		t.Errorf("Env = %q, want %q", cfg.Env, "prd")
	}
	if cfg.APIEndpoint != "https://api.usetero.com" {
		t.Errorf("APIEndpoint = %q, want production", cfg.APIEndpoint)
	}
}

func TestLoadCLIConfig_UnknownEnvFallsToPrd(t *testing.T) {
	t.Setenv("TERO_ENV", "staging")
	t.Setenv("TERO_API_ENDPOINT", "")
	t.Setenv("TERO_POWERSYNC_ENDPOINT", "")
	t.Setenv("TERO_CHAT_ENDPOINT", "")
	t.Setenv("WORKOS_CLIENT_ID", "")

	cfg := LoadCLIConfig()

	if cfg.Env != "staging" {
		t.Errorf("Env = %q, want %q", cfg.Env, "staging")
	}
	if cfg.Environment() != "staging" {
		t.Errorf("Environment() = %q, want %q", cfg.Environment(), "staging")
	}
	if cfg.APIEndpoint != "https://api.usetero.com" {
		t.Errorf("APIEndpoint = %q, want production fallback", cfg.APIEndpoint)
	}
}

func TestLoadCLIConfig_EnvVarOverridesDefaults(t *testing.T) {
	t.Setenv("TERO_ENV", "dev")
	t.Setenv("TERO_API_ENDPOINT", "http://custom:9999")
	t.Setenv("TERO_POWERSYNC_ENDPOINT", "")
	t.Setenv("TERO_CHAT_ENDPOINT", "")
	t.Setenv("WORKOS_CLIENT_ID", "")

	cfg := LoadCLIConfig()

	if cfg.APIEndpoint != "http://custom:9999" {
		t.Errorf("APIEndpoint = %q, want custom override", cfg.APIEndpoint)
	}
	// Other fields still use dev defaults
	if cfg.PowerSyncEndpoint != "https://powersync.usetero.dev" {
		t.Errorf("PowerSyncEndpoint = %q, want dev default", cfg.PowerSyncEndpoint)
	}
}

func TestLoadCLIConfig_Debug(t *testing.T) {
	t.Setenv("TERO_ENV", "prd")
	t.Setenv("TERO_API_ENDPOINT", "")
	t.Setenv("TERO_POWERSYNC_ENDPOINT", "")
	t.Setenv("TERO_CHAT_ENDPOINT", "")
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
