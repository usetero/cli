package datadogdiscovery

import (
	"testing"

	"github.com/usetero/cli/internal/domains/integrations"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

func TestModelStatusAndDetail(t *testing.T) {
	t.Parallel()

	t.Run("nil status", func(t *testing.T) {
		t.Parallel()

		m := New(theme.New(false))
		if got := m.statusLine(); got != "Connecting to Datadog and preparing discovery." {
			t.Fatalf("statusLine() = %q", got)
		}
		if detail := m.detailLine(); detail != "" {
			t.Fatalf("detailLine() = %q, want empty", detail)
		}
		if _, ok := m.percent(); ok {
			t.Fatalf("percent() should be unavailable without counts")
		}
	})

	t.Run("event counts drive progress", func(t *testing.T) {
		t.Parallel()

		m := New(theme.New(false))
		m.SetStatus(&integrations.DatadogStatus{
			EventCount:    100,
			AnalyzedCount: 42,
		})

		if got := m.statusLine(); got != "Analyzing Datadog events and building your account." {
			t.Fatalf("statusLine() = %q", got)
		}
		if got := m.detailLine(); got != "42 / 100 events analyzed" {
			t.Fatalf("detailLine() = %q", got)
		}
		pct, ok := m.percent()
		if !ok || pct != 42 {
			t.Fatalf("percent() = (%v, %t), want (42, true)", pct, ok)
		}
	})

	t.Run("active services are reported before event counts", func(t *testing.T) {
		t.Parallel()

		m := New(theme.New(false))
		m.SetStatus(&integrations.DatadogStatus{ActiveServices: 7})

		if got := m.detailLine(); got != "7 active services found" {
			t.Fatalf("detailLine() = %q", got)
		}
	})

	t.Run("ready status reads complete", func(t *testing.T) {
		t.Parallel()

		m := New(theme.New(false))
		m.SetStatus(&integrations.DatadogStatus{ReadyForUse: true})

		if got := m.statusLine(); got != "Discovery is complete." {
			t.Fatalf("statusLine() = %q", got)
		}
	})
}
