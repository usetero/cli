package config

// ChatConfig configures the chat service origin.
type ChatConfig struct {
	Origin string `name:"chat-origin" help:"Chat API origin. Do not include a path." env:"TERO_CHAT_ORIGIN"`
}
