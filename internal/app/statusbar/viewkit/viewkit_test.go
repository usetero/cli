package viewkit

import (
	"strings"
	"testing"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/styles"
)

func TestRenderPolicyEmptyState(t *testing.T) {
	theme := styles.NewTheme(true)

	t.Run("waiting when db not ready", func(t *testing.T) {
		got := RenderPolicyEmptyState(theme, false, domain.AccountSummary{}, "disabled hint", "healthy")
		if !strings.Contains(got, "Waiting for sync to start") {
			t.Fatalf("expected waiting message, got %q", got)
		}
	})

	t.Run("disabled services message", func(t *testing.T) {
		got := RenderPolicyEmptyState(theme, true, domain.AccountSummary{
			ServiceCount:   3,
			ActiveServices: 0,
		}, "enable services", "healthy")
		if !strings.Contains(got, "3 services discovered, all disabled") {
			t.Fatalf("expected disabled services message, got %q", got)
		}
		if !strings.Contains(got, "enable services") {
			t.Fatalf("expected disabled hint, got %q", got)
		}
	})
}

func TestComposeSummaryTableView(t *testing.T) {
	theme := styles.NewTheme(true)

	got := ComposeSummaryTableView(theme, "headline", "table", "description")
	if !strings.Contains(got, "headline") || !strings.Contains(got, "table") || !strings.Contains(got, "description") {
		t.Fatalf("expected composed view content, got %q", got)
	}
}

func TestRenderServicesEmptyState(t *testing.T) {
	theme := styles.NewTheme(true)

	got := RenderServicesEmptyState(theme, true, domain.AccountSummary{}, "hint")
	if !strings.Contains(got, "No services discovered yet") {
		t.Fatalf("expected no services message, got %q", got)
	}
}
