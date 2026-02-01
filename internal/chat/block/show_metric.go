package block

import "fmt"

// ShowMetricInput displays a single metric value.
type ShowMetricInput struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Unit  string `json:"unit,omitempty"`
}

// Validate checks that required fields are present.
func (i ShowMetricInput) Validate() error {
	if i.Label == "" {
		return fmt.Errorf("label is required")
	}
	if i.Value == "" {
		return fmt.Errorf("value is required")
	}
	return nil
}
