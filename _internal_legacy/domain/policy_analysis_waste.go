package domain

import "fmt"

// Waste category slug constants matching control plane schema.
const (
	CategoryHealthChecks     PolicyCategory = "health_checks"
	CategoryBotTraffic       PolicyCategory = "bot_traffic"
	CategoryDebugArtifacts   PolicyCategory = "debug_artifacts"
	CategoryMalformed        PolicyCategory = "malformed"
	CategoryBrokenRecords    PolicyCategory = "broken_records"
	CategoryCommodityTraffic PolicyCategory = "commodity_traffic"
	CategoryRedundantEvents  PolicyCategory = "redundant_events"
	CategoryDeadWeight       PolicyCategory = "dead_weight"
)

// HealthChecksAnalysis is the analysis for health check / readiness probe events.
type HealthChecksAnalysis struct {
	baseAnalysis
}

func (a HealthChecksAnalysis) Category() PolicyCategory { return CategoryHealthChecks }

// BotTrafficAnalysis identifies log events with a user-agent field for bot filtering.
type BotTrafficAnalysis struct {
	baseAnalysis
	UserAgentField FieldPath `json:"user_agent_field"`         // Path to user-agent attribute
	BotProportion  *float64  `json:"bot_proportion,omitempty"` // Fraction of traffic identified as bot/crawler (0.0–1.0)
}

func (a BotTrafficAnalysis) Category() PolicyCategory { return CategoryBotTraffic }

func (a BotTrafficAnalysis) Subtitle() string {
	if a.BotProportion != nil {
		return fmt.Sprintf("~%.0f%% bot/crawler traffic", *a.BotProportion*100)
	}
	return ""
}

func (a BotTrafficAnalysis) ActionDetail() string {
	if a.BotProportion != nil {
		return fmt.Sprintf("~%.0f%% bot traffic", *a.BotProportion*100)
	}
	return ""
}

func (a BotTrafficAnalysis) RelevantKeys() []FieldPath {
	if !a.UserAgentField.IsEmpty() {
		return []FieldPath{a.UserAgentField}
	}
	return nil
}

// DebugArtifactsAnalysis is the analysis for developer debugging code that shipped to production.
type DebugArtifactsAnalysis struct {
	baseAnalysis
}

func (a DebugArtifactsAnalysis) Category() PolicyCategory { return CategoryDebugArtifacts }

// MalformedAnalysis is the analysis for unparseable or corrupted log data.
type MalformedAnalysis struct {
	baseAnalysis
}

func (a MalformedAnalysis) Category() PolicyCategory { return CategoryMalformed }

// BrokenRecordsAnalysis identifies near-identical events repeating endlessly.
type BrokenRecordsAnalysis struct {
	baseAnalysis
	MinIntervalSeconds int `json:"min_interval_seconds"` // Suggested minimum interval between kept events (1-30s)
}

func (a BrokenRecordsAnalysis) Category() PolicyCategory { return CategoryBrokenRecords }

func (a BrokenRecordsAnalysis) Subtitle() string {
	if a.MinIntervalSeconds > 0 {
		return fmt.Sprintf("Sample to 1 event per %s", FormatInterval(a.MinIntervalSeconds))
	}
	return ""
}

func (a BrokenRecordsAnalysis) ActionDetail() string      { return "" }
func (a BrokenRecordsAnalysis) RelevantKeys() []FieldPath { return nil }

// CommodityTrafficAnalysis identifies high-volume events where aggregate patterns
// matter more than individual records.
type CommodityTrafficAnalysis struct {
	baseAnalysis
	MinIntervalSeconds int `json:"min_interval_seconds"` // Suggested minimum interval between kept events (1-30s)
}

func (a CommodityTrafficAnalysis) Category() PolicyCategory { return CategoryCommodityTraffic }

func (a CommodityTrafficAnalysis) Subtitle() string {
	if a.MinIntervalSeconds > 0 {
		return fmt.Sprintf("Sample to 1 event per %s", FormatInterval(a.MinIntervalSeconds))
	}
	return ""
}

func (a CommodityTrafficAnalysis) ActionDetail() string      { return "" }
func (a CommodityTrafficAnalysis) RelevantKeys() []FieldPath { return nil }

// RedundantEventsAnalysis identifies events where another event in the same
// execution context already captures the same information.
type RedundantEventsAnalysis struct {
	baseAnalysis
}

func (a RedundantEventsAnalysis) Category() PolicyCategory { return CategoryRedundantEvents }

// DeadWeightAnalysis identifies events with no discernible value across any dimension.
type DeadWeightAnalysis struct {
	baseAnalysis
}

func (a DeadWeightAnalysis) Category() PolicyCategory { return CategoryDeadWeight }
