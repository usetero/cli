package selectlist

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
)

type delegate struct {
	theme theme.Theme
}

func newList(appTheme theme.Theme) list.Model {
	l := list.New(nil, delegate{theme: appTheme}, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.FilterInput.Prompt = ""
	l.FilterInput.Placeholder = "Type to filter..."
	l.Styles.Filter.Focused.Prompt = appTheme.Text.Subtle
	l.Styles.Filter.Blurred.Prompt = appTheme.Text.Subtle
	l.Styles.Filter.Focused.Text = appTheme.Text.Body
	l.Styles.Filter.Blurred.Text = appTheme.Text.Body
	l.Styles.Filter.Focused.Placeholder = appTheme.Input.Placeholder
	l.Styles.Filter.Blurred.Placeholder = appTheme.Input.Placeholder
	l.Styles.Filter.Cursor.Color = appTheme.Palette.Brand
	l.Styles.Filter.Cursor.Shape = tea.CursorBar
	l.Styles.Filter.Cursor.Blink = true
	l.Styles.NoItems = appTheme.Text.Subtle
	return l
}

func (d delegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	row, ok := item.(Row)
	if !ok || m.Width() <= 0 {
		return
	}

	cursorStyle := lipgloss.NewStyle().
		Inherit(d.theme.Text.Subtle).
		Background(d.theme.Background)
	titleStyle := lipgloss.NewStyle().
		Inherit(d.theme.Text.Body).
		Background(d.theme.Background)
	subtitleStyle := lipgloss.NewStyle().
		Inherit(d.theme.Text.Subtle).
		Background(d.theme.Background)

	cursor := cursorStyle.Render("  ")
	title := titleStyle.Render(row.Title())
	if index == m.Index() {
		cursor = lipgloss.NewStyle().
			Inherit(d.theme.Input.Active).
			Background(d.theme.Background).
			Render("▶ ")
		title = lipgloss.NewStyle().
			Inherit(d.theme.Text.Body).
			Foreground(d.theme.Palette.Brand).
			Background(d.theme.Background).
			Bold(true).
			Render(row.Title())
	}

	line := lipgloss.JoinHorizontal(lipgloss.Left, cursor, title)
	if subtitle := strings.TrimSpace(row.Subtitle()); subtitle != "" {
		line = lipgloss.JoinHorizontal(lipgloss.Left, line, " ", subtitleStyle.Render(subtitle))
	}

	fmt.Fprint(w, ansi.Truncate(line, m.Width(), "…"))
}

func (d delegate) Height() int  { return 1 }
func (d delegate) Spacing() int { return 0 }
func (d delegate) Update(tea.Msg, *list.Model) tea.Cmd {
	return nil
}
