package messagelist

import (
	"fmt"
	"image"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/block"
	"github.com/usetero/cli/internal/app/chat/messagelist/round"
	"github.com/usetero/cli/internal/tea/highlight"
)

// View renders the visible portion of the message list.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	if len(m.blocks) == 0 {
		return m.emptyView()
	}

	lines := m.renderVisible()

	// Pad to viewport height
	for len(lines) < m.height {
		lines = append(lines, "")
	}

	output := strings.Join(lines, "\n")

	return lipgloss.NewStyle().
		PaddingLeft(outerBorderWidth).
		PaddingRight(outerPaddingRight).
		Render(output)
}

// renderVisible renders only the blocks visible in the viewport.
func (m *Model) renderVisible() []string {
	offsetIdx, offsetLine := m.vp.Offset()
	focusIdx := m.vp.FocusIdx()

	lines := make([]string, 0, m.height)
	reachedEnd := false

	// Highlight state
	startBlock, startLine, startCol, endBlock, endLine, endCol := m.getHighlightRange()
	hasHL := m.hasHighlight()
	highlighter := highlight.WithColors(m.theme.SelectionBg, m.theme.SelectionFg)

	for idx := offsetIdx; idx < len(m.blocks); idx++ {
		// Insert gap/divider before this block (except the first visible block)
		if idx > offsetIdx {
			lines = append(lines, m.gapLines(idx)...)
		}

		// Set focus state before rendering so the block can use it internally
		m.blocks[idx].block.SetFocused(m.focused && idx == focusIdx)

		// Render the block
		rendered := m.renderBlock(m.blocks[idx])

		// Apply text highlight if this block is in the selection range
		if hasHL {
			sLine, sCol, eLine, eCol := blockHighlightRange(idx, startBlock, startLine, startCol, endBlock, endLine, endCol)
			if sLine >= 0 {
				cw := m.contentWidth()
				h := lipgloss.Height(rendered)
				area := image.Rect(0, 0, cw, h)
				rendered = highlight.Apply(rendered, area, sLine, sCol, eLine, eCol, highlighter)
			}
		}

		blockLines := strings.Split(rendered, "\n")

		// For the first block, skip offsetLine lines (partial scroll into it)
		if idx == offsetIdx && offsetLine > 0 {
			if offsetLine < len(blockLines) {
				blockLines = blockLines[offsetLine:]
			} else {
				blockLines = nil
			}
		}

		lines = append(lines, blockLines...)

		if idx == len(m.blocks)-1 {
			reachedEnd = true
		}

		if len(lines) >= m.height {
			break
		}
	}

	// Add trailing divider for the last round if we rendered to the end
	if reachedEnd && len(m.blocks) > 0 {
		lastEntry := m.blocks[len(m.blocks)-1]
		lastRound := m.rounds[lastEntry.roundIndex]
		if lastRound.State() != round.StateActive {
			for range gapBeforeDivider {
				lines = append(lines, "")
			}
			lines = append(lines, m.divider(lastRound))
		}
	}

	// Trim to viewport height
	if len(lines) > m.height {
		lines = lines[:m.height]
	}

	return lines
}

// renderBlock renders a single block with appropriate width and padding.
func (m *Model) renderBlock(entry blockEntry) string {
	b := entry.block
	cw := m.contentWidth()

	// Determine border color based on block type and focus state.
	borderColor := m.theme.Bg
	if b.Kind() == block.KindUser {
		// User messages: accent border, brighter when focused.
		borderColor = m.theme.Accent
		if b.Focused() {
			borderColor = m.theme.AccentAlt
		}
	} else if b.Focused() {
		// Assistant blocks: invisible border, accent when focused.
		borderColor = m.theme.AccentAlt
	}

	return lipgloss.NewStyle().
		Width(cw).
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Render(b.View())
}

// gapLines returns the renderable lines to insert before block at idx.
// The length always equals gapSize(idx) — gapSize is the source of truth
// for measurement, this method produces the actual content.
func (m *Model) gapLines(idx int) []string {
	n := m.gapSize(idx)
	if n == 0 {
		return nil
	}

	prev := m.blocks[idx-1]
	curr := m.blocks[idx]

	// Same round — just blank lines
	if prev.roundIndex == curr.roundIndex {
		return make([]string, n)
	}

	// Different round — blank lines with an optional divider embedded
	lines := make([]string, 0, n)
	prevRound := m.rounds[prev.roundIndex]
	if prevRound.State() != round.StateActive {
		for range gapBeforeDivider {
			lines = append(lines, "")
		}
		lines = append(lines, m.divider(prevRound))
	}
	for range roundGap {
		lines = append(lines, "")
	}
	return lines
}

// divider renders "  ◇ Tero 4s ─────────" for a completed round,
// or "  ◇ Cancelled 1.2s ─────────" for a cancelled round.
func (m *Model) divider(r *round.Model) string {
	const indent = block.BorderWidth + block.PaddingX
	cw := m.contentWidth()

	colors := m.theme
	border := lipgloss.NewStyle().Foreground(colors.Border).Background(colors.Bg)

	duration := r.Duration()
	var durationStr string
	if duration < time.Minute {
		durationStr = fmt.Sprintf("%.1fs", duration.Seconds())
	} else {
		durationStr = fmt.Sprintf("%.1fm", duration.Minutes())
	}

	var prefix string
	var prefixStyle lipgloss.Style
	if r.State() == round.StateCancelled {
		prefix = fmt.Sprintf("◇ Cancelled %s ", durationStr)
		prefixStyle = lipgloss.NewStyle().Foreground(colors.ErrorFg).Background(colors.Bg)
	} else {
		prefix = fmt.Sprintf("◇ Tero %s ", durationStr)
		prefixStyle = lipgloss.NewStyle().Foreground(colors.TextMuted).Background(colors.Bg)
	}

	prefixWidth := lipgloss.Width(prefix)

	lineWidth := cw - indent - prefixWidth
	if lineWidth < 0 {
		lineWidth = 0
	}
	line := strings.Repeat("─", lineWidth)

	return strings.Repeat(" ", indent) + prefixStyle.Render(prefix) + border.Render(line)
}

// emptyView renders an empty view padded to height.
func (m *Model) emptyView() string {
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render("")
}
