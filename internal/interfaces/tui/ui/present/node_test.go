package present

import (
	"fmt"
	"strings"
	"testing"

	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

func TestPanelKeepsBackgroundAcrossHeadingRow(t *testing.T) {
	appTheme := theme.New(false).OnSurface()
	bg := backgroundSeq(appTheme)
	content := StackGap(
		1,
		appTheme.Text.Section.Render("Finish Authentication In Your Browser"),
		appTheme.Text.Body.Render("https://example.com/device?user_code=ABCD-EFGH"),
	)

	out := Panel(appTheme, 100, content)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected multi-line panel, got %q", out)
	}

	headingRow := lines[1]
	if !strings.Contains(headingRow, "Finish Authentication In Your Browser") {
		t.Fatalf("expected heading row, got %q", headingRow)
	}

	textEnd := strings.Index(headingRow, "Browser") + len("Browser")
	if textEnd <= 0 || textEnd >= len(headingRow) {
		t.Fatalf("could not locate heading end in %q", headingRow)
	}

	trailing := headingRow[textEnd:]
	if !strings.Contains(trailing, bg) {
		t.Fatalf("expected heading row trailing area to restore panel background, got %q", headingRow)
	}
}

func TestPanelKeepsBackgroundAcrossShortActionRow(t *testing.T) {
	appTheme := theme.New(false).OnSurface()
	bg := backgroundSeq(appTheme)
	content := appTheme.Text.Body.Render("[enter] Get started")

	out := Panel(appTheme, 100, content)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected panel lines, got %q", out)
	}

	actionRow := lines[1]
	if !strings.Contains(actionRow, "Get started") {
		t.Fatalf("expected action row, got %q", actionRow)
	}

	textEnd := strings.Index(actionRow, "started") + len("started")
	if textEnd <= 0 || textEnd >= len(actionRow) {
		t.Fatalf("could not locate action end in %q", actionRow)
	}

	trailing := actionRow[textEnd:]
	if !strings.Contains(trailing, bg) {
		t.Fatalf("expected action row trailing area to restore panel background, got %q", actionRow)
	}
}

func backgroundSeq(appTheme theme.Theme) string {
	r, g, b, _ := appTheme.Background.RGBA()
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", r>>8, g>>8, b>>8)
}
