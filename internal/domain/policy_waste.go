package domain

// WastePolicy is a single pending waste policy with context from joined tables.
type WastePolicy struct {
	LogEventName          string
	ServiceName           string
	VolumePerHour         float64
	EstimatedCostPerHour  float64
	EstimatedBytesPerHour float64
}
