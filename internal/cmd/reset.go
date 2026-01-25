package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/keyring"
	"github.com/usetero/cli/internal/styles"
)

func NewResetCmd(cliConfig *config.CLIConfig) *cobra.Command {
	return &cobra.Command{
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
			if err := cfg.Clear(); err != nil {
				return fmt.Errorf("failed to clear preferences: %w", err)
			}

			// Clear tokens
			tokenStore := keyring.New(namespace)
			if err := tokenStore.Delete("access_token"); err != nil {
				return fmt.Errorf("failed to clear access token: %w", err)
			}
			if err := tokenStore.Delete("refresh_token"); err != nil {
				return fmt.Errorf("failed to clear refresh token: %w", err)
			}

			fmt.Println(s.Success.Render("✓ Reset complete"))
			fmt.Println(s.Help.Render("  Cleared preferences and authentication for: " + namespace))
			return nil
		},
	}
}
