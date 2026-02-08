package domain

// LogEventStatus is the pipeline status for a log event.
type LogEventStatus string

const (
	LogEventStatusBroken      LogEventStatus = "BROKEN"
	LogEventStatusQuarantined LogEventStatus = "QUARANTINED"
	LogEventStatusResolved    LogEventStatus = "RESOLVED"
	LogEventStatusClean       LogEventStatus = "CLEAN"
	LogEventStatusPending     LogEventStatus = "PENDING"
	LogEventStatusAnalyzing   LogEventStatus = "ANALYZING"
	LogEventStatusDiscovering LogEventStatus = "DISCOVERING"
)

func (s LogEventStatus) String() string { return string(s) }
