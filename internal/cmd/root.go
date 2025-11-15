package cmd

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/spf13/cobra"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui"
)

func NewRootCmd(logger log.Logger, version string) *cobra.Command {
	// Load CLI configuration (env vars + smart defaults)
	cliConfig := config.LoadCLIConfig(version)

	rootCmd := &cobra.Command{
		Use:     "tero",
		Short:   "Tero - Your telemetry quality platform",
		Version: version,
		Long: `Tero is a telemetry quality platform that helps you understand and improve
your observability data across all your tools.

Just run 'tero' to start an interactive chat session.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load user preferences
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			// Get endpoint from flag (allows override of env var/default)
			endpoint, _ := cmd.Flags().GetString("endpoint")

			// Create and run the TUI
			p := tea.NewProgram(tui.New(cfg, endpoint, logger))
			if _, err := p.Run(); err != nil {
				logger.Error("bubbletea program error", "error", err)
				return err
			}
			return nil
		},
	}

	// Global flags with defaults from CLI config
	rootCmd.PersistentFlags().String("endpoint", cliConfig.APIEndpoint, "Tero control plane endpoint")
	rootCmd.PersistentFlags().BoolP("debug", "d", cliConfig.Debug, "Enable debug logging")

	// Subcommands (add later)
	// rootCmd.AddCommand(NewMCPCmd())

	return rootCmd
}
