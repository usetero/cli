package datadog

// datadogCheckCompletedMsg is sent when datadog check completes.
type datadogCheckCompletedMsg struct {
	hasDatadog bool
	err        error
}
