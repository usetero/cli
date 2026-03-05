package config

// ChatConfig configures the chat API endpoint.
type ChatConfig struct {
	URL string `name:"chat-url" help:"Chat API URL." env:"TERO_CHAT_URL"`
}
