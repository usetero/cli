package chat

import "fmt"

func validateTools(tools []Tool) error {
	seen := make(map[string]struct{}, len(tools))
	for i, tool := range tools {
		if tool.Name == "" {
			return fmt.Errorf("tools[%d]: name is required", i)
		}
		if _, exists := seen[tool.Name]; exists {
			return fmt.Errorf("tools[%d]: duplicate tool name %q", i, tool.Name)
		}
		seen[tool.Name] = struct{}{}

		if tool.Description == "" {
			return fmt.Errorf("tools[%d]: description is required", i)
		}
		if tool.InputSchema.Type != "object" {
			return fmt.Errorf("tools[%d]: input_schema.type must be object", i)
		}
		if tool.InputSchema.Properties == nil {
			return fmt.Errorf("tools[%d]: input_schema.properties is required", i)
		}
		for key := range tool.InputSchema.Properties {
			if key == "" {
				return fmt.Errorf("tools[%d]: property name must not be empty", i)
			}
		}
	}
	return nil
}
