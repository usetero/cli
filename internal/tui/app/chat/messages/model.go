package messages

import (
	"strings"

	"github.com/usetero/cli/internal/log"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/styles"
)

// Model is a scrollable, focusable list of message items.
// It handles keyboard navigation, mouse selection, and renders only visible items.
type Model struct {
	theme  *styles.Theme
	logger log.Logger
	items  []Item

	// Viewport dimensions
	width  int
	height int

	// Focus state - when focused, receives key events
	focused bool

	// Selection state (keyboard navigation)
	selectedIdx int // -1 means no selection

	// Scroll state
	offsetIdx  int // Index of first visible item
	offsetLine int // Lines scrolled within that item

	// Gap between items
	gap int

	// Mouse selection state
	mouseDown    bool // Is mouse button currently pressed?
	mouseDownIdx int  // Item index where drag started
	mouseDownX   int  // Column position where drag started
	mouseDownY   int  // Line within item where drag started
	mouseDragIdx int  // Current item index during drag
	mouseDragX   int  // Current column during drag
	mouseDragY   int  // Current line within item during drag
}

// New creates a new message list model.
func New(theme *styles.Theme, logger log.Logger) Model {
	return Model{
		theme:       theme,
		logger:      logger,
		selectedIdx: -1,
		gap:         1,
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseClickMsg:
		m, _ = m.handleMouseDown(msg.X, msg.Y)
		return m, nil

	case tea.MouseMotionMsg:
		m = m.handleMouseDrag(msg.X, msg.Y)
		return m, nil

	case tea.MouseReleaseMsg:
		return m.handleMouseUp(msg.X, msg.Y)

	case tea.MouseWheelMsg:
		switch msg.Button {
		case tea.MouseWheelUp:
			m = m.scrollUp(3)
		case tea.MouseWheelDown:
			m = m.scrollDown(3)
		}
		return m, nil

	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			m = m.scrollUp(1)
		case "down", "j":
			m = m.scrollDown(1)
		case "shift+up", "K":
			m = m.selectPrev()
		case "shift+down", "J":
			m = m.selectNext()
		case "g", "home":
			m = m.ScrollToTop()
		case "G", "end":
			m = m.ScrollToBottom()
		case "pgup", "b":
			m = m.scrollUp(m.height / 2)
		case "pgdown", "f":
			m = m.scrollDown(m.height / 2)
		case "enter", " ":
			// Toggle expand on selected item if expandable
			if m.selectedIdx >= 0 && m.selectedIdx < len(m.items) {
				if exp, ok := m.items[m.selectedIdx].(Expandable); ok {
					m.items[m.selectedIdx] = exp.ToggleExpanded()
				}
			}
		case "c", "y":
			// Copy selected item content or highlighted content
			if m.HasHighlight() {
				content := m.HighlightContent()
				m = m.ClearHighlight()
				return m, tea.SetClipboard(content)
			}
			if m.selectedIdx >= 0 && m.selectedIdx < len(m.items) {
				if cp, ok := m.items[m.selectedIdx].(Copyable); ok {
					return m, tea.SetClipboard(cp.CopyableContent())
				}
			}
		case "esc":
			// Clear highlight on escape
			if m.HasHighlight() {
				m = m.ClearHighlight()
				return m, nil
			}
		}

		// Forward to selected item
		if m.selectedIdx >= 0 && m.selectedIdx < len(m.items) {
			var cmd tea.Cmd
			m.items[m.selectedIdx], cmd = m.items[m.selectedIdx].Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// View renders the visible portion of the message list.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 || len(m.items) == 0 {
		return ""
	}

	var lines []string
	remainingHeight := m.height
	currentIdx := m.offsetIdx
	skipLines := m.offsetLine

	for remainingHeight > 0 && currentIdx < len(m.items) {
		item := m.items[currentIdx]
		rendered := item.Render(m.width)
		itemLines := strings.Split(rendered, "\n")

		// Skip lines for partial scroll
		if skipLines > 0 {
			if skipLines >= len(itemLines) {
				skipLines -= len(itemLines)
				currentIdx++
				continue
			}
			itemLines = itemLines[skipLines:]
			skipLines = 0
		}

		// Add visible lines from this item
		for _, line := range itemLines {
			if remainingHeight <= 0 {
				break
			}
			lines = append(lines, line)
			remainingHeight--
		}

		// Add gap after item (if not last and we have room)
		if currentIdx < len(m.items)-1 && m.gap > 0 {
			for i := 0; i < m.gap && remainingHeight > 0; i++ {
				lines = append(lines, "")
				remainingHeight--
			}
		}

		currentIdx++
	}

	// Pad to fill height
	for len(lines) < m.height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// SetSize sets the viewport dimensions.
func (m Model) SetSize(width, height int) Model {
	m.width = width
	m.height = height
	return m
}

// SetItems replaces all items in the list.
func (m Model) SetItems(items []Item) Model {
	m.items = items
	// Clamp selection
	if m.selectedIdx >= len(items) {
		m.selectedIdx = len(items) - 1
	}
	// Clamp scroll offset
	if m.offsetIdx >= len(items) {
		m.offsetIdx = max(0, len(items)-1)
		m.offsetLine = 0
	}
	return m
}

// AppendItem adds an item to the end of the list.
func (m Model) AppendItem(item Item) Model {
	m.items = append(m.items, item)
	return m
}

// UpdateItem updates an item by ID.
func (m Model) UpdateItem(id string, item Item) Model {
	for i, it := range m.items {
		if it.ID() == id {
			m.items[i] = item
			break
		}
	}
	return m
}

// Focus sets the focus state.
func (m Model) Focus() Model {
	m.focused = true
	// Select last item when focusing if nothing selected
	if m.selectedIdx < 0 && len(m.items) > 0 {
		m.selectedIdx = len(m.items) - 1
		m.items[m.selectedIdx] = m.items[m.selectedIdx].SetFocused(true)
	}
	return m
}

// Blur removes focus.
func (m Model) Blur() Model {
	m.focused = false
	// Clear selection styling
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.items) {
		m.items[m.selectedIdx] = m.items[m.selectedIdx].SetFocused(false)
	}
	return m
}

// IsFocused returns whether the list is focused.
func (m Model) IsFocused() bool {
	return m.focused
}

// ScrollToBottom scrolls to show the last item.
func (m Model) ScrollToBottom() Model {
	if len(m.items) == 0 {
		return m
	}

	// Calculate total height from bottom
	var totalHeight int
	lastIdx := len(m.items) - 1

	for idx := lastIdx; idx >= 0; idx-- {
		itemHeight := m.items[idx].Height(m.width)
		if idx < lastIdx {
			itemHeight += m.gap
		}
		totalHeight += itemHeight

		if totalHeight >= m.height {
			m.offsetIdx = idx
			m.offsetLine = totalHeight - m.height
			return m
		}
	}

	// All items fit
	m.offsetIdx = 0
	m.offsetLine = 0
	return m
}

// ScrollToTop scrolls to show the first item.
func (m Model) ScrollToTop() Model {
	m.offsetIdx = 0
	m.offsetLine = 0
	return m
}

// scrollUp scrolls up by n lines.
func (m Model) scrollUp(n int) Model {
	for n > 0 {
		if m.offsetLine > 0 {
			dec := min(n, m.offsetLine)
			m.offsetLine -= dec
			n -= dec
		} else if m.offsetIdx > 0 {
			m.offsetIdx--
			itemHeight := m.items[m.offsetIdx].Height(m.width)
			m.offsetLine = itemHeight - 1
			if m.gap > 0 {
				m.offsetLine += m.gap
			}
			n--
		} else {
			break
		}
	}
	return m
}

// scrollDown scrolls down by n lines.
func (m Model) scrollDown(n int) Model {
	for n > 0 && m.offsetIdx < len(m.items) {
		itemHeight := m.items[m.offsetIdx].Height(m.width)
		remainingInItem := itemHeight - m.offsetLine - 1
		if m.gap > 0 && m.offsetIdx < len(m.items)-1 {
			remainingInItem += m.gap
		}

		if n <= remainingInItem {
			m.offsetLine += n
			break
		}

		n -= remainingInItem + 1
		m.offsetIdx++
		m.offsetLine = 0
	}

	// Don't scroll past last item
	if m.offsetIdx >= len(m.items) {
		m = m.ScrollToBottom()
	}

	return m
}

// selectPrev selects the previous item.
func (m Model) selectPrev() Model {
	if len(m.items) == 0 {
		return m
	}

	// Clear current selection
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.items) {
		m.items[m.selectedIdx] = m.items[m.selectedIdx].SetFocused(false)
	}

	// Move selection
	if m.selectedIdx <= 0 {
		m.selectedIdx = 0
	} else {
		m.selectedIdx--
	}

	// Set new selection
	m.items[m.selectedIdx] = m.items[m.selectedIdx].SetFocused(true)

	// Ensure visible
	m = m.ensureSelectedVisible()
	return m
}

// selectNext selects the next item.
func (m Model) selectNext() Model {
	if len(m.items) == 0 {
		return m
	}

	// Clear current selection
	if m.selectedIdx >= 0 && m.selectedIdx < len(m.items) {
		m.items[m.selectedIdx] = m.items[m.selectedIdx].SetFocused(false)
	}

	// Move selection
	if m.selectedIdx < 0 {
		m.selectedIdx = 0
	} else if m.selectedIdx < len(m.items)-1 {
		m.selectedIdx++
	}

	// Set new selection
	m.items[m.selectedIdx] = m.items[m.selectedIdx].SetFocused(true)

	// Ensure visible
	m = m.ensureSelectedVisible()
	return m
}

// ensureSelectedVisible scrolls to make the selected item visible.
func (m Model) ensureSelectedVisible() Model {
	if m.selectedIdx < 0 || m.selectedIdx >= len(m.items) {
		return m
	}

	// Check if selected item is above viewport
	if m.selectedIdx < m.offsetIdx {
		m.offsetIdx = m.selectedIdx
		m.offsetLine = 0
		return m
	}

	// Calculate if selected item is below viewport
	var heightToSelected int
	for i := m.offsetIdx; i <= m.selectedIdx; i++ {
		if i == m.offsetIdx {
			heightToSelected = m.items[i].Height(m.width) - m.offsetLine
		} else {
			heightToSelected += m.items[i].Height(m.width)
		}
		if i < m.selectedIdx {
			heightToSelected += m.gap
		}
	}

	if heightToSelected > m.height {
		// Need to scroll down
		overflow := heightToSelected - m.height
		m = m.scrollDown(overflow)
	}

	return m
}

// AtBottom returns true if the list is scrolled to the bottom.
func (m Model) AtBottom() bool {
	if len(m.items) == 0 {
		return true
	}

	var totalHeight int
	for i := m.offsetIdx; i < len(m.items); i++ {
		if i == m.offsetIdx {
			totalHeight = m.items[i].Height(m.width) - m.offsetLine
		} else {
			totalHeight += m.items[i].Height(m.width)
		}
		if i < len(m.items)-1 {
			totalHeight += m.gap
		}
	}

	return totalHeight <= m.height
}

// Len returns the number of items.
func (m Model) Len() int {
	return len(m.items)
}

// SelectedItem returns the currently selected item, or nil if none.
func (m Model) SelectedItem() Item {
	if m.selectedIdx < 0 || m.selectedIdx >= len(m.items) {
		return nil
	}
	return m.items[m.selectedIdx]
}
