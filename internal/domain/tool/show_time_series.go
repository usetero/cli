package tool

import "time"

type ShowTimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type ShowTimeSeriesLine struct {
	Label  string                `json:"label"`
	Points []ShowTimeSeriesPoint `json:"points"`
}

type ShowTimeSeriesInput struct {
	Title string               `json:"title"`
	Lines []ShowTimeSeriesLine `json:"lines"`
	Unit  string               `json:"unit,omitempty"`
}
