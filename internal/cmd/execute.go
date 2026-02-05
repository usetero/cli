package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/log"
)

// Execute runs the root command
func Execute(version string) {
	// Load config to determine log level
	cliConfig := config.LoadCLIConfig()

	level := log.LevelInfo
	if cliConfig.Debug {
		level = log.LevelDebug
	}

	scope := log.RootScope(log.New(level))

	defer func() {
		if r := recover(); r != nil {
			scope.Error("panic recovered", "panic", r, "stack", string(debug.Stack()))
			fmt.Fprintf(os.Stderr, "Fatal error: %v\n", r)
			os.Exit(1)
		}
	}()

	rootCmd := NewRootCmd(scope, version)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
