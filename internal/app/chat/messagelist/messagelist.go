package messagelist

import (
	"charm.land/bubbles/v2/key"
	"github.com/usetero/cli/internal/app/chat/messagelist/block"
	"github.com/usetero/cli/internal/app/chat/messagelist/round"
	chatclient "github.com/usetero/cli/internal/chat"
	"github.com/usetero/cli/internal/chat/tools"
	"github.com/usetero/cli/internal/log"
	"github.com/usetero/cli/internal/sqlite"
	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/viewport"
)

var (
	scrollUpKey   = key.NewBinding(key.WithKeys("up"))
	scrollDownKey = key.NewBinding(key.WithKeys("down"))
	focusPrevKey  = key.NewBinding(key.WithKeys("shift+up"))
	focusNextKey  = key.NewBinding(key.WithKeys("shift+down"))
)

// Layout constants.
const (
	// outerBorderWidth is the left border on the entire message list (thick accent bar).
	outerBorderWidth = 1

	// blockGap is the number of blank lines between blocks within a round.
	blockGap = 1

	// roundGap is the number of blank lines between rounds.
	roundGap = 2

	// gapBeforeDivider is blank lines before the round completion divider.
	gapBeforeDivider = 1

	// dividerHeight is the number of lines the divider itself occupies.
	dividerHeight = 1
)

// blockEntry pairs a block with its round metadata for layout decisions.
type blockEntry struct {
	block      block.Block
	roundIndex int
}

// Model displays the conversation history and manages rounds.
// Scroll, focus, and hit-test math is delegated to a viewport.Model.
//
// Methods are split across files by responsibility:
//   - messagelist.go  struct, constructor, accessors, layout helpers
//   - update.go       message dispatch
//   - render.go       View, block rendering, gaps, dividers
//   - mouse.go        click routing, text selection
//   - rounds.go       round lifecycle, block collection
type Model struct {
	theme styles.Theme
	scope log.Scope

	// Data hierarchy (owns the blocks, routes updates)
	rounds []*round.Model

	// Flat block list for viewport rendering (rebuilt from rounds)
	blocks []blockEntry
	layout layoutProjection

	// Viewport: scroll, focus, hit testing (pure math)
	vp viewport.Model

	// Viewport dimensions
	width  int
	height int

	// Screen origin (top-left corner of this component in terminal coordinates).
	// Set by the parent layout so mouse clicks can be translated.
	originX int
	originY int

	// Whether this component has keyboard focus (different from block focus)
	focused bool

	// Mouse selection state
	mouseDown      bool
	mouseDownBlock int // block index where mouse was pressed (-1 = none)
	mouseDownX     int // X within block content
	mouseDownY     int // Y within block (line offset)
	mouseDragBlock int // current block during drag (-1 = none)
	mouseDragX     int // current X within block content
	mouseDragY     int // current Y within block

	// Dependencies
	db           sqlite.DB
	chatClient   chatclient.Client
	toolRegistry *tools.Registry
}

// New creates a new message list.
func New(
	theme styles.Theme,
	db sqlite.DB,
	chatClient chatclient.Client,
	toolRegistry *tools.Registry,
	scope log.Scope,
) *Model {
	scope = scope.Child("messagelist")

	return &Model{
		theme:          theme,
		scope:          scope,
		vp:             viewport.New(),
		mouseDownBlock: -1,
		mouseDragBlock: -1,
		db:             db,
		chatClient:     chatClient,
		toolRegistry:   toolRegistry,
	}
}

// --- Size and focus ---

// SetSize sets the dimensions.
func (m *Model) SetSize(width, height int) {
	atBottom := m.vp.AtBottom()
	m.width = width
	m.height = height
	m.vp.SetHeight(height)
	m.updateRoundWidths()
	if atBottom {
		m.vp.ScrollToBottom()
	}
}

// SetOrigin sets the terminal-absolute position of this component's top-left corner.
// Called by the parent during layout so mouse coordinates can be translated.
func (m *Model) SetOrigin(x, y int) {
	m.originX = x
	m.originY = y
}

// SetFocused sets focus state.
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

// Focused returns whether the message list is focused.
func (m *Model) Focused() bool {
	return m.focused
}

// Len returns the number of rounds.
func (m *Model) Len() int {
	return len(m.rounds)
}

// --- Layout helpers ---

// contentWidth returns the width available for block content.
func (m *Model) contentWidth() int {
	return m.width - outerBorderWidth
}

// blockHeight returns the line count of block at idx without rendering.
func (m *Model) blockHeight(idx int) int {
	h := m.blocks[idx].block.Height()
	if h < 1 {
		return 1
	}
	return h
}
