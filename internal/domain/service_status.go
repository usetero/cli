package domain

// ServiceLogStatus is the catalog health status for a service or account.
type ServiceLogStatus string

const (
	ServiceLogStatusDisabled    ServiceLogStatus = "DISABLED"
	ServiceLogStatusInactive    ServiceLogStatus = "INACTIVE"
	ServiceLogStatusBroken      ServiceLogStatus = "BROKEN"
	ServiceLogStatusStale       ServiceLogStatus = "STALE"
	ServiceLogStatusDiscovering ServiceLogStatus = "DISCOVERING"
	ServiceLogStatusAnalyzing   ServiceLogStatus = "ANALYZING"
	ServiceLogStatusReady       ServiceLogStatus = "READY"
)

func (s ServiceLogStatus) String() string { return string(s) }

// ServiceStatus is a service's catalog health with throughput metrics.
type ServiceStatus struct {
	Name            string
	Status          ServiceLogStatus
	Error           string
	PercentComplete float64
	EventCount      int64
	AnalyzedCount   int64
	VolumePerHour   float64
	BytesPerHour    float64
	CostPerHourUSD  float64
}
