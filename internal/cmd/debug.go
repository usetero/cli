package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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
			theme := styles.NewTheme(true)
			s := theme.Styles
			namespace := cliConfig.Namespace()

			// Load preferences to get account ID
			cfg, err := config.Load(namespace)
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
			fmt.Println()

			fmt.Printf("  %-25s %s\n", s.Help.Render("Status:"), s.Body.Render(string(ddStatus.Status)))
			fmt.Printf("  %-25s %s\n", s.Help.Render("Percent Complete:"), s.Body.Render(fmt.Sprintf("%.1f%%", ddStatus.PercentComplete*100)))
			fmt.Println()

			fmt.Println(s.Title.Render("Service Counts"))
			fmt.Printf("  %-25s %s\n", s.Help.Render("Total:"), s.Body.Render(fmt.Sprintf("%d", ddStatus.ServiceCount)))
			fmt.Printf("  %-25s %s\n", s.Help.Render("Active:"), s.Body.Render(fmt.Sprintf("%d", ddStatus.ActiveServices)))
			fmt.Printf("  %-25s %s\n", s.Help.Render("Ready:"), s.Success.Render(fmt.Sprintf("%d", ddStatus.ReadyServices)))
			fmt.Printf("  %-25s %s\n", s.Help.Render("Analyzing:"), s.Body.Render(fmt.Sprintf("%d", ddStatus.AnalyzingServices)))
			fmt.Printf("  %-25s %s\n", s.Help.Render("Discovering:"), s.Body.Render(fmt.Sprintf("%d", ddStatus.DiscoveringServices)))
			fmt.Printf("  %-25s %s\n", s.Help.Render("Stale:"), s.Body.Render(fmt.Sprintf("%d", ddStatus.StaleServices)))
			fmt.Printf("  %-25s %s\n", s.Help.Render("Broken:"), s.Error.Render(fmt.Sprintf("%d", ddStatus.BrokenServices)))
			fmt.Printf("  %-25s %s\n", s.Help.Render("Inactive:"), s.Help.Render(fmt.Sprintf("%d", ddStatus.InactiveServices)))
			fmt.Printf("  %-25s %s\n", s.Help.Render("Disabled:"), s.Help.Render(fmt.Sprintf("%d", ddStatus.DisabledServices)))
			fmt.Println()

			fmt.Println(s.Title.Render("Volume"))
			fmt.Printf("  %-25s %s\n", s.Help.Render("Service Volume:"), s.Body.Render(fmt.Sprintf("%d", ddStatus.ServiceLogVolume)))
			fmt.Printf("  %-25s %s\n", s.Help.Render("Discovered Volume:"), s.Body.Render(fmt.Sprintf("%d", ddStatus.DiscoveredLogVolume)))

			// Show onboarding readiness
			fmt.Println()
			fmt.Println(s.Title.Render("Onboarding"))
			fmt.Printf("  %-25s %s\n", s.Help.Render("Saved Count:"), s.Body.Render(fmt.Sprintf("%d / 50", ddStatus.SavedCount)))
			readyForUseStr := fmt.Sprintf("%v", ddStatus.ReadyForUse)
			if ddStatus.ReadyForUse {
				fmt.Printf("  %-25s %s\n", s.Help.Render("Ready for Use:"), s.Success.Render(readyForUseStr))
			} else {
				fmt.Printf("  %-25s %s\n", s.Help.Render("Ready for Use:"), s.Body.Render(readyForUseStr))
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
			theme := styles.NewTheme(true)
			s := theme.Styles
			namespace := cliConfig.Namespace()

			cfg, err := config.Load(namespace)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			prefs := preferences.NewService(cfg, scope)

			// Get config path for display
			homeDir, _ := os.UserHomeDir()
			var configPath string
			if namespace == "" {
				configPath = homeDir + "/.tero/config.yaml"
			} else {
				configPath = homeDir + "/.tero/" + namespace + "/config.yaml"
			}

			fmt.Println(s.Title.Render("Preferences"))
			fmt.Println(s.Help.Render("Path: " + configPath))
			fmt.Println()

			// Use preferences service to get values
			printPref := func(label, value string) {
				if value == "" {
					fmt.Printf("  %-25s %s\n", s.Help.Render(label+":"), s.Help.Render("(not set)"))
				} else {
					fmt.Printf("  %-25s %s\n", s.Help.Render(label+":"), s.Body.Render(value))
				}
			}

			printPref("Organization ID", prefs.GetDefaultOrgID().String())
			printPref("Organization Name", prefs.GetDefaultOrgName())
			printPref("Account ID", prefs.GetDefaultAccountID().String())
			printPref("Workspace ID", prefs.GetDefaultWorkspaceID().String())
			printPref("Role", prefs.GetRole())
			printPref("Email", prefs.GetEmail())

			// Services (list)
			services := prefs.GetServices()
			if len(services) == 0 {
				fmt.Printf("  %-25s %s\n", s.Help.Render("Services:"), s.Help.Render("(none)"))
			} else {
				fmt.Printf("  %-25s %s\n", s.Help.Render("Services:"), s.Body.Render(fmt.Sprintf("%v", services)))
			}

			// Has seen greeting
			fmt.Printf("  %-25s %s\n", s.Help.Render("Has Seen Greeting:"), s.Body.Render(fmt.Sprintf("%v", prefs.GetHasSeenGreeting())))

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
			cfg, _ := config.Load(cliConfig.Namespace())
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
			theme := styles.NewTheme(true)
			s := theme.Styles
			namespace := cliConfig.Namespace()

			homeDir, _ := os.UserHomeDir()

			fmt.Println(s.Title.Render("Tero Paths"))
			fmt.Println()
			fmt.Printf("  %-20s %s\n", s.Help.Render("Namespace:"), s.Body.Render(ifEmpty(namespace, "(production)")))
			fmt.Println()

			// Config path
			var configPath string
			if namespace == "" {
				configPath = homeDir + "/.tero/config.yaml"
			} else {
				configPath = homeDir + "/.tero/" + namespace + "/config.yaml"
			}
			fmt.Printf("  %-20s %s\n", s.Help.Render("Config:"), s.Body.Render(configPath))

			// Data directory
			cfg, err := config.Load(namespace)
			if err == nil {
				baseDir, _ := cfg.BaseDir()
				fmt.Printf("  %-20s %s\n", s.Help.Render("Base Dir:"), s.Body.Render(baseDir))

				// Show database path if account is configured
				accountID := cfg.Get("account_id")
				if accountID != "" {
					storage := sqlite.NewStorageService(cfg)
					dbPath, _ := storage.DatabasePath(accountID)
					fmt.Printf("  %-20s %s\n", s.Help.Render("Database:"), s.Body.Render(dbPath))
				}
			}

			return nil
		},
	}
}

// getAPIServices creates authenticated API services
func getAPIServices(ctx context.Context, scope log.Scope, cliConfig *config.CLIConfig) (api.APIServices, error) {
	namespace := cliConfig.Namespace()
	tokenStore := keyring.New(namespace)
	workosClient := workos.NewClient(cliConfig.WorkOSClientID, cliConfig.APIEndpoint, cliConfig.PowerSyncEndpoint)
	authService := auth.NewService(workosClient, tokenStore, scope)

	// Verify we're authenticated
	_, err := authService.GetAccessToken(ctx)
	if err != nil {
		return api.APIServices{}, fmt.Errorf("not authenticated: run 'tero auth login' first")
	}

	return api.NewServices(cliConfig.APIEndpoint+"/graphql", authService, scope), nil
}

func ifEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
