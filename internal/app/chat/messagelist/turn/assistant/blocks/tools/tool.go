package tools

import (
	"github.com/usetero/cli/internal/app/chat/messagelist/turn/assistant/blocks"
)

// Model is the interface for tool view models.
// Each tool type has its own implementation.
// Tools fire their own specific completion messages (e.g., msgs.QueryCompleted).
type Model interface {
	blocks.Block
	ToolID() string
	Name() string
	State() State
}

// State represents the current state of tool execution.
type State int

const (
	StateAccumulating State = iota
	StateExecuting
	StateComplete
)
