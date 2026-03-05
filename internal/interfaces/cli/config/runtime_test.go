package config

import "testing"

func TestResolve_UsesProfileDefaults(t *testing.T) {
	cfg, err := Resolve(RuntimeConfig{Env: ProfileDev})
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if cfg.API.URL != "https://api.usetero.dev" {
		t.Fatalf("unexpected api url: %s", cfg.API.URL)
	}
}

func TestResolve_OverridesProfileValues(t *testing.T) {
	cfg, err := Resolve(RuntimeConfig{Env: ProfileDev, API: APIConfig{URL: "https://api.override"}})
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if cfg.API.URL != "https://api.override" {
		t.Fatalf("expected override api url, got: %s", cfg.API.URL)
	}
}
