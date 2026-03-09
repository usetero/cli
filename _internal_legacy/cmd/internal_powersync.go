package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/powersync"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/sqlite"
)

func NewInternalPowerSyncCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	scope = scope.Child("powersync")

	cmd := &cobra.Command{
		Use:   "powersync",
		Short: "PowerSync operational commands",
	}
	cmd.AddCommand(newInternalPowerSyncCaptureCmd(scope, cliConfig))
	cmd.AddCommand(newInternalPowerSyncSanitizeFixtureCmd(scope))
	return cmd
}

func newInternalPowerSyncCaptureCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	var (
		accountID string
		output    string
		duration  time.Duration
		maxBytes  int64
	)

	cmd := &cobra.Command{
		Use:   "capture",
		Short: "Capture raw PowerSync NDJSON stream lines to a fixture file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if output == "" {
				return fmt.Errorf("--output is required")
			}
			if duration <= 0 {
				return fmt.Errorf("--duration must be > 0")
			}
			if maxBytes <= 0 {
				return fmt.Errorf("--max-bytes must be > 0")
			}

			env := cliConfig.Environment()

			resolvedOutput, err := resolveCaptureOutputPath(env, output)
			if err != nil {
				return err
			}

			orgCfg, err := config.LoadOrgPreferences(env, config.ActiveOrgID(env))
			if err != nil {
				return fmt.Errorf("load org preferences: %w", err)
			}
			orgPrefs := preferences.NewOrgService(orgCfg, scope)

			if accountID == "" {
				accountID = orgPrefs.GetDefaultAccountID().String()
			}
			if accountID == "" {
				return fmt.Errorf("no account configured; pass --account-id or complete onboarding first")
			}

			authService := newAuthService(cliConfig, scope)

			storage := sqlite.NewStorageService(orgCfg)
			dbPath, err := storage.DatabasePath(accountID)
			if err != nil {
				return fmt.Errorf("resolve database path: %w", err)
			}
			db, err := sqlite.Open(cmd.Context(), dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer db.Close()

			capture, err := powersync.NewNDJSONStreamCapture(resolvedOutput, maxBytes, scope)
			if err != nil {
				return fmt.Errorf("create stream capture: %w", err)
			}

			syncer := powersync.NewSyncer(
				cliConfig.PowerSyncOrigin,
				authService,
				scope,
				powersync.WithStreamCapture(capture),
			)

			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			if err := syncer.Start(ctx, db, accountID, nil); err != nil {
				return fmt.Errorf("start syncer: %w", err)
			}
			defer syncer.Stop()

			fmt.Printf("Capturing PowerSync stream for account %s\n", accountID)
			fmt.Printf("Output: %s\n", resolvedOutput)
			fmt.Printf("Duration: %s\n", duration)
			fmt.Printf("Max bytes: %d\n", maxBytes)

			sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()

			timer := time.NewTimer(duration)
			defer timer.Stop()

			select {
			case <-timer.C:
				fmt.Println("Capture complete")
			case <-sigCtx.Done():
				fmt.Println("Capture interrupted")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&accountID, "account-id", "", "Account ID to sync (defaults to org preference)")
	cmd.Flags().StringVar(&output, "output", "", "Output path for NDJSON fixture (required)")
	cmd.Flags().DurationVar(&duration, "duration", 90*time.Second, "Capture duration")
	cmd.Flags().Int64Var(&maxBytes, "max-bytes", 25*1024*1024, "Maximum capture file size in bytes")

	return cmd
}

func resolveCaptureOutputPath(env, output string) (string, error) {
	if filepath.IsAbs(output) {
		return output, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(homeDir, ".tero", "environments", env, output), nil
}
