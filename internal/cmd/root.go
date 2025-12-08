package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/spf13/cobra"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/keyring"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/tui"
	"github.com/usetero/cli/internal/workos"
)

func NewRootCmd(logger log.Logger, version string) *cobra.Command {
	// Load CLI configuration
	cliConfig := config.LoadCLIConfig()

	rootCmd := &cobra.Command{
		Use:     "tero",
		Short:   "Tero - Your telemetry quality platform",
		Version: version,
		Long: `Tero is a telemetry quality platform that helps you understand and improve
your observability data across all your tools.

Just run 'tero' to start an interactive chat session.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace := cliConfig.Namespace()

			// Load user preferences (namespaced by environment)
			cfg, err := config.Load(namespace)
			if err != nil {
				return err
			}

			// Create token store (namespaced by environment)
			tokenStore := keyring.New(namespace)

			// Handle --reset flag
			reset, _ := cmd.Flags().GetBool("reset")
			if reset {
				if err := cfg.Clear(); err != nil {
					return fmt.Errorf("failed to clear preferences: %w", err)
				}
				if err := tokenStore.Delete("access_token"); err != nil {
					return fmt.Errorf("failed to clear access token: %w", err)
				}
				if err := tokenStore.Delete("refresh_token"); err != nil {
					return fmt.Errorf("failed to clear refresh token: %w", err)
				}

				logger.Info("reset complete", "namespace", namespace)

				// Reload config after clearing
				cfg, err = config.Load(namespace)
				if err != nil {
					return err
				}
			}

			// Get endpoint from flag (allows override of env var/default)
			endpoint, _ := cmd.Flags().GetString("endpoint")

			// Create WorkOS client for authentication
			workosClient := workos.NewClient(cliConfig.WorkOSClientID)

			// Create and run the TUI
			p := tea.NewProgram(tui.New(cfg, tokenStore, workosClient, endpoint, logger))
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
	rootCmd.Flags().Bool("reset", false, "Clear all preferences and authentication")

	// Subcommands (add later)
	// rootCmd.AddCommand(NewMCPCmd())

	return rootCmd
}
