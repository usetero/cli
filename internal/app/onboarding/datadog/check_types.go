package datadog

// datadogCheckResultMsg is sent when datadog check completes.
type datadogCheckResultMsg struct {
	hasDatadog bool
	err        error
}
