// Admin tool for internal operations. Not shipped — used via Taskfile only.
//
// Requires WORKOS_API_KEY in the environment (provided by doppler run).
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/usetero/cli/internal/auth"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/keyring"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/workos"
	workosadmin "github.com/usetero/cli/internal/workos/admin"
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
			client, userID, err := setup(cmd)
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
			client, userID, err := setup(cmd)
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
func setup(cmd *cobra.Command) (*workosadmin.Client, string, error) {
	apiKey := os.Getenv("WORKOS_API_KEY")
	if apiKey == "" {
		return nil, "", fmt.Errorf("WORKOS_API_KEY is required (use doppler run)")
	}

	logger := log.Wrap(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))
	scope := log.RootScope(logger)

	cliConfig := config.LoadCLIConfig()
	tokenStore := keyring.New(cliConfig.Namespace())
	workosClient := workos.NewClient(cliConfig.WorkOSClientID, cliConfig.APIEndpoint, cliConfig.PowerSyncEndpoint)
	authService := auth.NewService(workosClient, tokenStore, scope)

	userID, err := authService.GetUserID(cmd.Context())
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user ID (are you logged in?): %w", err)
	}

	return workosadmin.NewClient(apiKey), userID, nil
}
