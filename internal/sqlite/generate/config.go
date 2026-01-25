package main

import "errors"

// Config holds the configuration for the generator.
type Config struct {
	URL   string
	Token string
}

// Validate checks that the configuration is complete.
func (c Config) Validate() error {
	if c.URL == "" {
		return errors.New("POWERSYNC_URL is required")
	}
	if c.Token == "" {
		return errors.New("POWERSYNC_API_TOKEN is required")
	}
	return nil
}
