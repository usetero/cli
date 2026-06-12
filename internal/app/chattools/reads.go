package tools

import (
	"encoding/json"
	"fmt"

	"github.com/usetero/cli/internal/app/chat/messagelist/round/turn/assistant/blocks/tools/action"
	"github.com/usetero/cli/internal/boundary/chat"
	graphql "github.com/usetero/cli/internal/boundary/graphql"
	"github.com/usetero/cli/internal/domain/tools"
)

// The read tools expose the control-plane catalog to the chat agent over
// GraphQL. Each returns structured rows in the tool result (fed back to the
// model) and a one-line summary for the conversation UI.

func deref(p *float64) any {
	if p == nil {
		return nil
	}
	return *p
}

// NewListServicesTool lists enabled services with their current status.
func NewListServicesTool(services graphql.ServiceSet) ActionTool {
	def := chat.Tool{
		Name:        "list_services",
		Description: "List enabled services with health, log-event counts, throughput, and cost. Use this to answer questions about the account's services.",
		InputSchema: chat.NewObjectSchema(map[string]chat.Property{}, nil),
	}

	executor := func(_ json.RawMessage) (tools.Result, error) {
		ctx, cancel := withToolTimeout()
		defer cancel()

		statuses, err := services.Status.ListServiceStatuses(ctx)
		if err != nil {
			return tools.Result{}, fmt.Errorf("list services: %w", err)
		}

		rows := make([]map[string]any, 0, len(statuses))
		for _, s := range statuses {
			rows = append(rows, map[string]any{
				"id":                  s.ID,
				"name":                s.Name,
				"health":              string(s.Health),
				"log_events":          s.LogEventCount,
				"events_per_hour":     deref(s.ServiceVolumePerHour),
				"cost_per_hour_usd":   deref(s.ServiceCostPerHourVolumeUSD),
				"analyzed_log_events": s.LogEventAnalyzedCount,
			})
		}
		return tools.Result{Content: map[string]any{"services": rows, "count": len(rows)}}, nil
	}

	return NewActionTool(def, executor, listConfig("Listing services", "services", "service"))
}

// NewListIssuesTool lists active issues with detail.
func NewListIssuesTool(services graphql.ServiceSet) ActionTool {
	def := chat.Tool{
		Name:        "list_issues",
		Description: "List active issues (highest priority first) with title, priority, owning service, and cost. Use this to answer questions about open issues.",
		InputSchema: chat.NewObjectSchema(map[string]chat.Property{}, nil),
	}

	executor := func(_ json.RawMessage) (tools.Result, error) {
		ctx, cancel := withToolTimeout()
		defer cancel()

		issues, err := services.Issues.List(ctx)
		if err != nil {
			return tools.Result{}, fmt.Errorf("list issues: %w", err)
		}

		rows := make([]map[string]any, 0, len(issues))
		for _, i := range issues {
			rows = append(rows, map[string]any{
				"id":                i.ID,
				"display_id":        i.DisplayID,
				"title":             i.Title,
				"priority":          string(i.Priority),
				"service":           i.ServiceName,
				"cost_per_hour_usd": deref(i.CostPerHour),
			})
		}
		return tools.Result{Content: map[string]any{"issues": rows, "count": len(rows)}}, nil
	}

	return NewActionTool(def, executor, listConfig("Listing issues", "issues", "issue"))
}

// NewListChecksTool lists product checks with posture.
func NewListChecksTool(services graphql.ServiceSet) ActionTool {
	def := chat.Tool{
		Name:        "list_checks",
		Description: "List product checks with their domain (cost/compliance) and account-scoped posture (open findings, active issues, affected services, current cost).",
		InputSchema: chat.NewObjectSchema(map[string]chat.Property{}, nil),
	}

	executor := func(_ json.RawMessage) (tools.Result, error) {
		ctx, cancel := withToolTimeout()
		defer cancel()

		catalog, err := services.Checks.List(ctx)
		if err != nil {
			return tools.Result{}, fmt.Errorf("list checks: %w", err)
		}

		rows := make([]map[string]any, 0, len(catalog.Checks))
		for _, c := range catalog.Checks {
			rows = append(rows, map[string]any{
				"id":                c.ID,
				"name":              c.Name,
				"domain":            string(c.Domain),
				"open_findings":     c.OpenFindingCount,
				"active_issues":     c.ActiveIssueCount,
				"affected_services": c.AffectedServiceCount,
				"cost_per_hour_usd": deref(c.CurrentCostPerHour),
			})
		}
		return tools.Result{Content: map[string]any{"checks": rows, "count": len(rows)}}, nil
	}

	return NewActionTool(def, executor, listConfig("Listing checks", "checks", "check"))
}

// NewListEdgeInstancesTool lists the account's edge instances.
func NewListEdgeInstancesTool(services graphql.ServiceSet) ActionTool {
	def := chat.Tool{
		Name:        "list_edge_instances",
		Description: "List edge instances syncing policies from this account, with the service they run and when they last synced.",
		InputSchema: chat.NewObjectSchema(map[string]chat.Property{}, nil),
	}

	executor := func(_ json.RawMessage) (tools.Result, error) {
		ctx, cancel := withToolTimeout()
		defer cancel()

		fleet, err := services.EdgeInstances.List(ctx)
		if err != nil {
			return tools.Result{}, fmt.Errorf("list edge instances: %w", err)
		}

		rows := make([]map[string]any, 0, len(fleet.Instances))
		for _, inst := range fleet.Instances {
			rows = append(rows, map[string]any{
				"id":           inst.ID,
				"instance_id":  inst.InstanceID,
				"service":      inst.ServiceName,
				"namespace":    inst.ServiceNamespace,
				"last_sync_at": inst.LastSyncAt,
			})
		}
		return tools.Result{Content: map[string]any{"edge_instances": rows, "count": len(rows)}}, nil
	}

	return NewActionTool(def, executor, listConfig("Listing edge instances", "edge instances", "edge instance"))
}

// NewAccountStatusTool returns the account-level status summary.
func NewAccountStatusTool(services graphql.ServiceSet) ActionTool {
	def := chat.Tool{
		Name:        "account_status",
		Description: "Get the account's overall status: readiness, service counts, log-event coverage, and total throughput/cost. Use this for high-level questions about the account.",
		InputSchema: chat.NewObjectSchema(map[string]chat.Property{}, nil),
	}

	executor := func(_ json.RawMessage) (tools.Result, error) {
		ctx, cancel := withToolTimeout()
		defer cancel()

		summary, err := services.Status.GetAccountSummary(ctx)
		if err != nil {
			return tools.Result{}, fmt.Errorf("account status: %w", err)
		}

		return tools.Result{Content: map[string]any{
			"ready_for_use":       summary.ReadyForUse,
			"health":              string(summary.Health),
			"services":            summary.ServiceCount,
			"active_services":     summary.ActiveServices,
			"ok_services":         summary.OkServices,
			"disabled_services":   summary.DisabledServices,
			"inactive_services":   summary.InactiveServices,
			"log_events":          summary.EventCount,
			"analyzed_log_events": summary.AnalyzedCount,
			"events_per_hour":     deref(summary.TotalVolumePerHour),
			"cost_per_hour_usd":   deref(summary.TotalCostPerHour),
		}}, nil
	}

	config := action.Config{
		DisplayName: func(json.RawMessage) string { return "Account status" },
		Status:      func(json.RawMessage) string { return "Checking account status" },
		Result:      func(tools.Result) string { return "Account status" },
	}
	return NewActionTool(def, executor, config)
}

// listConfig builds an action.Config for a list tool with a {count} result.
func listConfig(status, plural, singular string) action.Config {
	return action.Config{
		DisplayName: func(json.RawMessage) string { return status },
		Status:      func(json.RawMessage) string { return status },
		Result: func(result tools.Result) string {
			n, _ := result.Content["count"].(int)
			if n == 1 {
				return fmt.Sprintf("Found 1 %s", singular)
			}
			return fmt.Sprintf("Found %d %s", n, plural)
		},
	}
}
