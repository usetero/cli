package messagelist

import "testing"

func TestProjectMouseToView(t *testing.T) {
	t.Parallel()

	viewX, viewY := projectMouseToView(25, 11, 10, 3)
	if viewX != 14 || viewY != 8 {
		t.Fatalf("got (%d,%d), want (14,8)", viewX, viewY)
	}
}

func TestResolveMouseTarget(t *testing.T) {
	t.Parallel()

	t.Run("hit", func(t *testing.T) {
		t.Parallel()
		got := resolveMouseTarget(20, 9, 5, 2, func(viewY int) (int, int) {
			if viewY == 7 {
				return 3, 4
			}
			return -1, -1
		})
		if !got.hit || got.blockIdx != 3 || got.blockY != 4 {
			t.Fatalf("unexpected target: %+v", got)
		}
		if got.viewX != 14 || got.viewY != 7 {
			t.Fatalf("unexpected view coords: %+v", got)
		}
	})

	t.Run("miss", func(t *testing.T) {
		t.Parallel()
		got := resolveMouseTarget(20, 9, 5, 2, func(int) (int, int) {
			return -1, -1
		})
		if got.hit {
			t.Fatalf("expected miss: %+v", got)
		}
		if got.blockIdx != -1 || got.blockY != -1 {
			t.Fatalf("unexpected block coords: %+v", got)
		}
	})
}
