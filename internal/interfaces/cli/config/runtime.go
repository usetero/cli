package config

// RuntimeConfig is the typed configuration injected into app modes.
type RuntimeConfig struct {
	Env       EnvironmentProfile `name:"env" help:"Runtime environment profile." enum:"local,dev,prd" env:"TERO_ENV" default:"dev"`
	API       APIConfig          `embed:""`
	Chat      ChatConfig         `embed:""`
	PowerSync PowerSyncConfig    `embed:""`
	WorkOS    WorkOSConfig       `embed:""`
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
		API:       APIConfig{URL: defaults.APIURL},
		Chat:      ChatConfig{URL: defaults.ChatURL},
		PowerSync: PowerSyncConfig{URL: defaults.PowerSyncURL},
		WorkOS:    WorkOSConfig{ClientID: defaults.WorkOSClientID},
		Logging:   LoggingConfig{Level: defaults.LogLevel},
	}

	if in.API.URL != "" {
		cfg.API.URL = in.API.URL
	}
	if in.Chat.URL != "" {
		cfg.Chat.URL = in.Chat.URL
	}
	if in.PowerSync.URL != "" {
		cfg.PowerSync.URL = in.PowerSync.URL
	}
	if in.WorkOS.ClientID != "" {
		cfg.WorkOS.ClientID = in.WorkOS.ClientID
	}
	if in.Logging.Level != "" {
		cfg.Logging.Level = in.Logging.Level
	}

	if err := Validate(cfg); err != nil {
		return RuntimeConfig{}, err
	}
	return cfg, nil
}
