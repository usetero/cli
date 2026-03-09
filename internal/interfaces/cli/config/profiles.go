package config

import (
	"fmt"

	"github.com/usetero/cli/internal/infrastructure/logging"
)

// EnvironmentProfile identifies a named environment default set.
type EnvironmentProfile string

const (
	ProfileLocal EnvironmentProfile = "local"
	ProfileDev   EnvironmentProfile = "dev"
	ProfilePrd   EnvironmentProfile = "prd"
)

type profileDefaults struct {
	APIOrigin       string
	ChatOrigin      string
	PowerSyncOrigin string
	WorkOSClientID  string
	LogLevel        logging.Level
}

var profiles = map[EnvironmentProfile]profileDefaults{
	ProfileLocal: {
		APIOrigin:       "http://localhost:18081",
		ChatOrigin:      "http://localhost:18083",
		PowerSyncOrigin: "http://localhost:18084",
		WorkOSClientID:  "client_01JQCC2CJMTB8AY2JRMZXFY9R1",
		LogLevel:        logging.LevelDebug,
	},
	ProfileDev: {
		APIOrigin:       "https://api.usetero.dev",
		ChatOrigin:      "https://chat.usetero.dev",
		PowerSyncOrigin: "https://powersync.usetero.dev",
		WorkOSClientID:  "client_01JQCC2CJMTB8AY2JRMZXFY9R1",
		LogLevel:        logging.LevelInfo,
	},
	ProfilePrd: {
		APIOrigin:       "https://api.usetero.com",
		ChatOrigin:      "https://chat.usetero.com",
		PowerSyncOrigin: "https://powersync.usetero.com",
		WorkOSClientID:  "client_01JQCC2D06JF9ASFA6GRHMFA3N",
		LogLevel:        logging.LevelInfo,
	},
}

func profileByName(name EnvironmentProfile) (profileDefaults, error) {
	p, ok := profiles[name]
	if !ok {
		return profileDefaults{}, fmt.Errorf("invalid environment profile: %q", name)
	}
	return p, nil
}
