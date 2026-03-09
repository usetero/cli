package config

import (
	"strings"
	"testing"
)

func TestResolve_UsesProfileDefaults(t *testing.T) {
	cfg, err := Resolve(RuntimeConfig{Env: ProfileDev})
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if cfg.API.Origin != "https://api.usetero.dev" {
		t.Fatalf("unexpected api origin: %s", cfg.API.Origin)
	}
}

func TestResolve_OverridesProfileValues(t *testing.T) {
	cfg, err := Resolve(RuntimeConfig{Env: ProfileDev, API: APIConfig{Origin: "https://api.override"}})
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if cfg.API.Origin != "https://api.override" {
		t.Fatalf("expected override api origin, got: %s", cfg.API.Origin)
	}
}

func TestResolve_LocalProfileUsesServiceOrigin(t *testing.T) {
	cfg, err := Resolve(RuntimeConfig{Env: ProfileLocal})
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if cfg.API.Origin != "http://localhost:18081" {
		t.Fatalf("unexpected local api origin: %s", cfg.API.Origin)
	}
	if cfg.Theme.Mode != ThemeModeAuto {
		t.Fatalf("expected auto theme mode, got %q", cfg.Theme.Mode)
	}
}

func TestResolve_OverridesThemeMode(t *testing.T) {
	cfg, err := Resolve(RuntimeConfig{Env: ProfileDev, Theme: ThemeConfig{Mode: ThemeModeDark}})
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if cfg.Theme.Mode != ThemeModeDark {
		t.Fatalf("expected dark theme mode, got %q", cfg.Theme.Mode)
	}
}

func TestResolve_RejectsOriginsWithPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  RuntimeConfig
	}{
		{
			name: "api graphql endpoint",
			cfg:  RuntimeConfig{Env: ProfileDev, API: APIConfig{Origin: "https://api.usetero.dev/graphql"}},
		},
		{
			name: "chat path",
			cfg:  RuntimeConfig{Env: ProfileDev, Chat: ChatConfig{Origin: "https://chat.usetero.dev/api/chat/v2/messages"}},
		},
		{
			name: "powersync path",
			cfg:  RuntimeConfig{Env: ProfileDev, PowerSync: PowerSyncConfig{Origin: "https://powersync.usetero.dev/sync/stream"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Resolve(tt.cfg)
			if err == nil || !strings.Contains(err.Error(), "must not include a path") {
				t.Fatalf("expected path validation error, got %v", err)
			}
		})
	}
}
