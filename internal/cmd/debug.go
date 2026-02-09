package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/usetero/cli/internal/api"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/keyring"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/workos"
)

func NewDebugCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	scope = scope.Child("debug")

	debugCmd := &cobra.Command{
		Use:   "debug",
		Short: "Debug and diagnostic commands",
		Long:  "Commands for debugging and diagnosing issues with Tero.",
	}

	debugCmd.AddCommand(newDebugStatusCmd(scope, cliConfig))
	debugCmd.AddCommand(newDebugPrefsCmd(scope, cliConfig))
	debugCmd.AddCommand(newDebugGraphQLCmd(scope, cliConfig))
	debugCmd.AddCommand(newDebugPathsCmd(scope, cliConfig))

	return debugCmd
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

			// Load preferences to get account ID
			orgID := config.ActiveOrgID(env)
			cfg, err := config.Load(env, orgID)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			prefs := preferences.NewService(cfg, scope)

			accountID := prefs.GetDefaultAccountID()
			if accountID == "" {
				fmt.Println(s.Help.Render("No account configured"))
				fmt.Println(s.Help.Render("Run 'tero' to complete onboarding"))
				return nil
			}

			// Get API services
			services, err := getAPIServices(cmd.Context(), scope, cliConfig)
			if err != nil {
				return err
			}

			// Get the datadog account for this account
			ddAccount, err := services.DatadogAccounts.GetAccount(cmd.Context(), accountID.String())
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
			fmt.Println(kv(s, "Status", string(ddStatus.Status)))
			fmt.Println(kv(s, "Percent Complete", fmt.Sprintf("%.1f%%", ddStatus.PercentComplete*100)))

			fmt.Println(section(s, "Service Counts"))
			fmt.Println(kv(s, "Total", fmt.Sprintf("%d", ddStatus.ServiceCount)))
			fmt.Println(kv(s, "Active", fmt.Sprintf("%d", ddStatus.ActiveServices)))
			fmt.Println(kvStyled(s, "Ready", s.Success.Render(fmt.Sprintf("%d", ddStatus.ReadyServices))))
			fmt.Println(kv(s, "Analyzing", fmt.Sprintf("%d", ddStatus.AnalyzingServices)))
			fmt.Println(kv(s, "Discovering", fmt.Sprintf("%d", ddStatus.DiscoveringServices)))
			fmt.Println(kv(s, "Stale", fmt.Sprintf("%d", ddStatus.StaleServices)))
			fmt.Println(kvStyled(s, "Broken", s.Error.Render(fmt.Sprintf("%d", ddStatus.BrokenServices))))
			fmt.Println(kvStyled(s, "Inactive", s.Help.Render(fmt.Sprintf("%d", ddStatus.InactiveServices))))
			fmt.Println(kvStyled(s, "Disabled", s.Help.Render(fmt.Sprintf("%d", ddStatus.DisabledServices))))

			fmt.Println(section(s, "Volume"))
			fmt.Println(kv(s, "Service Volume", fmt.Sprintf("%d", ddStatus.ServiceLogVolume)))
			fmt.Println(kv(s, "Discovered Volume", fmt.Sprintf("%d", ddStatus.DiscoveredLogVolume)))

			fmt.Println(section(s, "Onboarding"))
			fmt.Println(kv(s, "Analyzed Count", fmt.Sprintf("%d / 50", ddStatus.AnalyzedCount)))
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

			orgID := config.ActiveOrgID(env)
			cfg, err := config.Load(env, orgID)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			prefs := preferences.NewService(cfg, scope)

			// Get config path for display
			configPath, _ := config.ConfigPath(env, orgID)

			fmt.Println(s.Title.Render("Preferences"))
			fmt.Println(s.Help.Render("Path: " + configPath))

			printPref := func(label, value string) {
				if value == "" {
					fmt.Println(kvStyled(s, label, s.Help.Render("(not set)")))
				} else {
					fmt.Println(kv(s, label, value))
				}
			}

			printPref("Organization ID", prefs.GetDefaultOrgID().String())
			printPref("Organization Name", prefs.GetDefaultOrgName())
			printPref("Account ID", prefs.GetDefaultAccountID().String())
			printPref("Workspace ID", prefs.GetDefaultWorkspaceID().String())
			printPref("Role", prefs.GetRole())
			printPref("Email", prefs.GetEmail())

			services := prefs.GetServices()
			if len(services) == 0 {
				fmt.Println(kvStyled(s, "Services", s.Help.Render("(none)")))
			} else {
				fmt.Println(kv(s, "Services", fmt.Sprintf("%v", services)))
			}

			fmt.Println(kv(s, "Has Seen Greeting", fmt.Sprintf("%v", prefs.GetHasSeenGreeting())))

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
			services, err := getAPIServices(cmd.Context(), scope, cliConfig)
			if err != nil {
				return err
			}

			// Set account ID header if available
			env := cliConfig.Environment()
			cfg, _ := config.Load(env, config.ActiveOrgID(env))
			if cfg != nil {
				prefs := preferences.NewService(cfg, scope)
				if accountID := prefs.GetDefaultAccountID(); accountID != "" {
					services.SetAccountID(accountID)
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
				fmt.Println(kv(s, "Active Org", orgID))
			}

			configPath, _ := config.ConfigPath(env, orgID)
			fmt.Println(kv(s, "Config", configPath))

			cfg, err := config.Load(env, orgID)
			if err == nil {
				baseDir, _ := cfg.BaseDir()
				fmt.Println(kv(s, "Base Dir", baseDir))

				accountID := cfg.Get("account_id")
				if accountID != "" {
					storage := sqlite.NewStorageService(cfg)
					dbPath, _ := storage.DatabasePath(accountID)
					fmt.Println(kv(s, "Database", dbPath))
				}
			}

			return nil
		},
	}
}

// getAPIServices creates authenticated API services
func getAPIServices(ctx context.Context, scope log.Scope, cliConfig *config.CLIConfig) (api.APIServices, error) {
	env := cliConfig.Environment()
	tokenStore := keyring.New(env)
	workosClient := workos.NewClient(cliConfig.WorkOSClientID, cliConfig.APIEndpoint, cliConfig.PowerSyncEndpoint)
	authService := auth.NewService(workosClient, tokenStore, scope)

	// Verify we're authenticated
	_, err := authService.GetAccessToken(ctx)
	if err != nil {
		return api.APIServices{}, fmt.Errorf("not authenticated: run 'tero auth login' first")
	}

	return api.NewServices(cliConfig.APIEndpoint+"/graphql", authService, scope), nil
}
