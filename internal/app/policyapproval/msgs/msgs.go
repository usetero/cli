package msgs

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
	PolicyIDs []string
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
