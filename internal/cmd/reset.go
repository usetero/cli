package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
)

// listOrgIDs returns all org IDs found in the filesystem for the given environment.
func listOrgIDs(env string) ([]domain.OrganizationID, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	orgsDir := filepath.Join(homeDir, ".tero", "environments", env, "orgs")
	entries, err := os.ReadDir(orgsDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var orgIDs []domain.OrganizationID
	for _, entry := range entries {
		if entry.IsDir() {
			orgIDs = append(orgIDs, domain.OrganizationID(entry.Name()))
		}
	}

	return orgIDs, nil
}

func NewResetCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	scope = scope.Child("reset")

	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Clear all preferences and authentication",
		Long:  "Remove stored authentication tokens and user preferences for the current environment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			theme := styles.DetectTheme()
			s := theme.Styles
			env := cliConfig.Environment()

			// Get list of all orgs from filesystem
			orgIDs, err := listOrgIDs(env)
			if err != nil {
				return fmt.Errorf("failed to list orgs: %w", err)
			}

			// Clear org preferences for all orgs
			for _, orgID := range orgIDs {
				orgCfg, err := config.LoadOrgPreferences(env, orgID)
				if err == nil {
					orgPrefs := preferences.NewOrgService(orgCfg, scope)
					if err := orgPrefs.Clear(); err != nil {
						return fmt.Errorf("failed to clear org preferences for %s: %w", orgID, err)
					}
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
			authService := newAuthService(cliConfig, scope)
			if err := authService.ClearTokens(); err != nil {
				return fmt.Errorf("failed to clear tokens: %w", err)
			}

			// Print results
			fmt.Println(s.Success.Render("✓ Reset complete"))
			fmt.Println(s.Help.Render("Cleared preferences and authentication for: " + env))

			return nil
		},
	}

	return cmd
}
