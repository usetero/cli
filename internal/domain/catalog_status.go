package domain

// CatalogSummary is the aggregated catalog health across all Datadog accounts.
type CatalogSummary struct {
	ReadyForUse      bool
	ServiceCount     int64
	ActiveServices   int64
	EventCount       int64
	AnalyzedCount    int64
	AnalyzingCount   int64
	DiscoveringCount int64
	BrokenServices   int64
	StaleServices    int64
	PercentComplete  float64
	WorstStatus      ServiceLogStatus
	LogError         string
}
