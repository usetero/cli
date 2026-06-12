package cmd

import (
	"github.com/spf13/cobra"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/log"
)

func NewInternalCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	scope = scope.Child("internal")

	internalCmd := &cobra.Command{
		Use:    "internal",
		Short:  "Internal developer and operations commands",
		Long:   "Internal commands for diagnostics and operational workflows.",
		Hidden: true,
	}

	internalCmd.AddCommand(NewInternalInspectCmd(scope, cliConfig))

	return internalCmd
}
