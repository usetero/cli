package datadog

import "testing"

func TestMain(m *testing.M) {
	// Disable browser opening during tests
	openBrowser = func(url string) error { return nil }
	m.Run()
}
