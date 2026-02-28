package datadog

// checkResultMsg is sent when datadog check completes.
type checkResultMsg struct {
	hasDatadog bool
	err        error
}
