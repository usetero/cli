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

			// Clear org preferences
			orgID := config.ActiveOrgID(env)
			if orgID != "" {
				orgCfg, err := config.LoadOrgPreferences(env, orgID)
				if err != nil {
					return fmt.Errorf("failed to load org preferences: %w", err)
				}
				orgPrefs := preferences.NewOrgService(orgCfg, scope)
				if err := orgPrefs.Clear(); err != nil {
					return fmt.Errorf("failed to clear org preferences: %w", err)
				}
			}

			// Clear user preferences
			userCfg, err := config.LoadUserPreferences(env)
			if err != nil {
				return fmt.Errorf("failed to load user preferences: %w", err)
			}
			userPrefs := preferences.NewUserService(userCfg, scope)
			if err := userPrefs.Clear(); err != nil {
				return fmt.Errorf("failed to clear user preferences: %w", err)
			}

			// Clear auth tokens
			tokenStore := keyring.New(env)
			workosClient := workos.NewClient(cliConfig.WorkOSClientID, cliConfig.APIEndpoint, cliConfig.PowerSyncEndpoint)
			authService := auth.NewService(workosClient, tokenStore, scope)
			if err := authService.ClearTokens(); err != nil {
				return fmt.Errorf("failed to clear tokens: %w", err)
			}

			// Clear database if requested
			if includeDB && orgID != "" {
				orgCfg, err := config.Load(env, orgID)
				if err == nil {
					storage := sqlite.NewStorageService(orgCfg)
					if err := storage.Clear(); err != nil {
						return fmt.Errorf("failed to clear database: %w", err)
					}
				}

				fmt.Println(s.Success.Render("✓ Reset complete"))
				fmt.Println(s.Help.Render("Cleared preferences, authentication, and database for: " + env))
			} else if includeDB {
				fmt.Println(s.Success.Render("✓ Reset complete"))
				fmt.Println(s.Help.Render("Cleared preferences and authentication for: " + env))
				fmt.Println(s.Help.Render("No active org — skipped database"))
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
