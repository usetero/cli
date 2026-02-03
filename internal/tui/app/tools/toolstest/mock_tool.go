// Package toolstest provides test doubles for the tools package.
package toolstest

import (
	"encoding/json"

	"github.com/usetero/cli/internal/chat"
)

// MockTool is a configurable tool for testing.
type MockTool struct {
	NameVal   string
	ExecuteFn func(input json.RawMessage) (any, error)
}

func (m MockTool) Definition() chat.Tool {
	return chat.Tool{Name: m.NameVal}
}

func (m MockTool) Execute(input json.RawMessage) (any, error) {
	if m.ExecuteFn != nil {
		return m.ExecuteFn(input)
	}
	return map[string]string{"status": "ok"}, nil
}
