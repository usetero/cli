package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"github.com/usetero/cli/internal/auth"
	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/keyring"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/workos"
)

// org represents an organization for selection
type org struct {
	ID       domain.OrganizationID
	Name     string
	WorkosID domain.WorkosOrganizationID
}

func NewAuthCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	scope = scope.Child("auth")

	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  "Commands for managing authentication with Tero.",
	}

	authCmd.AddCommand(newLoginCmd(scope, cliConfig))
	authCmd.AddCommand(newTokenCmd(scope, cliConfig))
	authCmd.AddCommand(newLogoutCmd(scope, cliConfig))
	authCmd.AddCommand(newStatusCmd(scope, cliConfig))
	authCmd.AddCommand(newSwitchCmd(scope, cliConfig))

	return authCmd
}

func newLoginCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	var noOrg bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Tero",
		Long:  "Authenticate with Tero using the device authorization flow.",
		RunE: func(cmd *cobra.Command, args []string) error {
			theme := styles.DetectTheme()
			s := theme.Styles
			env := cliConfig.Environment()
			tokenStore := keyring.New(env)
			workosClient := workos.NewClient(cliConfig.WorkOSClientID, cliConfig.APIEndpoint, cliConfig.PowerSyncEndpoint, cliConfig.ChatEndpoint)
			authService := auth.NewService(workosClient, tokenStore, scope)

			ctx := cmd.Context()

			// Start device authorization
			deviceAuth, err := authService.StartDeviceAuth(ctx)
			if err != nil {
				return fmt.Errorf("failed to start authentication: %w", err)
			}

			// Print instructions
			fmt.Println(s.Body.Render("Opening browser to authenticate..."))
			fmt.Println(s.Help.Render("If the browser doesn't open, visit: ") + s.URL.Render(deviceAuth.VerificationURIComplete))
			fmt.Println(s.Help.Render("And enter code: ") + s.Title.Render(deviceAuth.UserCode))

			// Try to open browser
			_ = browser.OpenURL(deviceAuth.VerificationURIComplete)

			// Poll for completion
			fmt.Println(s.Help.Render("\nWaiting for authentication..."))
			interval := time.Duration(deviceAuth.Interval) * time.Second
			if interval == 0 {
				interval = 5 * time.Second
			}

			result, err := authService.WaitForAuth(ctx, deviceAuth.DeviceCode, interval)
			if err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			fmt.Println(s.Success.Render("\n✓ Authenticated as " + result.User.Email))

			// If --no-org flag is set, refresh token without org scope and return
			if noOrg {
				_, err := authService.RefreshTokenWithoutOrganization(ctx)
				if err != nil {
					return fmt.Errorf("failed to get user-scoped token: %w", err)
				}
				fmt.Println(s.Help.Render("Using user-scoped token (no organization)"))
				return nil
			}

			// Fetch organizations
			services := graphql.NewServiceSet(cliConfig.APIEndpoint+"/graphql", authService, scope)
			orgs, err := fetchOrganizations(ctx, services)
			if err != nil {
				// Don't fail login if org fetch fails - user can use 'tero auth switch' later
				fmt.Println(s.Help.Render("\nCould not fetch organizations: " + err.Error()))
				fmt.Println(s.Help.Render("Run 'tero auth switch' to select an organization later"))
				return nil //nolint:nilerr // intentional - login succeeded, org fetch is optional
			}

			if len(orgs) == 0 {
				fmt.Println(s.Help.Render("\nNo organizations found. Create one to get started."))
				return nil
			}

			// Auto-select if only one org
			var selectedOrg *org
			if len(orgs) == 1 {
				selectedOrg = &orgs[0]
				fmt.Println(s.Body.Render("\nUsing organization: ") + s.Title.Render(selectedOrg.Name))
			} else {
				// Prompt user to select
				selectedOrg, err = promptOrgSelection(orgs)
				if err != nil {
					return err
				}
			}

			// Refresh token with selected org
			_, err = authService.RefreshTokenWithOrganization(ctx, selectedOrg.WorkosID)
			if err != nil {
				return fmt.Errorf("failed to select organization: %w", err)
			}

			// Set active org so preferences and databases are org-scoped
			_ = config.SetActiveOrgID(env, selectedOrg.ID)

			fmt.Println(s.Success.Render("✓ Switched to organization: " + selectedOrg.Name))
			return nil
		},
	}

	cmd.Flags().BoolVar(&noOrg, "no-org", false, "Skip organization selection (for bootstrap flow)")

	return cmd
}

func newSwitchCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "switch [organization]",
		Short: "Switch to a different organization",
		Long:  "Switch to a different organization. Lists available orgs if none specified.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			theme := styles.DetectTheme()
			s := theme.Styles
			env := cliConfig.Environment()
			tokenStore := keyring.New(env)
			workosClient := workos.NewClient(cliConfig.WorkOSClientID, cliConfig.APIEndpoint, cliConfig.PowerSyncEndpoint, cliConfig.ChatEndpoint)
			authService := auth.NewService(workosClient, tokenStore, scope)

			ctx := cmd.Context()

			// Verify authenticated
			if _, err := authService.GetAccessToken(ctx); err != nil {
				return fmt.Errorf("not authenticated: run 'tero auth login' first")
			}

			// Fetch organizations
			services := graphql.NewServiceSet(cliConfig.APIEndpoint+"/graphql", authService, scope)
			orgs, err := fetchOrganizations(ctx, services)
			if err != nil {
				return fmt.Errorf("failed to fetch organizations: %w", err)
			}

			if len(orgs) == 0 {
				fmt.Println(s.Help.Render("No organizations found"))
				return nil
			}

			var selectedOrg *org

			// If org name provided as argument, find it
			if len(args) == 1 {
				orgName := args[0]
				for i := range orgs {
					if strings.EqualFold(orgs[i].Name, orgName) {
						selectedOrg = &orgs[i]
						break
					}
				}
				if selectedOrg == nil {
					return fmt.Errorf("organization %q not found", orgName)
				}
			} else {
				// Prompt user to select
				selectedOrg, err = promptOrgSelection(orgs)
				if err != nil {
					return err
				}
			}

			// Refresh token with selected org
			_, err = authService.RefreshTokenWithOrganization(ctx, selectedOrg.WorkosID)
			if err != nil {
				return fmt.Errorf("failed to switch organization: %w", err)
			}

			// Set active org so preferences and databases are org-scoped
			_ = config.SetActiveOrgID(env, selectedOrg.ID)

			fmt.Println(s.Success.Render("✓ Switched to organization: " + selectedOrg.Name))
			return nil
		},
	}
}

func newTokenCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "token",
		Short: "Print the current access token",
		Long:  "Print the current access token to stdout. Refreshes the token if expired.",
		RunE: func(cmd *cobra.Command, args []string) error {
			env := cliConfig.Environment()
			tokenStore := keyring.New(env)
			workosClient := workos.NewClient(cliConfig.WorkOSClientID, cliConfig.APIEndpoint, cliConfig.PowerSyncEndpoint, cliConfig.ChatEndpoint)
			authService := auth.NewService(workosClient, tokenStore, scope)

			token, err := authService.GetAccessToken(cmd.Context())
			if err != nil {
				return fmt.Errorf("not authenticated: run 'tero auth login' first")
			}

			fmt.Print(token)
			return nil
		},
	}
}

func newLogoutCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear stored credentials",
		Long:  "Remove stored authentication credentials.",
		RunE: func(cmd *cobra.Command, args []string) error {
			theme := styles.DetectTheme()
			s := theme.Styles
			env := cliConfig.Environment()
			tokenStore := keyring.New(env)
			workosClient := workos.NewClient(cliConfig.WorkOSClientID, cliConfig.APIEndpoint, cliConfig.PowerSyncEndpoint, cliConfig.ChatEndpoint)
			authService := auth.NewService(workosClient, tokenStore, scope)

			if err := authService.ClearTokens(); err != nil {
				return fmt.Errorf("failed to clear credentials: %w", err)
			}

			fmt.Println(s.Success.Render("✓ Logged out successfully"))
			return nil
		},
	}
}

func newStatusCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Long:  "Show current authentication status including user and token state.",
		RunE: func(cmd *cobra.Command, args []string) error {
			theme := styles.DetectTheme()
			s := theme.Styles
			env := cliConfig.Environment()
			tokenStore := keyring.New(env)

			// Get raw token to check status
			token, err := tokenStore.Get("access_token")
			if err != nil {
				return fmt.Errorf("failed to check authentication: %w", err)
			}

			if token == "" {
				fmt.Println(s.Help.Render("Not authenticated"))
				fmt.Println(s.Help.Render("Run 'tero auth login' to authenticate"))
				return nil
			}

			// Parse token to get claims
			claims, err := auth.ParseToken(token)
			if err != nil {
				// Token exists but can't be parsed - still authenticated, just can't show details
				fmt.Println(s.Body.Render("Authenticated (invalid token format)"))
				return nil //nolint:nilerr // intentional - we report status, not error
			}

			// Check expiration
			expiresAt := claims.ExpiresAt()
			expired := claims.IsExpired()

			// Extract user info
			email := claims.Email
			workosOrgID := claims.OrgID

			fmt.Println(s.Success.Render("✓ Authenticated"))
			if email != "" {
				fmt.Println(kv(s, "User", email))
			}

			// Try to get org name from API
			if workosOrgID != "" {
				orgName := workosOrgID // fallback to ID
				workosClient := workos.NewClient(cliConfig.WorkOSClientID, cliConfig.APIEndpoint, cliConfig.PowerSyncEndpoint, cliConfig.ChatEndpoint)
				authService := auth.NewService(workosClient, tokenStore, scope)
				services := graphql.NewServiceSet(cliConfig.APIEndpoint+"/graphql", authService, scope)
				if orgs, err := fetchOrganizations(cmd.Context(), services); err == nil {
					for _, o := range orgs {
						if o.WorkosID == domain.WorkosOrganizationID(workosOrgID) {
							orgName = o.Name
							break
						}
					}
				}
				fmt.Println(kv(s, "Organization", orgName))
			} else {
				fmt.Println(kvStyled(s, "Organization", s.Help.Render("(none)")))
			}

			if !expiresAt.IsZero() {
				if expired {
					fmt.Println(kvStyled(s, "Token", s.Error.Render("expired at "+expiresAt.Format(time.RFC3339))))
					fmt.Println(s.Help.Render("Run 'tero auth login' to re-authenticate"))
				} else {
					fmt.Println(kv(s, "Token", "valid until "+expiresAt.Format(time.RFC3339)))
				}
			}

			return nil
		},
	}
}

// fetchOrganizations fetches the list of organizations from the API
func fetchOrganizations(ctx context.Context, services graphql.ServiceSet) ([]org, error) {
	apiOrgs, err := services.Organizations.List(ctx)
	if err != nil {
		return nil, err
	}

	orgs := make([]org, 0, len(apiOrgs))
	for _, o := range apiOrgs {
		orgs = append(orgs, org{
			ID:       o.ID,
			Name:     o.Name,
			WorkosID: o.WorkosOrganizationID,
		})
	}
	return orgs, nil
}

// promptOrgSelection prompts the user to select an organization
func promptOrgSelection(orgs []org) (*org, error) {
	theme := styles.DetectTheme()
	s := theme.Styles
	fmt.Println(s.Title.Render("\nSelect an organization:"))
	for i, o := range orgs {
		fmt.Printf("  %s %s\n", s.Help.Render(fmt.Sprintf("%d.", i+1)), s.Body.Render(o.Name))
	}
	fmt.Print(s.Action.Render("Enter number: "))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	num, err := strconv.Atoi(input)
	if err != nil || num < 1 || num > len(orgs) {
		return nil, fmt.Errorf("invalid selection: %s", input)
	}

	return &orgs[num-1], nil
}
