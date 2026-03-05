package messagelist

import "testing"

func TestProjectLayout(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		p := projectLayout(nil, func(int) bool { return true })
		if len(p.heights) != 0 || len(p.gaps) != 0 || p.trailingDividerRound >= 0 {
			t.Fatalf("unexpected layout: %+v", p)
		}
	})

	t.Run("same round uses block gap", func(t *testing.T) {
		t.Parallel()
		items := []projectedItem{
			{roundIndex: 0, height: 3},
			{roundIndex: 0, height: 4},
		}
		p := projectLayout(items, func(int) bool { return true })
		if p.heights[0] != 3 || p.heights[1] != 4 {
			t.Fatalf("unexpected heights: %v", p.heights)
		}
		if p.gaps[1].height != blockGap {
			t.Fatalf("gap[1]=%d, want %d", p.gaps[1].height, blockGap)
		}
		if p.trailingDividerRound >= 0 {
			t.Fatalf("trailingDividerRound=%d, want none", p.trailingDividerRound)
		}
	})

	t.Run("round boundary with completed previous includes divider gap", func(t *testing.T) {
		t.Parallel()
		items := []projectedItem{
			{roundIndex: 0, height: 2},
			{roundIndex: 1, height: 2},
		}
		p := projectLayout(items, func(round int) bool {
			return round == 1 // round 0 completed, round 1 active
		})
		want := roundGap + gapBeforeDivider + dividerHeight
		if p.gaps[1].height != want {
			t.Fatalf("gap[1]=%d, want %d", p.gaps[1].height, want)
		}
		if p.gaps[1].dividerRound != 0 {
			t.Fatalf("dividerRound=%d, want 0", p.gaps[1].dividerRound)
		}
		if p.trailingDividerRound >= 0 {
			t.Fatalf("trailingDividerRound=%d, want none", p.trailingDividerRound)
		}
	})

	t.Run("trailing divider for completed last round", func(t *testing.T) {
		t.Parallel()
		items := []projectedItem{{roundIndex: 0, height: 2}}
		p := projectLayout(items, func(int) bool { return false })
		if p.trailingDividerRound != 0 {
			t.Fatalf("trailingDividerRound=%d, want 0", p.trailingDividerRound)
		}
	})
}
