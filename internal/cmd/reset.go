package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/keyring"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/workos"
)

func NewResetCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	scope = scope.Child("reset")
	var includeDB bool

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Clear all preferences and authentication",
		Long:  "Remove stored authentication tokens and user preferences for the current environment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			theme := styles.DetectTheme()
			s := theme.Styles
			env := cliConfig.Environment()

			// Clear preferences
			orgID := config.ActiveOrgID(env)
			cfg, err := config.Load(env, orgID)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			prefs := preferences.NewService(cfg, scope)
			if err := prefs.Clear(); err != nil {
				return fmt.Errorf("failed to clear preferences: %w", err)
			}

			// Clear auth tokens
			tokenStore := keyring.New(env)
			workosClient := workos.NewClient(cliConfig.WorkOSClientID, cliConfig.APIEndpoint, cliConfig.PowerSyncEndpoint)
			authService := auth.NewService(workosClient, tokenStore, scope)
			if err := authService.ClearTokens(); err != nil {
				return fmt.Errorf("failed to clear tokens: %w", err)
			}

			// Clear database if requested
			if includeDB {
				storage := sqlite.NewStorageService(cfg)
				if err := storage.Clear(); err != nil {
					return fmt.Errorf("failed to clear database: %w", err)
				}

				fmt.Println(s.Success.Render("✓ Reset complete"))
				fmt.Println(s.Help.Render("Cleared preferences, authentication, and database for: " + env))
			} else {
				fmt.Println(s.Success.Render("✓ Reset complete"))
				fmt.Println(s.Help.Render("Cleared preferences and authentication for: " + env))
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&includeDB, "db", false, "Also delete the local SQLite database")

	return cmd
}
