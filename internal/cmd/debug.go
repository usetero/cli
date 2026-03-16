package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
)

func NewInternalInspectCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	scope = scope.Child("inspect")

	inspectCmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect local and remote state",
		Long:  "Read-only diagnostics for local preferences, paths, and API state.",
	}

	inspectCmd.AddCommand(newDebugStatusCmd(scope, cliConfig))
	inspectCmd.AddCommand(newDebugPrefsCmd(scope, cliConfig))
	inspectCmd.AddCommand(newDebugGraphQLCmd(scope, cliConfig))
	inspectCmd.AddCommand(newDebugPathsCmd(scope, cliConfig))

	return inspectCmd
}

// newDebugStatusCmd shows the current datadog account status
func newDebugStatusCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show Datadog account discovery status",
		Long:  "Show the current discovery status for the Datadog account, including service counts and progress.",
		RunE: func(cmd *cobra.Command, args []string) error {
			theme := styles.DetectTheme()
			s := theme.Styles
			env := cliConfig.Environment()

			// Load org preferences to get account ID
			orgID := config.ActiveOrgID(env)
			orgCfg, err := config.LoadOrgPreferences(env, orgID)
			if err != nil {
				return fmt.Errorf("failed to load org preferences: %w", err)
			}
			orgPrefs := preferences.NewOrgService(orgCfg, scope)

			accountID := orgPrefs.GetDefaultAccountID()
			if accountID == "" {
				fmt.Println(s.Help.Render("No account configured"))
				fmt.Println(s.Help.Render("Run 'tero' to complete onboarding"))
				return nil
			}

			// Get API services
			services, err := newAuthenticatedGraphQLServiceSet(cmd.Context(), cliConfig, scope)
			if err != nil {
				return err
			}
			// Datadog status queries require account scoping via X-Account-ID.
			services = services.WithAccountID(accountID)

			// Get the datadog account for this account
			ddAccount, err := services.DatadogAccounts.GetAccount(cmd.Context(), accountID)
			if err != nil {
				return fmt.Errorf("failed to get datadog account: %w", err)
			}
			if ddAccount == nil {
				fmt.Println(s.Help.Render("No Datadog account configured for this account"))
				fmt.Println(s.Help.Render("Run 'tero' to complete onboarding"))
				return nil
			}

			// Fetch status
			ddStatus, err := services.DatadogAccounts.GetStatus(cmd.Context(), ddAccount.ID)
			if err != nil {
				return fmt.Errorf("failed to fetch status: %w", err)
			}
			if ddStatus == nil {
				fmt.Println(s.Help.Render("No status found"))
				return nil
			}

			// Print status
			fmt.Println(s.Title.Render("Datadog Account Status"))
			fmt.Println(kv(s, "Health", string(ddStatus.Health)))

			fmt.Println(section(s, "Service Counts"))
			fmt.Println(kv(s, "Total", fmt.Sprintf("%d", ddStatus.ServiceCount)))
			fmt.Println(kv(s, "Active", fmt.Sprintf("%d", ddStatus.ActiveServices)))
			fmt.Println(kvStyled(s, "OK", s.Success.Render(fmt.Sprintf("%d", ddStatus.OkServices))))
			fmt.Println(kvStyled(s, "Inactive", s.Help.Render(fmt.Sprintf("%d", ddStatus.InactiveServices))))
			fmt.Println(kvStyled(s, "Disabled", s.Help.Render(fmt.Sprintf("%d", ddStatus.DisabledServices))))

			fmt.Println(section(s, "Events"))
			fmt.Println(kv(s, "Total", fmt.Sprintf("%d", ddStatus.EventCount)))
			fmt.Println(kv(s, "Analyzed", fmt.Sprintf("%d", ddStatus.AnalyzedCount)))

			fmt.Println(section(s, "Policies"))
			fmt.Println(kv(s, "Pending", fmt.Sprintf("%d", ddStatus.PendingPolicyCount)))
			fmt.Println(kv(s, "Approved", fmt.Sprintf("%d", ddStatus.ApprovedPolicyCount)))
			fmt.Println(kv(s, "Dismissed", fmt.Sprintf("%d", ddStatus.DismissedPolicyCount)))

			fmt.Println(section(s, "Onboarding"))
			readyForUseStr := fmt.Sprintf("%v", ddStatus.ReadyForUse)
			if ddStatus.ReadyForUse {
				fmt.Println(kvStyled(s, "Ready for Use", s.Success.Render(readyForUseStr)))
			} else {
				fmt.Println(kv(s, "Ready for Use", readyForUseStr))
			}

			return nil
		},
	}
}

// newDebugPrefsCmd shows the current preferences
func newDebugPrefsCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "prefs",
		Short: "Show current preferences",
		Long:  "Show the current preferences stored in the config file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			theme := styles.DetectTheme()
			s := theme.Styles
			env := cliConfig.Environment()

			printPref := func(label, value string) {
				if value == "" {
					fmt.Println(kvStyled(s, label, s.Help.Render("(not set)")))
				} else {
					fmt.Println(kv(s, label, value))
				}
			}

			// User preferences (env-level)
			userCfg, err := config.LoadUserPreferences(env)
			if err != nil {
				return fmt.Errorf("failed to load user preferences: %w", err)
			}
			userPrefs := preferences.NewUserService(userCfg, scope)

			userPrefsPath, _ := config.UserPreferencesPath(env)
			fmt.Println(s.Title.Render("User Preferences"))
			fmt.Println(s.Help.Render("Path: " + userPrefsPath))
			printPref("Theme", string(userPrefs.GetTheme()))
			printPref("Active Org ID", userPrefs.GetActiveOrgID().String())
			printPref("Role", userPrefs.GetRole())

			// Org preferences (org-level)
			orgID := userPrefs.GetActiveOrgID()
			if orgID != "" {
				orgCfg, err := config.LoadOrgPreferences(env, orgID)
				if err != nil {
					return fmt.Errorf("failed to load org preferences: %w", err)
				}
				orgPrefs := preferences.NewOrgService(orgCfg, scope)

				orgPrefsPath, _ := config.OrgPreferencesPath(env, orgID)
				fmt.Println()
				fmt.Println(s.Title.Render("Org Preferences"))
				fmt.Println(s.Help.Render("Path: " + orgPrefsPath))
				printPref("Account ID", orgPrefs.GetDefaultAccountID().String())
				printPref("Workspace ID", orgPrefs.GetDefaultWorkspaceID().String())
			}

			return nil
		},
	}
}

// newDebugGraphQLCmd runs a free-form GraphQL query
func newDebugGraphQLCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	var variables string

	cmd := &cobra.Command{
		Use:   "graphql <query>",
		Short: "Run a GraphQL query",
		Long:  "Run a free-form GraphQL query against the control plane API.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			// Parse variables if provided
			var vars map[string]interface{}
			if variables != "" {
				if err := json.Unmarshal([]byte(variables), &vars); err != nil {
					return fmt.Errorf("invalid variables JSON: %w", err)
				}
			}

			// Get API services
			services, err := newAuthenticatedGraphQLServiceSet(cmd.Context(), cliConfig, scope)
			if err != nil {
				return err
			}

			// Set account ID header if available
			env := cliConfig.Environment()
			orgCfg, _ := config.LoadOrgPreferences(env, config.ActiveOrgID(env))
			if orgCfg != nil {
				orgPrefs := preferences.NewOrgService(orgCfg, scope)
				if accountID := orgPrefs.GetDefaultAccountID(); accountID != "" {
					services = services.WithAccountID(accountID)
				}
			}

			// Execute query
			result, err := services.RawQuery(cmd.Context(), query, vars)
			if err != nil {
				return fmt.Errorf("query failed: %w", err)
			}

			// Pretty print result
			output, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format result: %w", err)
			}

			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().StringVarP(&variables, "variables", "v", "", "JSON object of variables")

	return cmd
}

// newDebugPathsCmd shows relevant file paths
func newDebugPathsCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "paths",
		Short: "Show file paths used by Tero",
		Long:  "Show the file paths used by Tero for config, data, and credentials.",
		RunE: func(cmd *cobra.Command, args []string) error {
			theme := styles.DetectTheme()
			s := theme.Styles
			env := cliConfig.Environment()

			orgID := config.ActiveOrgID(env)

			fmt.Println(s.Title.Render("Tero Paths"))
			fmt.Println(kv(s, "Environment", env))
			if orgID != "" {
				fmt.Println(kv(s, "Active Org", orgID.String()))
			}

			userPrefsPath, _ := config.UserPreferencesPath(env)
			fmt.Println(kv(s, "User Prefs", userPrefsPath))

			if orgID != "" {
				orgPrefsPath, _ := config.OrgPreferencesPath(env, orgID)
				fmt.Println(kv(s, "Org Prefs", orgPrefsPath))

				orgCfg, err := config.LoadOrgPreferences(env, orgID)
				if err == nil {
					baseDir, _ := orgCfg.BaseDir()
					fmt.Println(kv(s, "Base Dir", baseDir))

					orgPrefs := preferences.NewOrgService(orgCfg, scope)
					if accountID := orgPrefs.GetDefaultAccountID(); accountID != "" {
						storage := sqlite.NewStorageService(orgCfg)
						dbPath, _ := storage.DatabasePath(accountID.String())
						fmt.Println(kv(s, "Database", dbPath))
					}
				}
			}

			return nil
		},
	}
}
