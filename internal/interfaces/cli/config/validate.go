package config

import (
	"fmt"
	"net/url"
)

func Validate(cfg RuntimeConfig) error {
	if cfg.Env == "" {
		return fmt.Errorf("environment is required")
	}
	if err := validateURL("api", cfg.API.URL); err != nil {
		return err
	}
	if err := validateURL("chat", cfg.Chat.URL); err != nil {
		return err
	}
	if err := validateURL("powersync", cfg.PowerSync.URL); err != nil {
		return err
	}
	if cfg.WorkOS.ClientID == "" {
		return fmt.Errorf("workos client id is required")
	}
	if !cfg.Logging.Level.Valid() {
		return fmt.Errorf("invalid log level: %q", cfg.Logging.Level)
	}
	return nil
}

func validateURL(name, raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("invalid %s url: %q", name, raw)
	}
	return nil
}
