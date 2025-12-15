package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/keyring"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/workos"
)

func NewAuthCmd(logger log.Logger, cliConfig *config.CLIConfig) *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  "Commands for managing authentication with Tero.",
	}

	authCmd.AddCommand(newLoginCmd(logger, cliConfig))
	authCmd.AddCommand(newTokenCmd(logger, cliConfig))
	authCmd.AddCommand(newLogoutCmd(logger, cliConfig))
	authCmd.AddCommand(newStatusCmd(logger, cliConfig))

	return authCmd
}

func newLoginCmd(logger log.Logger, cliConfig *config.CLIConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Tero",
		Long:  "Authenticate with Tero using the device authorization flow.",
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace := cliConfig.Namespace()
			tokenStore := keyring.New(namespace)
			workosClient := workos.NewClient(cliConfig.WorkOSClientID)
			authService := auth.NewService(workosClient, tokenStore, logger)

			ctx := cmd.Context()

			// Start device authorization
			deviceAuth, err := authService.StartDeviceAuth(ctx)
			if err != nil {
				return fmt.Errorf("failed to start authentication: %w", err)
			}

			// Print instructions
			fmt.Println("Opening browser to authenticate...")
			fmt.Printf("If the browser doesn't open, visit: %s\n", deviceAuth.VerificationURIComplete)
			fmt.Printf("And enter code: %s\n", deviceAuth.UserCode)

			// Try to open browser
			_ = browser.OpenURL(deviceAuth.VerificationURIComplete)

			// Poll for completion
			fmt.Println("\nWaiting for authentication...")
			interval := time.Duration(deviceAuth.Interval) * time.Second
			if interval == 0 {
				interval = 5 * time.Second
			}

			result, err := authService.WaitForAuth(ctx, deviceAuth.DeviceCode, interval)
			if err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			fmt.Printf("\nAuthenticated as %s\n", result.User.Email)
			return nil
		},
	}
}

func newTokenCmd(logger log.Logger, cliConfig *config.CLIConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "token",
		Short: "Print the current access token",
		Long:  "Print the current access token to stdout. Refreshes the token if expired.",
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace := cliConfig.Namespace()
			tokenStore := keyring.New(namespace)
			workosClient := workos.NewClient(cliConfig.WorkOSClientID)
			authService := auth.NewService(workosClient, tokenStore, logger)

			token, err := authService.GetAccessToken(cmd.Context())
			if err != nil {
				return fmt.Errorf("not authenticated: run 'tero auth login' first")
			}

			fmt.Print(token)
			return nil
		},
	}
}

func newLogoutCmd(logger log.Logger, cliConfig *config.CLIConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear stored credentials",
		Long:  "Remove stored authentication credentials.",
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace := cliConfig.Namespace()
			tokenStore := keyring.New(namespace)
			workosClient := workos.NewClient(cliConfig.WorkOSClientID)
			authService := auth.NewService(workosClient, tokenStore, logger)

			if err := authService.ClearTokens(); err != nil {
				return fmt.Errorf("failed to clear credentials: %w", err)
			}

			fmt.Println("Logged out successfully")
			return nil
		},
	}
}

func newStatusCmd(logger log.Logger, cliConfig *config.CLIConfig) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Long:  "Show current authentication status including user and token state.",
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace := cliConfig.Namespace()
			tokenStore := keyring.New(namespace)

			// Get raw token to check status
			token, err := tokenStore.Get("access_token")
			if err != nil {
				return fmt.Errorf("failed to check authentication: %w", err)
			}

			if token == "" {
				fmt.Println("Not authenticated")
				fmt.Println("Run 'tero auth login' to authenticate")
				return nil
			}

			// Parse token to get claims
			claims, err := parseTokenClaims(token)
			if err != nil {
				// Token exists but can't be parsed - still authenticated, just can't show details
				fmt.Println("Authenticated (invalid token format)")
				return nil //nolint:nilerr // intentional - we report status, not error
			}

			// Check expiration
			expired := false
			var expiresAt time.Time
			if exp, ok := claims["exp"].(float64); ok {
				expiresAt = time.Unix(int64(exp), 0)
				expired = time.Now().After(expiresAt)
			}

			// Extract user info
			email, _ := claims["email"].(string)
			orgID, _ := claims["org_id"].(string)

			fmt.Println("Authenticated")
			if email != "" {
				fmt.Printf("  User: %s\n", email)
			}
			if orgID != "" {
				fmt.Printf("  Organization: %s\n", orgID)
			}
			if !expiresAt.IsZero() {
				if expired {
					fmt.Printf("  Token: expired at %s\n", expiresAt.Format(time.RFC3339))
					fmt.Println("  Run 'tero auth login' to re-authenticate")
				} else {
					fmt.Printf("  Token: valid until %s\n", expiresAt.Format(time.RFC3339))
				}
			}

			return nil
		},
	}
}

// parseTokenClaims extracts claims from a JWT token without verification.
func parseTokenClaims(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode token: %w", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return claims, nil
}
