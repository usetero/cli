package cmd

import (
	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/keyring"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/tui"
	"github.com/usetero/cli/internal/workos"
)

func NewRootCmd(logger log.Logger, version string) *cobra.Command {
	// Load CLI configuration
	cliConfig := config.LoadCLIConfig()

	rootCmd := newRootCmd(logger, version, cliConfig)

	// Subcommands
	rootCmd.AddCommand(NewAuthCmd(logger, cliConfig))
	rootCmd.AddCommand(NewResetCmd(logger, cliConfig))
	rootCmd.AddCommand(NewDebugCmd(logger, cliConfig))

	return rootCmd
}

func newRootCmd(logger log.Logger, version string, cliConfig *config.CLIConfig) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "tero",
		Short:   "Tero - Your telemetry quality platform",
		Version: version,
		Long: `Tero is a telemetry quality platform that helps you understand and improve
your observability data across all your tools.

Just run 'tero' to start an interactive chat session.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace := cliConfig.Namespace()

			// Load user preferences
			cfg, err := config.Load(namespace)
			if err != nil {
				return err
			}
			prefs := preferences.NewService(cfg, logger)

			// Create token store
			tokenStore := keyring.New(namespace)

			// Create WorkOS client for OAuth
			workosClient := workos.NewClient(cliConfig.WorkOSClientID, cliConfig.APIEndpoint, cliConfig.PowerSyncEndpoint)

			// Create auth service
			authService := auth.NewService(workosClient, tokenStore, logger)

			// Create storage service for database management
			storage := sqlite.NewStorageService(cfg)

			// Create PowerSync syncer
			syncer := powersync.NewSyncer(cliConfig.PowerSyncEndpoint, authService, logger)

			// Create chat client
			chatClient := chat.NewClient(cliConfig.ChatEndpoint, authService, logger)

			// Create and run the TUI
			p := tea.NewProgram(
				tui.New(cliConfig, authService, prefs, storage, syncer, chatClient, logger),
				tea.WithFilter(tui.MouseEventFilter),
			)
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

	return rootCmd
}
