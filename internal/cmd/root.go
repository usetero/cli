package cmd

import (
	"context"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/app"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/keyring"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/filter"
	"github.com/usetero/cli/internal/workos"
)

func NewRootCmd(scope log.Scope, version string) *cobra.Command {
	// Load CLI configuration
	cliConfig := config.LoadCLIConfig()

	rootCmd := newRootCmd(scope, version, cliConfig)

	// Subcommands
	rootCmd.AddCommand(NewAuthCmd(scope, cliConfig))
	rootCmd.AddCommand(NewResetCmd(scope, cliConfig))
	rootCmd.AddCommand(NewDebugCmd(scope, cliConfig))

	return rootCmd
}

func newRootCmd(scope log.Scope, version string, cliConfig *config.CLIConfig) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "tero",
		Short:   "Tero - Your telemetry quality platform",
		Version: version,
		Long: `Tero is a telemetry quality platform that helps you understand and improve
your observability data across all your tools.

Just run 'tero' to start an interactive chat session.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			env := cliConfig.Environment()

			// Create context
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Load user preferences (env-level)
			userCfg, err := config.LoadUserPreferences(env)
			if err != nil {
				return err
			}
			userPrefs := preferences.NewUserService(userCfg, scope)

			// Resolve theme from user preference
			theme := resolveTheme(userPrefs.GetTheme())

			// Load org preferences (org-level)
			orgID := userPrefs.GetActiveOrgID()
			orgCfg, err := config.LoadOrgPreferences(env, orgID)
			if err != nil {
				return err
			}
			orgPrefs := preferences.NewOrgService(orgCfg, scope)

			// Create token store
			tokenStore := keyring.New(env)

			// Create WorkOS client for OAuth
			workosClient := workos.NewClient(cliConfig.WorkOSClientID, cliConfig.APIEndpoint, cliConfig.PowerSyncEndpoint, cliConfig.ChatEndpoint)

			// Create auth service
			authService := auth.NewService(workosClient, tokenStore, scope)

			// Create API services
			services := api.NewServices(cliConfig.APIEndpoint+"/graphql", authService, scope)

			// Create storage for SQLite databases
			storage := sqlite.NewStorageService(orgCfg)

			// Create PowerSync syncer
			syncer := powersync.NewSyncer(cliConfig.PowerSyncEndpoint, authService, scope)

			// Create and run the TUI
			p := tea.NewProgram(
				app.New(ctx, cliConfig, theme, version, services, authService, userPrefs, orgPrefs, storage, syncer, scope),
				tea.WithContext(ctx),
				tea.WithEnvironment(os.Environ()),
				tea.WithFilter(filter.Mouse),
			)
			if _, err := p.Run(); err != nil {
				scope.Error("bubbletea program error", "error", err)
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

// resolveTheme creates a theme from the user's preference.
func resolveTheme(pref preferences.Theme) styles.Theme {
	switch pref {
	case preferences.ThemeDark:
		return styles.NewTheme(true)
	case preferences.ThemeLight:
		return styles.NewTheme(false)
	default:
		return styles.DetectTheme()
	}
}
