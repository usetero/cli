package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/keyring"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/workos"
)

func NewResetCmd(logger log.Logger, cliConfig *config.CLIConfig) *cobra.Command {
	var includeDB bool

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Clear all preferences and authentication",
		Long:  "Remove stored authentication tokens and user preferences for the current environment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			theme := styles.NewTheme(true)
			s := theme.Styles
			namespace := cliConfig.Namespace()

			// Clear preferences
			cfg, err := config.Load(namespace)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			prefs := preferences.NewService(cfg, logger)
			if err := prefs.Clear(); err != nil {
				return fmt.Errorf("failed to clear preferences: %w", err)
			}

			// Clear auth tokens
			tokenStore := keyring.New(namespace)
			workosClient := workos.NewClient(cliConfig.WorkOSClientID, cliConfig.APIEndpoint, cliConfig.PowerSyncEndpoint)
			authService := auth.NewService(workosClient, tokenStore, logger)
			if err := authService.ClearTokens(); err != nil {
				return fmt.Errorf("failed to clear tokens: %w", err)
			}

			// Clear database if requested
			if includeDB {
				psConfig := &powersync.Config{Namespace: namespace}
				if err := psConfig.Clear(); err != nil {
					return fmt.Errorf("failed to clear database: %w", err)
				}

				fmt.Println(s.Success.Render("Reset complete"))
				fmt.Println(s.Help.Render("  Cleared preferences, authentication, and database for: " + namespace))
			} else {
				fmt.Println(s.Success.Render("Reset complete"))
				fmt.Println(s.Help.Render("  Cleared preferences and authentication for: " + namespace))
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&includeDB, "db", false, "Also delete the local SQLite database")

	return cmd
}
