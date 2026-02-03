package chat

// Tool defines a tool the AI can call.
// This is the wire format sent to the Chat API.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema Schema `json:"input_schema"`
}

// Schema defines the JSON Schema for tool input.
type Schema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

// Property defines a single property in a JSON Schema.
type Property struct {
	Type        string   `json:"type,omitempty"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Items       *Items   `json:"items,omitempty"`
}

// Items defines the schema for array items.
type Items struct {
	Type  string `json:"type,omitempty"`
	Items *Items `json:"items,omitempty"`
}
