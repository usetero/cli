package config

// RuntimeConfig is the typed configuration injected into app modes.
type RuntimeConfig struct {
	Env       EnvironmentProfile `name:"env" help:"Runtime environment profile." enum:"local,dev,prd" env:"TERO_ENV" default:"dev"`
	API       APIConfig          `embed:""`
	Chat      ChatConfig         `embed:""`
	PowerSync PowerSyncConfig    `embed:""`
	WorkOS    WorkOSConfig       `embed:""`
	Theme     ThemeConfig        `embed:""`
	Logging   LoggingConfig      `embed:""`
}

// Resolve builds runtime config by applying explicit inputs over profile defaults.
func Resolve(in RuntimeConfig) (RuntimeConfig, error) {
	env := in.Env
	if env == "" {
		env = ProfileDev
	}

	defaults, err := profileByName(env)
	if err != nil {
		return RuntimeConfig{}, err
	}

	cfg := RuntimeConfig{
		Env:       env,
		API:       APIConfig{Origin: defaults.APIOrigin},
		Chat:      ChatConfig{Origin: defaults.ChatOrigin},
		PowerSync: PowerSyncConfig{Origin: defaults.PowerSyncOrigin},
		WorkOS:    WorkOSConfig{ClientID: defaults.WorkOSClientID},
		Theme:     ThemeConfig{Mode: ThemeModeAuto},
		Logging:   LoggingConfig{Level: defaults.LogLevel},
	}

	if in.API.Origin != "" {
		cfg.API.Origin = in.API.Origin
	}
	if in.Chat.Origin != "" {
		cfg.Chat.Origin = in.Chat.Origin
	}
	if in.PowerSync.Origin != "" {
		cfg.PowerSync.Origin = in.PowerSync.Origin
	}
	if in.WorkOS.ClientID != "" {
		cfg.WorkOS.ClientID = in.WorkOS.ClientID
	}
	if in.Theme.Mode != "" {
		cfg.Theme.Mode = in.Theme.Mode
	}
	if in.Logging.Level != "" {
		cfg.Logging.Level = in.Logging.Level
	}

	if err := Validate(cfg); err != nil {
		return RuntimeConfig{}, err
	}
	return cfg, nil
}
