package domain

import "time"

// EdgeInstance is one edge runtime registered for the account.
type EdgeInstance struct {
	ID               string
	InstanceID       string
	ServiceName      string
	ServiceNamespace string
	LastSyncAt       time.Time
}

// EdgeFleet is the set of edge instances for the active account. Total is the
// server-reported fleet size.
type EdgeFleet struct {
	Total     int64
	Instances []EdgeInstance
}

// ConnectedCount returns the number of instances that synced within the given
// recency window relative to now.
func (f EdgeFleet) ConnectedCount(now time.Time, within time.Duration) int64 {
	var connected int64
	cutoff := now.Add(-within)
	for _, inst := range f.Instances {
		if inst.LastSyncAt.After(cutoff) {
			connected++
		}
	}
	return connected
}
