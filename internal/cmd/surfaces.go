package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/config"
	"github.com/usetero/cli/internal/format"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/preferences"
	"github.com/usetero/cli/internal/styles"
)

// emit renders a command's result. When --output=json it writes data as
// indented JSON; otherwise it runs the table renderer. Every surface command
// routes through this so JSON support is uniform and free per command.
func emit(cmd *cobra.Command, data any, table func() error) error {
	if format, _ := cmd.Flags().GetString("output"); strings.EqualFold(format, "json") {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}
	return table()
}

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

type issueOut struct {
	ID        string `json:"id"`
	DisplayID string `json:"display_id"`
	Priority  string `json:"priority"`
	Service   string `json:"service,omitempty"`
	Title     string `json:"title"`
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
			out := make([]issueOut, len(issues))
			for i, is := range issues {
				out[i] = issueOut{ID: is.ID, DisplayID: is.DisplayID, Priority: string(is.Priority), Service: is.ServiceName, Title: is.Title}
			}
			return emit(cmd, out, func() error {
				if len(issues) == 0 {
					fmt.Println(styles.DetectTheme().Styles.Help.Render("No active issues."))
					return nil
				}
				w := newTabWriter()
				fmt.Fprintln(w, "PRIORITY\tID\tSERVICE\tTITLE")
				for _, i := range issues {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
						i.Priority, dashIfEmpty(i.DisplayID), dashIfEmpty(i.ServiceName), i.Title)
				}
				return w.Flush()
			})
		},
	}
}

type serviceOut struct {
	Name           string   `json:"name"`
	Health         string   `json:"health"`
	LogEvents      int64    `json:"log_events"`
	EventsPerHour  *float64 `json:"events_per_hour,omitempty"`
	CostPerHourUSD *float64 `json:"cost_per_hour_usd,omitempty"`
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
			out := make([]serviceOut, len(statuses))
			for i, svc := range statuses {
				out[i] = serviceOut{
					Name: svc.Name, Health: string(svc.Health), LogEvents: svc.LogEventCount,
					EventsPerHour: svc.ServiceVolumePerHour, CostPerHourUSD: svc.ServiceCostPerHourVolumeUSD,
				}
			}
			return emit(cmd, out, func() error {
				if len(statuses) == 0 {
					fmt.Println(styles.DetectTheme().Styles.Help.Render("No enabled services."))
					return nil
				}
				w := newTabWriter()
				fmt.Fprintln(w, "SERVICE\tHEALTH\tLOG EVENTS\tVOLUME\tCOST/YR")
				for _, svc := range statuses {
					fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n",
						svc.Name, svc.Health, svc.LogEventCount, rate(svc.ServiceVolumePerHour), cost(svc.ServiceCostPerHourVolumeUSD))
				}
				return w.Flush()
			})
		},
	}
}

type checkOut struct {
	Name           string   `json:"name"`
	Domain         string   `json:"domain"`
	OpenFindings   int64    `json:"open_findings"`
	ActiveIssues   int64    `json:"active_issues"`
	CostPerHourUSD *float64 `json:"cost_per_hour_usd,omitempty"`
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
			out := make([]checkOut, len(catalog.Checks))
			for i, c := range catalog.Checks {
				out[i] = checkOut{
					Name: c.Name, Domain: string(c.Domain), OpenFindings: c.OpenFindingCount,
					ActiveIssues: c.ActiveIssueCount, CostPerHourUSD: c.CurrentCostPerHour,
				}
			}
			return emit(cmd, out, func() error {
				if len(catalog.Checks) == 0 {
					fmt.Println(styles.DetectTheme().Styles.Help.Render("No checks."))
					return nil
				}
				w := newTabWriter()
				fmt.Fprintln(w, "CHECK\tDOMAIN\tOPEN FINDINGS\tACTIVE ISSUES\tCOST/YR")
				for _, c := range catalog.Checks {
					fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\n",
						c.Name, c.Domain, c.OpenFindingCount, c.ActiveIssueCount, cost(c.CurrentCostPerHour))
				}
				return w.Flush()
			})
		},
	}
}

type statusOut struct {
	Health         string   `json:"health"`
	ReadyForUse    bool     `json:"ready_for_use"`
	ActiveServices int64    `json:"active_services"`
	TotalServices  int64    `json:"total_services"`
	LogEvents      int64    `json:"log_events"`
	AnalyzedEvents int64    `json:"analyzed_events"`
	EventsPerHour  *float64 `json:"events_per_hour,omitempty"`
	CostPerHourUSD *float64 `json:"cost_per_hour_usd,omitempty"`
	OpenIssues     int64    `json:"open_issues"`
	HighIssues     int64    `json:"high_issues"`
	MediumIssues   int64    `json:"medium_issues"`
	LowIssues      int64    `json:"low_issues"`
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
			out := statusOut{
				Health: string(summary.Health), ReadyForUse: summary.ReadyForUse,
				ActiveServices: summary.ActiveServices, TotalServices: summary.ServiceCount,
				LogEvents: summary.EventCount, AnalyzedEvents: summary.AnalyzedCount,
				EventsPerHour: summary.TotalVolumePerHour, CostPerHourUSD: summary.TotalCostPerHour,
				OpenIssues: issues.Open, HighIssues: issues.HighCount, MediumIssues: issues.MediumCount, LowIssues: issues.LowCount,
			}
			return emit(cmd, out, func() error {
				s := styles.DetectTheme().Styles
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
			})
		},
	}
}

type edgeOut struct {
	Service    string `json:"service"`
	Namespace  string `json:"namespace,omitempty"`
	InstanceID string `json:"instance_id"`
	LastSyncAt string `json:"last_sync_at"`
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
			out := make([]edgeOut, len(fleet.Instances))
			for i, inst := range fleet.Instances {
				out[i] = edgeOut{
					Service: inst.ServiceName, Namespace: inst.ServiceNamespace,
					InstanceID: inst.InstanceID, LastSyncAt: inst.LastSyncAt.Format(time.RFC3339),
				}
			}
			return emit(cmd, out, func() error {
				if len(fleet.Instances) == 0 {
					fmt.Println(styles.DetectTheme().Styles.Help.Render("No edge instances registered."))
					return nil
				}
				w := newTabWriter()
				fmt.Fprintln(w, "SERVICE\tNAMESPACE\tINSTANCE\tLAST SYNC")
				for _, inst := range fleet.Instances {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
						inst.ServiceName, dashIfEmpty(inst.ServiceNamespace), inst.InstanceID, inst.LastSyncAt.Format("2006-01-02 15:04"))
				}
				return w.Flush()
			})
		},
	}
}

func dashIfEmpty(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
