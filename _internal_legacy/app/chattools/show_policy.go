package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/domain/tools"
	"github.com/usetero/cli/internal/format"
	"github.com/usetero/cli/internal/sqlite"
)

// fetchPolicyCard loads a policy card via the typed wrapper.
func fetchPolicyCard(ctx context.Context, statuses sqlite.LogEventPolicyStatuses, id string) (tools.ShowResult, error) {
	card, err := statuses.GetPolicyCard(ctx, id)
	if err != nil {
		return tools.ShowResult{}, fmt.Errorf("fetch policy card: %w", err)
	}

	policy := domain.ParsePolicy(card)
	idShort := shortID(policy.ID.String())

	// Build the data map for the AI context. The AI sees these fields so it can
	// reference the policy conversationally without repeating the card.
	data := map[string]any{
		"policy":         policy, // typed policy for TUI rendering
		"policy_id":      policy.ID.String(),
		"service_name":   policy.ServiceName,
		"log_event_name": policy.LogEventName,
		"category":       policy.Category,
		"category_type":  string(policy.CategoryType),
		"action":         string(policy.Action),
		"status":         string(policy.Status),
	}

	if policy.CategoryDisplayName != "" {
		data["category_display_name"] = policy.CategoryDisplayName
	}
	if policy.Severity != "" {
		data["severity"] = string(policy.Severity)
	}
	if policy.Analysis != nil {
		data["rationale"] = policy.Analysis.Rationale()
		if subtitle := policy.Analysis.Subtitle(); subtitle != "" {
			data["subtitle"] = subtitle
		}
		if detail := policy.Analysis.ActionDetail(); detail != "" {
			data["action_detail"] = detail
		}
	}
	if policy.VolumePerHour != nil {
		data["volume_per_hour"] = *policy.VolumePerHour
	}
	if policy.BytesPerHour != nil {
		data["bytes_per_hour"] = *policy.BytesPerHour
	}
	if policy.EstimatedCostPerHour != nil {
		data["estimated_cost_per_hour"] = *policy.EstimatedCostPerHour
	}
	if policy.EstimatedVolumePerHour != nil {
		data["estimated_volume_per_hour"] = *policy.EstimatedVolumePerHour
	}
	if policy.EstimatedBytesPerHour != nil {
		data["estimated_bytes_per_hour"] = *policy.EstimatedBytesPerHour
	}
	if policy.SurvivalRate != nil {
		data["survival_rate"] = *policy.SurvivalRate
	}

	// Recommendation — same methods the card's viewRecommendation uses.
	if h := policy.Headline(); h != "" {
		data["headline"] = h
	}
	if m := policy.Mechanism(); m != "" {
		data["mechanism"] = m
	}

	// Impact — same method the card's viewImpact uses.
	if impact := policy.Impact(); impact != nil {
		impactMap := map[string]string{}
		if impact.VolumeFrom != "" {
			impactMap["volume_from"] = impact.VolumeFrom
			impactMap["volume_to"] = impact.VolumeTo
			impactMap["volume_reduction"] = impact.VolumePct
		}
		if impact.StorageFrom != "" {
			impactMap["storage_from"] = impact.StorageFrom
			impactMap["storage_to"] = impact.StorageTo
			impactMap["storage_reduction"] = impact.StoragePct
		}
		if impact.Savings != "" {
			impactMap["savings"] = impact.Savings
		}
		data["impact"] = impactMap
	}

	// Evidence — same method the card's viewEvidence uses.
	if ev := domain.BuildEvidence(policy); ev != nil {
		data["evidence"] = summarizeEvidence(ev)
	}

	// Build card summary for AI context.
	categoryName := policy.CategoryDisplayName
	if categoryName == "" {
		categoryName = format.TitleCase(string(policy.Category))
	}
	summary := buildCardSummary(idShort, categoryName, policy)

	return tools.ShowResult{
		Entity:      tools.EntityPolicy,
		ID:          policy.ID.String(),
		IDShort:     idShort,
		CardSummary: summary,
		Data:        data,
	}, nil
}

// shortID returns the first 4 hex characters of a UUID (stripping dashes).
func shortID(uuid string) string {
	hex := strings.ReplaceAll(uuid, "-", "")
	if len(hex) > 4 {
		return hex[:4]
	}
	return hex
}

// buildCardSummary creates a human-readable description of what the card shows,
// so the AI knows what's on the user's screen without repeating it.
func buildCardSummary(idShort, categoryName string, p *domain.Policy) string {
	var parts []string
	parts = append(parts, fmt.Sprintf(
		"Showing policy pol-%s: %s on %s/%s",
		idShort, categoryName, p.ServiceName, p.LogEventName,
	))

	sections := []string{"category", "rationale"}
	if domain.BuildEvidence(p) != nil {
		sections = append(sections, "sample log")
	}
	if p.Action != "" && p.Action != domain.PolicyActionNone {
		sections = append(sections, fmt.Sprintf("action (%s)", string(p.Action)))
	}
	if p.VolumePerHour != nil {
		sections = append(sections, fmt.Sprintf("volume (%s evt/hr)", format.Volume(*p.VolumePerHour)))
	}
	if cost := p.CostPerYear(); cost != "" {
		sections = append(sections, fmt.Sprintf("savings (%s)", cost))
	}
	parts = append(parts, "Card displays: "+strings.Join(sections, ", ")+".")
	parts = append(parts, "Status: "+string(p.Status)+".")

	if p.Analysis != nil {
		if subtitle := p.Analysis.Subtitle(); subtitle != "" {
			parts = append(parts, "Key detail: "+subtitle+".")
		}
	}

	return strings.Join(parts, " ")
}

// summarizeEvidence creates a flat map describing the evidence shown on the card.
// Uses the same domain types that the policycard's viewEvidence dispatches on.
func summarizeEvidence(ev domain.Evidence) map[string]any {
	switch ev := ev.(type) {
	case *domain.ConstantVariesEvidence:
		constantKeys := make([]string, len(ev.Constant))
		for i, f := range ev.Constant {
			constantKeys[i] = f.Key + "=" + f.Value
		}
		varyingKeys := make([]string, len(ev.Varying))
		for i, f := range ev.Varying {
			varyingKeys[i] = f.Key
		}
		return map[string]any{
			"type":            "constant_varies",
			"example_count":   ev.ExampleCount,
			"constant_count":  len(ev.Constant),
			"varying_count":   len(ev.Varying),
			"total_fields":    len(ev.Constant) + len(ev.Varying),
			"constant_fields": constantKeys,
			"varying_fields":  varyingKeys,
		}
	case *domain.HighlightedExampleEvidence:
		relevant := make([]string, len(ev.RelevantKeys))
		for i, k := range ev.RelevantKeys {
			relevant[i] = k.String()
		}
		return map[string]any{
			"type":            "highlighted_example",
			"total_fields":    len(ev.Attrs),
			"relevant_fields": relevant,
		}
	case *domain.FieldListEvidence:
		fields := make([]string, len(ev.Fields))
		for i, f := range ev.Fields {
			fields[i] = f.Key
		}
		return map[string]any{
			"type":           "field_list",
			"fields":         fields,
			"total_bytes":    ev.TotalBytes,
			"bytes_fraction": ev.BytesFraction,
		}
	default:
		return nil
	}
}
