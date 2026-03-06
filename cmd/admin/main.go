// Admin tool for internal operations. Not shipped — used via Taskfile only.
//
// Requires WORKOS_API_KEY in the environment (provided by doppler run).
package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	workosadmin "github.com/usetero/cli/_internal_legacy/workos/admin"
	"github.com/usetero/cli/internal/domains/identity"
	authkeyring "github.com/usetero/cli/internal/infrastructure/auth/keyring"
)

func main() {
	root := &cobra.Command{
		Use:   "admin",
		Short: "Internal admin tooling (not shipped)",
	}

	root.AddCommand(newJoinOrgCmd())
	root.AddCommand(newLeaveOrgCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func newJoinOrgCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "join-org <workos-org-id>",
		Short: "Join a client organization (for support/debugging)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, userID, err := setup()
			if err != nil {
				return err
			}

			m, err := client.CreateMembership(ctx, userID, args[0])
			if err != nil {
				return fmt.Errorf("failed to join org: %w", err)
			}

			fmt.Printf("Joined org %s (membership %s)\n", m.OrganizationID, m.ID)
			fmt.Printf("Run 'tero auth switch' to switch to the organization\n")
			return nil
		},
	}
}

func newLeaveOrgCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "leave-org <workos-org-id>",
		Short: "Leave a client organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, userID, err := setup()
			if err != nil {
				return err
			}

			orgID := args[0]
			m, err := client.FindMembership(ctx, userID, orgID)
			if err != nil {
				return fmt.Errorf("failed to find membership: %w", err)
			}

			if err := client.DeleteMembership(ctx, m.ID); err != nil {
				return fmt.Errorf("failed to leave org: %w", err)
			}

			fmt.Printf("Left org %s (deleted membership %s)\n", orgID, m.ID)
			return nil
		},
	}
}

// setup creates the admin client and resolves the current user's WorkOS ID.
func setup() (*workosadmin.Client, string, error) {
	apiKey := os.Getenv("WORKOS_API_KEY")
	if apiKey == "" {
		return nil, "", fmt.Errorf("WORKOS_API_KEY is required (use doppler run)")
	}

	tokenStore, err := currentTokenStore()
	if err != nil {
		return nil, "", fmt.Errorf("failed to initialize token store: %w", err)
	}

	accessToken, err := tokenStore.AccessToken()
	if err != nil {
		return nil, "", fmt.Errorf("failed to load access token: %w", err)
	}
	if accessToken == "" {
		return nil, "", fmt.Errorf("not authenticated: no access token in keyring")
	}

	userID, err := userIDFromAccessToken(string(accessToken))
	if err != nil {
		return nil, "", fmt.Errorf("failed to resolve user ID from access token: %w", err)
	}

	return workosadmin.NewClient(apiKey), userID, nil
}

func currentTokenStore() (*identity.KeyringTokenStore, error) {
	env := os.Getenv("TERO_ENV")
	if env == "" {
		env = "local"
	}
	store, err := authkeyring.NewStore(env)
	if err != nil {
		return nil, err
	}
	return identity.NewKeyringTokenStore(store), nil
}

type tokenClaims struct {
	Subject string `json:"sub"`
}

func userIDFromAccessToken(token string) (string, error) {
	claims, err := parseTokenClaims(token)
	if err != nil {
		return "", err
	}
	if claims.Subject == "" {
		return "", fmt.Errorf("token subject claim is empty")
	}
	return claims.Subject, nil
}

func parseTokenClaims(token string) (tokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return tokenClaims{}, fmt.Errorf("invalid jwt format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return tokenClaims{}, fmt.Errorf("decode jwt payload: %w", err)
	}

	var claims tokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return tokenClaims{}, fmt.Errorf("unmarshal jwt claims: %w", err)
	}
	return claims, nil
}
