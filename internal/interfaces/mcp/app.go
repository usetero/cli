package mcp

import (
	"fmt"

	"github.com/usetero/cli/internal/infrastructure/logging"
	"github.com/usetero/cli/internal/interfaces/cli/config"
)

// Start runs MCP mode (scaffold).
func Start(_ config.RuntimeConfig, scope logging.Scope) error {
	scope.Info("mcp started")
	fmt.Println("MCP interface scaffold is ready.")
	return nil
}
