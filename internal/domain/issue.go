package domain

// IssuePriority is how much attention a kept finding deserves.
type IssuePriority string

const (
	IssuePriorityLow    IssuePriority = "low"
	IssuePriorityMedium IssuePriority = "medium"
	IssuePriorityHigh   IssuePriority = "high"
)

func (p IssuePriority) String() string { return string(p) }

// IssueSummary is the server-computed aggregate state for active issues
// (issues whose closedAt and ignoredAt are both nil). The control plane
// computes the count and per-priority breakdown; the CLI never aggregates
// issue rows locally.
type IssueSummary struct {
	// Open is the total number of active issues.
	Open int64

	// Per-priority counts of active issues.
	HighCount   int64
	MediumCount int64
	LowCount    int64
}
