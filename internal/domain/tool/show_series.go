package tool

type ShowSeriesItem struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

type ShowSeriesInput struct {
	Title string           `json:"title"`
	Items []ShowSeriesItem `json:"items"`
	Unit  string           `json:"unit,omitempty"`
}
