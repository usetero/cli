package tools

// SetServiceEnabledInput is the input schema for the set_service_enabled tool.
type SetServiceEnabledInput struct {
	ServiceID string `json:"service_id"`
	Enabled   bool   `json:"enabled"`
}

// SetServiceEnabledResult is the typed output of a set_service_enabled tool execution.
type SetServiceEnabledResult struct {
	ServiceID   string `json:"service_id"`
	ServiceName string `json:"service_name"`
	Enabled     bool   `json:"enabled"`
}

// ToMap serializes the result for the GraphQL API.
func (r SetServiceEnabledResult) ToMap() map[string]any {
	return map[string]any{
		"service_id":   r.ServiceID,
		"service_name": r.ServiceName,
		"enabled":      r.Enabled,
	}
}
