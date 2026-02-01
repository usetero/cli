package block

import "fmt"

// ShowSeriesInput displays a categorical series (bar chart, pie chart).
type ShowSeriesInput struct {
	Title  string       `json:"title"`
	Series []SeriesItem `json:"series"`
}

// SeriesItem is a single data point in a series.
type SeriesItem struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

// Validate checks that required fields are present.
func (i ShowSeriesInput) Validate() error {
	if i.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(i.Series) == 0 {
		return fmt.Errorf("series is required")
	}
	return nil
}
