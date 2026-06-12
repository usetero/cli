package domain

// CheckDomain is the lane a product check belongs to.
type CheckDomain string

const (
	CheckDomainCost       CheckDomain = "cost"
	CheckDomainCompliance CheckDomain = "compliance"
)

func (d CheckDomain) String() string { return string(d) }

// Check is one code-defined product check with its account-scoped posture.
// Posture counts and cost totals are computed server-side from findings and
// issues.
type Check struct {
	ID     string
	Name   string
	Domain CheckDomain

	OpenFindingCount      int64
	PendingFindingCount   int64
	EscalatedFindingCount int64
	ActiveIssueCount      int64
	AffectedServiceCount  int64

	// CurrentCostPerHour is the current spend attributable to this check, or
	// nil when unmeasured.
	CurrentCostPerHour *float64
}

// CheckCatalog is the full set of product checks plus server-computed
// per-domain counts.
type CheckCatalog struct {
	Total    int64
	Checks   []Check
	ByDomain map[CheckDomain]int64
}

// DomainCount returns the number of checks in a domain.
func (c CheckCatalog) DomainCount(d CheckDomain) int64 {
	return c.ByDomain[d]
}
