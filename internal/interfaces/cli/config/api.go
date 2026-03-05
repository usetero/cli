package config

// APIConfig configures the control-plane API endpoint.
type APIConfig struct {
	URL string `name:"api-url" help:"Control-plane API URL." env:"TERO_API_URL"`
}
