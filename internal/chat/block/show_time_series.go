package block

import "fmt"

// ShowTimeSeriesInput displays a time series chart.
type ShowTimeSeriesInput struct {
	Title      string            `json:"title"`
	DataPoints []TimeSeriesPoint `json:"data_points"`
}

// TimeSeriesPoint is a single data point in a time series.
type TimeSeriesPoint struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

// Validate checks that required fields are present.
func (i ShowTimeSeriesInput) Validate() error {
	if i.Title == "" {
		return fmt.Errorf("title is required")
	}
	return nil
}
