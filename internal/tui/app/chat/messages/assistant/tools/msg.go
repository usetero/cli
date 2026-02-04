package tools

import "github.com/usetero/cli/internal/domain"

// ResultMsg is emitted by a tool model when execution completes.
// Chat catches this to track pending tools.
type ResultMsg struct {
	ToolUseID string
	Result    *domain.ToolResult
}
