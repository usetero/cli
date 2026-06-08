package surfaces

import (
	"strings"
	"testing"

	"github.com/usetero/cli/internal/domain"
	"github.com/usetero/cli/internal/log/logtest"
	"github.com/usetero/cli/internal/styles"
)

func TestCompactViewHidesEmptyAndZeroPrimary(t *testing.T) {
	m := NewIssues(styles.NewTheme(true), logtest.NewScope(t))
	m.hasData = true

	cases := []struct {
		name    string
		primary Metric
		want    string
	}{
		{name: "no data loaded", primary: Metric{Value: "3"}, want: ""},
		{name: "zero value", primary: Metric{Value: "0"}, want: ""},
		{name: "empty value", primary: Metric{Value: ""}, want: ""},
		{name: "real value", primary: Metric{Value: "3"}, want: "issues"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m.hasData = tc.name != "no data loaded"
			m.snapshot = Snapshot{Title: "Issues", Primary: tc.primary}

			got := m.CompactView()
			if tc.want == "" {
				if got != "" {
					t.Fatalf("CompactView() = %q, want empty", got)
				}
				return
			}
			if !strings.Contains(got, tc.primary.Value) || !strings.Contains(got, tc.want) {
				t.Fatalf("CompactView() = %q, want it to contain %q and %q", got, tc.primary.Value, tc.want)
			}
		})
	}
}

func TestPendingCategorySignalBranches(t *testing.T) {
	cost := 4.0

	highPriority := []domain.PolicyCategoryStatus{{PolicyPendingHighCount: 2, EstimatedCostPerHour: &cost}}
	if got := pendingCategorySignal(highPriority); !strings.Contains(got, "high-priority") {
		t.Fatalf("expected high-priority signal to win, got %q", got)
	}

	costOnly := []domain.PolicyCategoryStatus{{EstimatedCostPerHour: &cost}}
	if got := pendingCategorySignal(costOnly); strings.Contains(got, "high-priority") || got == "" {
		t.Fatalf("expected a cost signal with no high-priority work, got %q", got)
	}

	idle := []domain.PolicyCategoryStatus{{}}
	if got := pendingCategorySignal(idle); got != "no high-priority policies" {
		t.Fatalf("expected idle signal, got %q", got)
	}
}

func TestAnalyzedToneBranches(t *testing.T) {
	if got := analyzedTone(domain.AccountSummary{}, 0); got != "" {
		t.Fatalf("expected no tone with zero total, got %q", got)
	}

	ready := domain.AccountSummary{EventCount: 100, AnalyzedCount: 100}
	if got := analyzedTone(ready, 100); got != "success" {
		t.Fatalf("expected success tone when analysis ready, got %q", got)
	}

	lagging := domain.AccountSummary{EventCount: 100, AnalyzedCount: 1}
	if got := analyzedTone(lagging, 100); got != "warning" {
		t.Fatalf("expected warning tone when analysis lagging, got %q", got)
	}
}

func TestSnapshotKeyDetectsContentChange(t *testing.T) {
	base := Snapshot{
		Title:   "Issues",
		Primary: Metric{Label: "open", Value: "3"},
		Metrics: []Metric{{Label: "high", Value: "1", Tone: "danger"}},
		Rows:    [][]string{{"Critical", "1", "needs review"}},
	}

	if snapshotKey(base) != snapshotKey(base) {
		t.Fatal("snapshotKey must be stable for identical snapshots")
	}

	changed := base
	changed.Primary.Value = "4"
	if snapshotKey(base) == snapshotKey(changed) {
		t.Fatal("snapshotKey must change when a metric value changes")
	}
}
