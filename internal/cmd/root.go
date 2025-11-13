package cmd

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/spf13/cobra"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui"
)

func NewRootCmd(logger log.Logger) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "tero",
		Short: "Tero - Your telemetry quality platform",
		Long: `Tero is a telemetry quality platform that helps you understand and improve
your observability data across all your tools.

Just run 'tero' to start an interactive chat session.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			// Create and run the TUI (TUI will handle routing based on config)
			p := tea.NewProgram(tui.New(cfg, logger))
			if _, err := p.Run(); err != nil {
				logger.Error("bubbletea program error", "error", err)
				return err
			}
			return nil
		},
	}

	// Global flags
	rootCmd.PersistentFlags().String("endpoint", "http://localhost:8081/graphql", "Tero service endpoint")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug logging")

	// Subcommands (add later)
	// rootCmd.AddCommand(NewMCPCmd())

	return rootCmd
}
