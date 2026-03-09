package config

import (
	"fmt"
	"net/url"
	"strings"
)

func Validate(cfg RuntimeConfig) error {
	if cfg.Env == "" {
		return fmt.Errorf("environment is required")
	}
	if err := validateOrigin("api", cfg.API.Origin); err != nil {
		return err
	}
	if err := validateOrigin("chat", cfg.Chat.Origin); err != nil {
		return err
	}
	if err := validateOrigin("powersync", cfg.PowerSync.Origin); err != nil {
		return err
	}
	if cfg.WorkOS.ClientID == "" {
		return fmt.Errorf("workos client id is required")
	}
	if !cfg.Theme.Mode.Valid() {
		return fmt.Errorf("invalid theme mode: %q", cfg.Theme.Mode)
	}
	if !cfg.Logging.Level.Valid() {
		return fmt.Errorf("invalid log level: %q", cfg.Logging.Level)
	}
	return nil
}

func validateOrigin(name, raw string) error {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("invalid %s origin: %q", name, raw)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid %s origin scheme: %q", name, raw)
	}
	if u.Path != "" && u.Path != "/" {
		return fmt.Errorf("invalid %s origin: %q must not include a path", name, raw)
	}
	if u.RawQuery != "" {
		return fmt.Errorf("invalid %s origin: %q must not include a query", name, raw)
	}
	if u.Fragment != "" {
		return fmt.Errorf("invalid %s origin: %q must not include a fragment", name, raw)
	}
	return nil
}
