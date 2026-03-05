package syncer

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Endpoint is the validated base URL for the PowerSync API.
type Endpoint struct {
	raw string
}

// ParseEndpoint validates and canonicalizes a PowerSync endpoint.
func ParseEndpoint(raw string) (Endpoint, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Endpoint{}, fmt.Errorf("%w: endpoint is required", ErrInvalidInput)
	}

	u, err := url.Parse(raw)
	if err != nil {
		return Endpoint{}, fmt.Errorf("%w: invalid endpoint: %v", ErrInvalidInput, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return Endpoint{}, fmt.Errorf("%w: endpoint must use http or https", ErrInvalidInput)
	}
	if u.Host == "" {
		return Endpoint{}, fmt.Errorf("%w: endpoint host is required", ErrInvalidInput)
	}

	u.RawQuery = ""
	u.Fragment = ""
	return Endpoint{raw: strings.TrimRight(u.String(), "/")}, nil
}

func (e Endpoint) String() string { return e.raw }

// RetryPolicy controls reconnect behavior after sync failures.
type RetryPolicy struct {
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	ErrorStateAfter int
}

// DefaultRetryPolicy returns production defaults.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		InitialDelay:    1 * time.Second,
		MaxDelay:        10 * time.Second,
		ErrorStateAfter: 3,
	}
}

// Validate checks policy fields are sane.
func (p RetryPolicy) Validate() error {
	if p.InitialDelay <= 0 {
		return fmt.Errorf("%w: retry initial delay must be > 0", ErrInvalidInput)
	}
	if p.MaxDelay <= 0 {
		return fmt.Errorf("%w: retry max delay must be > 0", ErrInvalidInput)
	}
	if p.MaxDelay < p.InitialDelay {
		return fmt.Errorf("%w: retry max delay must be >= initial delay", ErrInvalidInput)
	}
	if p.ErrorStateAfter <= 0 {
		return fmt.Errorf("%w: retry error-state threshold must be > 0", ErrInvalidInput)
	}
	return nil
}
