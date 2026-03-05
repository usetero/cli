package cli

import mcpiface "github.com/usetero/cli/internal/interfaces/mcp"

// MCP contains command-specific options for MCP mode.
type MCP struct{}

// Run starts MCP mode.
func (m *MCP) Run(exec *runner) error {
	return mcpiface.Start(exec.cfg, exec.scope.Child("mcp"))
}
