package datadog

// apiKeyValidatedMsg is sent when API key validation completes.
type apiKeyValidatedMsg struct {
	valid    bool
	errorMsg string
	err      error
}

type validationError struct {
	msg string
}

func (e *validationError) Error() string { return e.msg }
