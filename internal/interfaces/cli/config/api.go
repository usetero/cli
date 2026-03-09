package config

// APIConfig configures the control-plane service origin.
type APIConfig struct {
	Origin string `name:"api-origin" help:"Control-plane API origin. Do not include /graphql." env:"TERO_API_ORIGIN"`
}
