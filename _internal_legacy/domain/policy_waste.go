package domain

// WastePolicy is a single pending waste policy with context from joined tables.
// Float pointers are nil when the underlying data hasn't been measured yet.
type WastePolicy struct {
	LogEventName               string
	ServiceName                string
	VolumePerHour              *float64 // Current throughput in events/hour; nil when unmeasured
	BytesPerHour               *float64 // Current throughput in bytes/hour; nil when unmeasured
	EstimatedCostPerHour       *float64 // Estimated cost reduction (total); nil when unmeasured
	EstimatedCostPerHourBytes  *float64 // Estimated cost reduction from bytes; nil when unmeasured
	EstimatedCostPerHourVolume *float64 // Estimated cost reduction from volume; nil when unmeasured
	EstimatedBytesPerHour      *float64 // Estimated bytes reduction; nil when unmeasured
	EstimatedVolumePerHour     *float64 // Estimated volume reduction; nil when unmeasured
}
