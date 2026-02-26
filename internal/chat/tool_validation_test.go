package chat

import "testing"

func TestValidateTools(t *testing.T) {
	t.Parallel()

	valid := []Tool{{
		Name:        "query",
		Description: "Run SQL",
		InputSchema: NewObjectSchema(map[string]Property{"sql": {Type: "string"}}, []string{"sql"}),
	}}
	if err := validateTools(valid); err != nil {
		t.Fatalf("validateTools(valid) error = %v", err)
	}

	tests := []struct {
		name  string
		tools []Tool
	}{
		{name: "missing name", tools: []Tool{{Description: "d", InputSchema: NewObjectSchema(map[string]Property{}, nil)}}},
		{name: "duplicate name", tools: []Tool{{Name: "q", Description: "d", InputSchema: NewObjectSchema(map[string]Property{}, nil)}, {Name: "q", Description: "d2", InputSchema: NewObjectSchema(map[string]Property{}, nil)}}},
		{name: "missing description", tools: []Tool{{Name: "q", InputSchema: NewObjectSchema(map[string]Property{}, nil)}}},
		{name: "wrong schema type", tools: []Tool{{Name: "q", Description: "d", InputSchema: Schema{Type: "string", Properties: map[string]Property{}}}}},
		{name: "nil properties", tools: []Tool{{Name: "q", Description: "d", InputSchema: Schema{Type: "object"}}}},
		{name: "empty property name", tools: []Tool{{Name: "q", Description: "d", InputSchema: NewObjectSchema(map[string]Property{"": {Type: "string"}}, nil)}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := validateTools(tt.tools); err == nil {
				t.Fatalf("validateTools(%s): expected error", tt.name)
			}
		})
	}
}
