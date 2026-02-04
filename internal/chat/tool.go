package chat

import "encoding/json"

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
	Properties map[string]Property `json:"properties"` // Always required by Anthropic API
	Required   []string            `json:"required,omitempty"`
}

// MarshalJSON ensures Properties is never null (Anthropic API requires it).
func (s Schema) MarshalJSON() ([]byte, error) {
	type schema Schema // avoid recursion
	if s.Properties == nil {
		s.Properties = map[string]Property{}
	}
	return json.Marshal(schema(s))
}

// NewObjectSchema creates an object schema with the given properties.
func NewObjectSchema(properties map[string]Property, required []string) Schema {
	if properties == nil {
		properties = map[string]Property{}
	}
	return Schema{
		Type:       "object",
		Properties: properties,
		Required:   required,
	}
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
