package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/format"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
)

// accountServices resolves an authenticated, account-scoped service set from
// the active org's default account. Returns a helpful error when the user is
// not authenticated or has not finished onboarding.
func accountServices(ctx context.Context, cliConfig *config.CLIConfig, scope log.Scope) (graphql.ServiceSet, error) {
	services, err := newAuthenticatedGraphQLServiceSet(ctx, cliConfig, scope)
	if err != nil {
		return graphql.ServiceSet{}, err
	}

	env := cliConfig.Environment()
	orgID := config.ActiveOrgID(env)
	if orgID == "" {
		return graphql.ServiceSet{}, fmt.Errorf("no organization selected — run 'tero' to complete onboarding")
	}
	orgCfg, err := config.LoadOrgPreferences(env, orgID)
	if err != nil {
		return graphql.ServiceSet{}, fmt.Errorf("load org preferences: %w", err)
	}
	accountID := preferences.NewOrgService(orgCfg, scope).GetDefaultAccountID()
	if accountID == "" {
		return graphql.ServiceSet{}, fmt.Errorf("no account configured — run 'tero' to complete onboarding")
	}
	return services.WithAccountID(accountID), nil
}

// newTabWriter returns a tabwriter writing to stdout with padded columns.
func newTabWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
}

func cost(p *float64) string {
	if p == nil {
		return "—"
	}
	return format.YearlyCost(*p)
}

func rate(p *float64) string {
	if p == nil {
		return "—"
	}
	return format.Volume(*p) + "/hr"
}

// NewIssuesCmd lists the account's active issues.
func NewIssuesCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	scope = scope.Child("issues")
	return &cobra.Command{
		Use:   "issues",
		Short: "List active issues for the current account",
		RunE: func(cmd *cobra.Command, _ []string) error {
			services, err := accountServices(cmd.Context(), cliConfig, scope)
			if err != nil {
				return err
			}
			issues, err := services.Issues.List(cmd.Context())
			if err != nil {
				return fmt.Errorf("list issues: %w", err)
			}
			s := styles.DetectTheme().Styles
			if len(issues) == 0 {
				fmt.Println(s.Help.Render("No active issues."))
				return nil
			}
			w := newTabWriter()
			fmt.Fprintln(w, "PRIORITY\tID\tSERVICE\tCOST/YR\tTITLE")
			for _, i := range issues {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					i.Priority, dashIfEmpty(i.DisplayID), dashIfEmpty(i.ServiceName), cost(i.CostPerHour), i.Title)
			}
			return w.Flush()
		},
	}
}

// NewServicesCmd lists enabled services and their status.
func NewServicesCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	scope = scope.Child("services")
	return &cobra.Command{
		Use:   "services",
		Short: "List enabled services for the current account",
		RunE: func(cmd *cobra.Command, _ []string) error {
			services, err := accountServices(cmd.Context(), cliConfig, scope)
			if err != nil {
				return err
			}
			statuses, err := services.Status.ListServiceStatuses(cmd.Context())
			if err != nil {
				return fmt.Errorf("list services: %w", err)
			}
			s := styles.DetectTheme().Styles
			if len(statuses) == 0 {
				fmt.Println(s.Help.Render("No enabled services."))
				return nil
			}
			w := newTabWriter()
			fmt.Fprintln(w, "SERVICE\tHEALTH\tLOG EVENTS\tVOLUME\tCOST/YR")
			for _, svc := range statuses {
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
					svc.Name, svc.Health, svc.LogEventCount, rate(svc.ServiceVolumePerHour), cost(svc.ServiceCostPerHourVolumeUSD))
			}
			return w.Flush()
		},
	}
}

// NewChecksCmd lists product checks and their posture.
func NewChecksCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	scope = scope.Child("checks")
	return &cobra.Command{
		Use:   "checks",
		Short: "List product checks and their posture for the current account",
		RunE: func(cmd *cobra.Command, _ []string) error {
			services, err := accountServices(cmd.Context(), cliConfig, scope)
			if err != nil {
				return err
			}
			catalog, err := services.Checks.List(cmd.Context())
			if err != nil {
				return fmt.Errorf("list checks: %w", err)
			}
			s := styles.DetectTheme().Styles
			if len(catalog.Checks) == 0 {
				fmt.Println(s.Help.Render("No checks."))
				return nil
			}
			w := newTabWriter()
			fmt.Fprintln(w, "CHECK\tDOMAIN\tOPEN FINDINGS\tACTIVE ISSUES\tCOST/YR")
			for _, c := range catalog.Checks {
				fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\n",
					c.Name, c.Domain, c.OpenFindingCount, c.ActiveIssueCount, cost(c.CurrentCostPerHour))
			}
			return w.Flush()
		},
	}
}

// NewStatusCmd prints the account-level status summary.
func NewStatusCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	scope = scope.Child("status")
	return &cobra.Command{
		Use:   "status",
		Short: "Show the current account's overall status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			services, err := accountServices(cmd.Context(), cliConfig, scope)
			if err != nil {
				return err
			}
			summary, err := services.Status.GetAccountSummary(cmd.Context())
			if err != nil {
				return fmt.Errorf("account status: %w", err)
			}
			issues, err := services.Issues.GetSummary(cmd.Context())
			if err != nil {
				return fmt.Errorf("issue summary: %w", err)
			}
			theme := styles.DetectTheme()
			s := theme.Styles
			fmt.Println(s.Title.Render("Account Status"))
			w := newTabWriter()
			fmt.Fprintf(w, "Health\t%s\n", summary.Health)
			fmt.Fprintf(w, "Ready for use\t%t\n", summary.ReadyForUse)
			fmt.Fprintf(w, "Services\t%d active / %d total\n", summary.ActiveServices, summary.ServiceCount)
			fmt.Fprintf(w, "Log events\t%d (%d analyzed)\n", summary.EventCount, summary.AnalyzedCount)
			fmt.Fprintf(w, "Volume\t%s\n", rate(summary.TotalVolumePerHour))
			fmt.Fprintf(w, "Cost\t%s\n", cost(summary.TotalCostPerHour))
			fmt.Fprintf(w, "Open issues\t%d (%d high, %d medium, %d low)\n",
				issues.Open, issues.HighCount, issues.MediumCount, issues.LowCount)
			return w.Flush()
		},
	}
}

// NewEdgeCmd lists the account's edge instances.
func NewEdgeCmd(scope log.Scope, cliConfig *config.CLIConfig) *cobra.Command {
	scope = scope.Child("edge")
	return &cobra.Command{
		Use:   "edge",
		Short: "List edge instances for the current account",
		RunE: func(cmd *cobra.Command, _ []string) error {
			services, err := accountServices(cmd.Context(), cliConfig, scope)
			if err != nil {
				return err
			}
			fleet, err := services.EdgeInstances.List(cmd.Context())
			if err != nil {
				return fmt.Errorf("list edge instances: %w", err)
			}
			s := styles.DetectTheme().Styles
			if len(fleet.Instances) == 0 {
				fmt.Println(s.Help.Render("No edge instances registered."))
				return nil
			}
			w := newTabWriter()
			fmt.Fprintln(w, "SERVICE\tNAMESPACE\tINSTANCE\tLAST SYNC")
			for _, inst := range fleet.Instances {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					inst.ServiceName, dashIfEmpty(inst.ServiceNamespace), inst.InstanceID, inst.LastSyncAt.Format("2006-01-02 15:04"))
			}
			return w.Flush()
		},
	}
}

func dashIfEmpty(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
