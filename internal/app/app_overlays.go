package app

import (
	"charm.land/lipgloss/v2"

	"github.com/usetero/cli/internal/tea/cursor"
)

// renderPaletteOverlay renders the palette centered on screen.
func (m *Model) renderPaletteOverlay(base string) string {
	paletteView := m.palette.View()
	paletteW := lipgloss.Width(paletteView)
	paletteH := lipgloss.Height(paletteView)
	centerX := (m.width - paletteW) / 2
	centerY := (m.height - paletteH) / 2

	// Extract cursor marker before compositing (compositor strips OSC sequences).
	cleanPalette, paletteCur := cursor.Extract(paletteView)

	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(base).X(0).Y(0),
		lipgloss.NewLayer(cleanPalette).X(centerX).Y(centerY),
	}
	result := lipgloss.NewCompositor(layers...).Render()

	// Re-insert cursor marker at the composited position.
	if paletteCur != nil {
		result = cursor.Insert(result, paletteCur.X+centerX, paletteCur.Y+centerY)
	}

	return result
}

func (m *Model) overlayDrawer(frame renderFrame) string {
	drawerWidth := frame.contentWidth - 2
	drawerHeight := frame.pageHeight - 2 // fill page area, leave gap at bottom
	if drawerHeight < 6 {
		drawerHeight = 6
	}
	drawer := m.statusBar.DrawerView(drawerWidth, drawerHeight)

	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(frame.paddedView).X(0).Y(0),
		lipgloss.NewLayer(drawer).X(horizontalPadding + 1).Y(frame.toastHeight + frame.statusBarHeight),
	}
	return lipgloss.NewCompositor(layers...).Render()
}

func (m *Model) overlayQuitDialog(base string) string {
	dialog := m.quitDlg.View()
	dialogW := lipgloss.Width(dialog)
	dialogH := lipgloss.Height(dialog)
	centerX := (m.width - dialogW) / 2
	centerY := (m.height - dialogH) / 2

	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(base).X(0).Y(0),
		lipgloss.NewLayer(dialog).X(centerX).Y(centerY),
	}
	return lipgloss.NewCompositor(layers...).Render()
}
