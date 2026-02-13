package msgs

import "github.com/usetero/cli/internal/domain"

type Policy struct {
	ID                     string
	ServiceName            string
	Category               string
	EstimatedVolumePerHour float64
	EstimatedCostPerHour   float64
}

type Start struct {
	ToolUseID string
	Policies  []Policy
}

type PoliciesSelected struct {
	Category string
	Policies []domain.PolicyDetail
}

type Confirmed struct{}

type CategorySelected struct {
	Category string
}

type ApproveAllLowRisk struct {
	Categories []string
}

type BackToSummary struct{}

type Cancelled struct {
	ToolUseID string
}

type PolicyApprovalComplete struct {
	ToolUseID     string
	ApprovedCount int
	FailedCount   int
	TotalSavings  float64
}
