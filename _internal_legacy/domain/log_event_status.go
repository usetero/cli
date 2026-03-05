package domain

// LogEventStatus is a per-log-event status within a service.
// Used for the service detail drill-down in the statusbar drawer.
type LogEventStatus struct {
	Name                       string
	VolumePerHour              *float64
	BytesPerHour               *float64
	CostPerHourUSD             *float64
	PendingPolicyCount         int64
	ApprovedPolicyCount        int64
	PolicyPendingCriticalCount int64
	PolicyPendingHighCount     int64
	PolicyPendingMediumCount   int64
	PolicyPendingLowCount      int64
}
