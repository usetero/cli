package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/usetero/cli/internal/log"
)

// Execute runs the root command
func Execute() {
	// Create logger once at the top level
	logger := log.New()

	defer func() {
		if r := recover(); r != nil {
			logger.Error("panic recovered", "panic", r, "stack", string(debug.Stack()))
			fmt.Fprintf(os.Stderr, "Fatal error: %v\n", r)
			os.Exit(1)
		}
	}()

	rootCmd := NewRootCmd(logger)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
